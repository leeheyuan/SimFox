package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

type llmSignalSuggestionResponse struct {
	Summary  string                     `json:"summary"`
	Analysis SignalOptimizationAnalysis `json:"analysis"`
	Proposal SignalOptimizationProposal `json:"proposal"`
}

func generateSignalSuggestionWithLLM(
	metrics SignalOptimizationMetrics,
	heuristicAnalysis SignalOptimizationAnalysis,
	heuristicProposal SignalOptimizationProposal,
) (SignalOptimizationAnalysis, SignalOptimizationProposal, string, string, error) {
	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey == "" {
		return heuristicAnalysis, heuristicProposal, "", "", fmt.Errorf("OPENAI_API_KEY is not configured")
	}

	model := strings.TrimSpace(os.Getenv("SIMFOX_SIGNAL_OPT_MODEL"))
	if model == "" {
		model = "gpt-4o-mini"
	}

	promptPayload := map[string]any{
		"metrics":            metrics,
		"heuristicAnalysis":  heuristicAnalysis,
		"heuristicProposal":  heuristicProposal,
		"outputRequirements": "Return strict JSON with summary, analysis, proposal. Keep phase indices aligned with existing tlLogic phases. Do not invent junction IDs or source files outside candidateSignals.",
	}
	payloadBytes, err := json.Marshal(promptPayload)
	if err != nil {
		return heuristicAnalysis, heuristicProposal, "", "", err
	}

	config := openai.DefaultConfig(apiKey)
	if baseURL := strings.TrimSpace(os.Getenv("OPENAI_BASE_URL")); baseURL != "" {
		config.BaseURL = baseURL
	}
	client := openai.NewClientWithConfig(config)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleSystem,
				Content: strings.Join([]string{
					"You are a traffic signal optimization assistant for a SUMO-based simulation platform.",
					"Analyze congestion and propose cautious traffic-light timing adjustments.",
					"Return only valid JSON with fields: summary, analysis, proposal.",
					"analysis must contain: mode, congestedJunctions, reasons, recommendationLevel.",
					"proposal must contain: adjustments, nextStep.",
					"Each adjustment must contain: junctionId, programId, sourceFile, beforeCycle, afterCycle, phaseChanges.",
					"Each phaseChanges item must contain: index, state, oldDuration, newDuration, comment.",
					"Prefer conservative changes and keep all values grounded in the provided candidate signals.",
				}, " "),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: string(payloadBytes),
			},
		},
		Temperature: 0.2,
	})
	if err != nil {
		return heuristicAnalysis, heuristicProposal, "", "", err
	}
	if len(resp.Choices) == 0 {
		return heuristicAnalysis, heuristicProposal, "", "", fmt.Errorf("model returned no choices")
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	content = trimCodeFence(content)

	var llmResponse llmSignalSuggestionResponse
	if err := json.Unmarshal([]byte(content), &llmResponse); err != nil {
		return heuristicAnalysis, heuristicProposal, "", "", err
	}

	analysis := llmResponse.Analysis
	if analysis.Mode == "" {
		analysis.Mode = "openai"
	}
	proposal := normalizeLLMProposal(llmResponse.Proposal, metrics, heuristicProposal)
	summary := strings.TrimSpace(llmResponse.Summary)
	if summary == "" {
		summary = buildSignalSummary(metrics, analysis, proposal)
	}

	return analysis, proposal, summary, "openai:" + model, nil
}

func normalizeLLMProposal(
	proposal SignalOptimizationProposal,
	metrics SignalOptimizationMetrics,
	fallback SignalOptimizationProposal,
) SignalOptimizationProposal {
	candidateByID := make(map[string]SignalCandidateBrief, len(metrics.CandidateSignals))
	for _, candidate := range metrics.CandidateSignals {
		candidateByID[candidate.ID] = candidate
	}

	fallbackByID := make(map[string]SignalAdjustment, len(fallback.Adjustments))
	for _, adjustment := range fallback.Adjustments {
		fallbackByID[adjustment.JunctionID] = adjustment
	}

	normalized := SignalOptimizationProposal{
		Adjustments: make([]SignalAdjustment, 0, len(proposal.Adjustments)),
		NextStep:    strings.TrimSpace(proposal.NextStep),
	}

	for _, adjustment := range proposal.Adjustments {
		candidate, ok := candidateByID[adjustment.JunctionID]
		if !ok {
			continue
		}
		if strings.TrimSpace(adjustment.SourceFile) == "" {
			adjustment.SourceFile = candidate.SourceFile
		}
		if strings.TrimSpace(adjustment.ProgramID) == "" {
			adjustment.ProgramID = candidate.ProgramID
		}
		if adjustment.BeforeCycle <= 0 {
			adjustment.BeforeCycle = candidate.CycleLength
		}

		fallbackAdjustment, hasFallback := fallbackByID[adjustment.JunctionID]
		normalizedChanges := make([]SignalPhaseChange, 0, len(adjustment.PhaseChanges))
		for _, change := range adjustment.PhaseChanges {
			if change.Index < 0 {
				continue
			}
			if change.NewDuration <= 0 {
				continue
			}

			if hasFallback {
				for _, fallbackChange := range fallbackAdjustment.PhaseChanges {
					if fallbackChange.Index == change.Index {
						if strings.TrimSpace(change.State) == "" {
							change.State = fallbackChange.State
						}
						if change.OldDuration <= 0 {
							change.OldDuration = fallbackChange.OldDuration
						}
						break
					}
				}
			}

			if strings.TrimSpace(change.Comment) == "" {
				change.Comment = "LLM suggested timing adjustment."
			}
			normalizedChanges = append(normalizedChanges, change)
		}

		if len(normalizedChanges) == 0 {
			continue
		}
		adjustment.PhaseChanges = normalizedChanges
		if adjustment.AfterCycle <= 0 {
			adjustment.AfterCycle = estimateCycleLength(adjustment.BeforeCycle, normalizedChanges)
		}
		normalized.Adjustments = append(normalized.Adjustments, adjustment)
	}

	if len(normalized.Adjustments) == 0 {
		return fallback
	}
	if normalized.NextStep == "" {
		normalized.NextStep = fallback.NextStep
	}

	return normalized
}

func estimateCycleLength(before float64, changes []SignalPhaseChange) float64 {
	total := before
	for _, change := range changes {
		total += change.NewDuration - change.OldDuration
	}
	return round2(total)
}

func trimCodeFence(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "```json")
	value = strings.TrimPrefix(value, "```")
	value = strings.TrimSuffix(value, "```")
	return strings.TrimSpace(value)
}
