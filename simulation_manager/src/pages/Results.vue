<template>
  <section class="page">
    <div class="page-header">
      <div>
        <h2>Results</h2>
        <p>Browse completed simulation runs and open their generated logs.</p>
      </div>
      <el-button type="primary" :loading="loading" @click="refreshResults">Refresh</el-button>
    </div>

    <el-table v-loading="loading" :data="rows" style="width: 100%">
      <el-table-column prop="taskId" label="Task ID" width="120" />
      <el-table-column prop="project" label="Project" />
      <el-table-column prop="status" label="Status" width="130" />
      <el-table-column prop="summary" label="Summary" min-width="260" />
      <el-table-column prop="generatedAt" label="Generated At" width="180" />
      <el-table-column label="Actions" width="220">
        <template #default="scope">
          <el-button size="small" @click="openDetails(scope.row.taskId)">View</el-button>
          <el-button size="small" @click="openSignalOptimization(scope.row.taskId)">AI Suggestion</el-button>
          <el-button
            size="small"
            type="primary"
            :disabled="!canOpenResult(scope.row.taskId)"
            @click="openResult(scope.row.taskId)"
          >
            Download
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showDetailsDialog" title="Result Details" width="600px">
      <template v-if="selectedResult">
        <div class="meta-grid">
          <div><strong>Task ID:</strong> {{ selectedResult.taskId }}</div>
          <div><strong>Project:</strong> {{ selectedResult.projectName || `Project ${selectedResult.projectId}` }}</div>
          <div><strong>Status:</strong> {{ selectedResult.status }}</div>
          <div><strong>Generated:</strong> {{ formatDate(selectedResult.generatedAt) }}</div>
        </div>
        <el-form label-position="top">
          <el-form-item label="Summary">
            <el-input :model-value="describeResult(selectedResult)" type="textarea" :rows="3" readonly />
          </el-form-item>
          <el-form-item label="Log Path">
            <el-input :model-value="selectedResult.logUrl || 'No log path recorded'" readonly />
          </el-form-item>
          <el-form-item label="Last Error">
            <el-input :model-value="selectedResult.lastError || 'No error recorded'" type="textarea" :rows="4" readonly />
          </el-form-item>
        </el-form>
      </template>
      <template #footer>
        <el-button @click="showDetailsDialog = false">Close</el-button>
        <el-button
          v-if="selectedResult && canOpenInSumoGui(selectedResult)"
          @click="openInSumoGui(selectedResult)"
        >
          Open in SUMO GUI
        </el-button>
        <el-button
          v-if="selectedResult && canOpenResult(selectedResult.taskId)"
          type="primary"
          @click="openResult(selectedResult.taskId)"
        >
          Open Log
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showSignalDialog" title="AI Signal Optimization" width="760px">
      <template v-if="signalSuggestion">
        <div class="signal-header">
          <div>
            <div class="signal-title">{{ signalSuggestion.summary }}</div>
            <p class="signal-subtitle">
              Engine: {{ signalSuggestion.engine }} · Status: {{ signalSuggestion.status }}
            </p>
          </div>
          <el-button v-if="selectedResult" @click="openRoadEditor">Open Road Editor</el-button>
        </div>

        <div class="signal-kpis">
          <el-card shadow="never">
            <div class="signal-kpi-label">Trips</div>
            <strong>{{ signalSuggestion.metrics.tripCount || 0 }}</strong>
          </el-card>
          <el-card shadow="never">
            <div class="signal-kpi-label">Avg Wait</div>
            <strong>{{ signalSuggestion.metrics.averageWaitingTime || 0 }} s</strong>
          </el-card>
          <el-card shadow="never">
            <div class="signal-kpi-label">Avg Time Loss</div>
            <strong>{{ signalSuggestion.metrics.averageTimeLoss || 0 }} s</strong>
          </el-card>
        </div>

        <el-form label-position="top">
          <el-form-item label="Congested Junctions">
            <el-input
              :model-value="signalSuggestion.analysis.congestedJunctions.join(', ') || 'No junction inferred'"
              readonly
            />
          </el-form-item>
          <el-form-item label="Analysis Notes">
            <el-input
              :model-value="signalSuggestion.analysis.reasons.join('\n') || 'No analysis notes'"
              type="textarea"
              :rows="4"
              readonly
            />
          </el-form-item>
        </el-form>

        <el-table :data="signalSuggestion.proposal.adjustments" border empty-text="No writable signal plan found">
          <el-table-column prop="junctionId" label="Junction" width="120" />
          <el-table-column prop="programId" label="Program" width="120" />
          <el-table-column label="Cycle" width="140">
            <template #default="scope">
              {{ scope.row.beforeCycle }}s -> {{ scope.row.afterCycle }}s
            </template>
          </el-table-column>
          <el-table-column label="Phase Changes" min-width="320">
            <template #default="scope">
              <div class="phase-change-list">
                <div v-for="change in scope.row.phaseChanges" :key="`${scope.row.junctionId}-${change.index}`">
                  #{{ change.index }} {{ change.oldDuration }}s -> {{ change.newDuration }}s ({{ change.comment }})
                </div>
              </div>
            </template>
          </el-table-column>
        </el-table>

        <p class="signal-next-step">{{ signalSuggestion.proposal.nextStep }}</p>
        <p v-if="signalSuggestion.rejectedReason" class="signal-reason">
          Rejected reason: {{ signalSuggestion.rejectedReason }}
        </p>
        <p v-if="signalSuggestion.appliedSignalFile" class="signal-reason">
          Applied file: {{ signalSuggestion.appliedSignalFile }}
        </p>
        <div v-if="signalSuggestion.comparison" class="comparison-card">
          <div class="comparison-header">
            <strong>Optimization Comparison</strong>
            <span>
              Baseline #{{ signalSuggestion.comparison.baselineTaskId }}
              vs Optimized #{{ signalSuggestion.comparison.optimizedTaskId }}
              ({{ signalSuggestion.comparison.optimizedTaskStatus }})
            </span>
          </div>
          <div v-if="signalSuggestion.comparison.optimizedMetrics && signalSuggestion.comparison.delta" class="comparison-grid">
            <el-card shadow="never">
              <div class="signal-kpi-label">Avg Wait</div>
              <strong>
                {{ signalSuggestion.comparison.baselineMetrics.averageWaitingTime }}s ->
                {{ signalSuggestion.comparison.optimizedMetrics.averageWaitingTime }}s
              </strong>
              <div :class="deltaClass(signalSuggestion.comparison.delta.averageWaitingTimeDelta)">
                {{ formatDelta(signalSuggestion.comparison.delta.averageWaitingTimeDelta) }}s
                ({{ formatPercent(signalSuggestion.comparison.delta.averageWaitingTimeDeltaPercent) }})
              </div>
            </el-card>
            <el-card shadow="never">
              <div class="signal-kpi-label">Avg Time Loss</div>
              <strong>
                {{ signalSuggestion.comparison.baselineMetrics.averageTimeLoss }}s ->
                {{ signalSuggestion.comparison.optimizedMetrics.averageTimeLoss }}s
              </strong>
              <div :class="deltaClass(signalSuggestion.comparison.delta.averageTimeLossDelta)">
                {{ formatDelta(signalSuggestion.comparison.delta.averageTimeLossDelta) }}s
                ({{ formatPercent(signalSuggestion.comparison.delta.averageTimeLossDeltaPercent) }})
              </div>
            </el-card>
            <el-card shadow="never">
              <div class="signal-kpi-label">Avg Duration</div>
              <strong>
                {{ signalSuggestion.comparison.baselineMetrics.averageDuration }}s ->
                {{ signalSuggestion.comparison.optimizedMetrics.averageDuration }}s
              </strong>
              <div :class="deltaClass(signalSuggestion.comparison.delta.averageDurationDelta)">
                {{ formatDelta(signalSuggestion.comparison.delta.averageDurationDelta) }}s
                ({{ formatPercent(signalSuggestion.comparison.delta.averageDurationDeltaPercent) }})
              </div>
            </el-card>
          </div>
          <p v-else class="signal-reason">
            Optimized rerun task has been queued. This panel will update automatically after the rerun completes.
          </p>
        </div>
      </template>
      <template #footer>
        <el-button @click="showSignalDialog = false">Close</el-button>
        <el-button
          :disabled="!signalSuggestion || signalSuggestion.status === 'accepted'"
          @click="handleRejectSignal"
        >
          Reject
        </el-button>
        <el-button
          v-if="signalSuggestion"
          type="primary"
          :disabled="signalSuggestion.status === 'accepted' || !signalSuggestion.proposal.adjustments.length"
          @click="handleAcceptSignal"
        >
          Accept + Re-run
        </el-button>
      </template>
    </el-dialog>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  acceptSignalOptimization,
  getSignalOptimization,
  listResults,
  rejectSignalOptimization,
  type ResultListItem,
  type SignalOptimizationSuggestion,
} from '@/api/simulation'

const loading = ref(false)
const results = ref<ResultListItem[]>([])
const showDetailsDialog = ref(false)
const selectedResult = ref<ResultListItem | null>(null)
const showSignalDialog = ref(false)
const signalSuggestion = ref<SignalOptimizationSuggestion | null>(null)
const canUseSumoGui = computed(() => Boolean(window.simfox?.openInSumoGui))
const selectedSignalTaskId = ref<number | null>(null)
let signalRefreshTimer: number | null = null
const roadEditorBaseUrl = import.meta.env.VITE_ROAD_EDITOR_URL || 'http://localhost:5174/editor'
const roadEditorUrl = computed(() => {
  if (!selectedResult.value) {
    return ''
  }

  try {
    const url = new URL(roadEditorBaseUrl)
    const token = localStorage.getItem('token')
    if (token) {
      url.searchParams.set('token', token)
    }
    url.searchParams.set('projectId', String(selectedResult.value.projectId))
    return url.toString()
  } catch {
    return roadEditorBaseUrl
  }
})

const rows = computed(() =>
  results.value.map((result) => ({
    taskId: result.taskId,
    project: result.projectName || `Project ${result.projectId}`,
    status: result.status,
    summary: describeResult(result),
    generatedAt: formatDate(result.generatedAt),
  })),
)

onMounted(async () => {
  await refreshResults()
})

onBeforeUnmount(() => {
  stopSignalRefresh()
})

watch(showSignalDialog, (visible) => {
  if (!visible) {
    stopSignalRefresh()
    return
  }
  syncSignalRefresh()
})

watch(signalSuggestion, () => {
  syncSignalRefresh()
})

async function refreshResults() {
  loading.value = true
  try {
    const data = await listResults()
    results.value = data.results ?? []
  } catch (_error) {
    ElMessage.error('Failed to load results')
  } finally {
    loading.value = false
  }
}

function openDetails(taskId: number) {
  selectedResult.value = results.value.find((item) => item.taskId === taskId) ?? null
  showDetailsDialog.value = true
}

function canOpenResult(taskId: number) {
  const result = results.value.find((item) => item.taskId === taskId)
  return Boolean(result?.logUrl && window.simfox?.openPath)
}

function canOpenInSumoGui(result: ResultListItem) {
  return Boolean(result.configPath && canUseSumoGui.value)
}

async function openResult(taskId: number) {
  const result = results.value.find((item) => item.taskId === taskId)
  if (!result?.logUrl || !window.simfox?.openPath) {
    return
  }

  try {
    await window.simfox.openPath(result.logUrl)
  } catch (_error) {
    ElMessage.error('Failed to open result log')
  }
}

async function openInSumoGui(result: ResultListItem) {
  if (!result.configPath || !window.simfox?.openInSumoGui) {
    return
  }

  try {
    await window.simfox.openInSumoGui(result.configPath, result.projectName || `task-${result.taskId}`)
    ElMessage.success('SUMO GUI launched')
  } catch (_error) {
    ElMessage.error('Failed to open project in SUMO GUI')
  }
}

async function openSignalOptimization(taskId: number) {
  selectedResult.value = results.value.find((item) => item.taskId === taskId) ?? null
  selectedSignalTaskId.value = taskId
  showSignalDialog.value = true
  signalSuggestion.value = null

  try {
    signalSuggestion.value = await getSignalOptimization(taskId)
  } catch (_error) {
    ElMessage.warning('No AI signal suggestion available for this task yet')
    showSignalDialog.value = false
  }
}

async function handleAcceptSignal() {
  if (!signalSuggestion.value) {
    return
  }

  try {
    signalSuggestion.value = await acceptSignalOptimization(signalSuggestion.value.taskId)
    await refreshResults()
    const rerunTaskId = signalSuggestion.value.rerunTaskId
    ElMessage.success(
      rerunTaskId
        ? `Signal adjustment accepted. Optimization rerun task #${rerunTaskId} has been queued.`
        : 'Signal adjustment accepted and optimization rerun queued.',
    )
  } catch (_error) {
    ElMessage.error('Failed to apply AI signal adjustment')
  }
}

async function handleRejectSignal() {
  if (!signalSuggestion.value) {
    return
  }

  try {
    signalSuggestion.value = await rejectSignalOptimization(signalSuggestion.value.taskId, 'User rejected the proposed timing plan')
    ElMessage.success('Signal adjustment rejected')
  } catch (_error) {
    ElMessage.error('Failed to reject AI signal adjustment')
  }
}

async function openRoadEditor() {
  if (!selectedResult.value) {
    return
  }

  if (selectedResult.value.configPath && window.simfox?.openInNetedit) {
    try {
      await window.simfox.openInNetedit(
        selectedResult.value.configPath,
        selectedResult.value.projectName || `task-${selectedResult.value.taskId}`,
      )
      ElMessage.success('netedit launched')
      return
    } catch (_error) {
      ElMessage.warning('Failed to launch netedit, falling back to web road editor')
    }
  }

  if (roadEditorUrl.value) {
    window.open(roadEditorUrl.value, '_blank', 'noopener,noreferrer')
    return
  }

  ElMessage.error('No road editor is available for this result')
}

function describeResult(result: ResultListItem) {
  if (result.status === 'succeeded') {
    return result.logUrl ? 'Simulation completed successfully. Log file is available.' : 'Simulation completed successfully.'
  }
  if (result.status === 'failed') {
    return result.lastError || 'Simulation failed.'
  }
  if (result.status === 'cancelled') {
    return result.lastError || 'Simulation was cancelled.'
  }
  return 'Result available.'
}

function syncSignalRefresh() {
  if (!showSignalDialog.value || !signalSuggestion.value) {
    stopSignalRefresh()
    return
  }

  const status = signalSuggestion.value.rerunTaskStatus
  const shouldPoll = Boolean(
    signalSuggestion.value.rerunTaskId &&
    status &&
    ['queued', 'dispatching', 'running'].includes(status),
  )

  if (!shouldPoll) {
    stopSignalRefresh()
    return
  }

  if (signalRefreshTimer !== null) {
    return
  }

  signalRefreshTimer = window.setInterval(() => {
    void refreshSignalSuggestionInBackground()
  }, 4000)
}

function stopSignalRefresh() {
  if (signalRefreshTimer !== null) {
    window.clearInterval(signalRefreshTimer)
    signalRefreshTimer = null
  }
}

async function refreshSignalSuggestionInBackground() {
  if (!selectedSignalTaskId.value) {
    return
  }

  try {
    signalSuggestion.value = await getSignalOptimization(selectedSignalTaskId.value)
    await refreshResults()
  } catch {
    stopSignalRefresh()
  }
}

function formatDelta(value: number) {
  return value > 0 ? `+${value}` : `${value}`
}

function formatPercent(value: number) {
  return value > 0 ? `+${value}%` : `${value}%`
}

function deltaClass(value: number) {
  if (value < 0) {
    return 'delta-better'
  }
  if (value > 0) {
    return 'delta-worse'
  }
  return 'delta-neutral'
}

function formatDate(value?: string | null) {
  if (!value) {
    return '-'
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}`
}
</script>

<style scoped>
.page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.meta-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px 16px;
  margin-bottom: 16px;
  color: #334155;
}

h2 {
  margin: 0 0 6px;
}

p {
  margin: 0;
  color: #64748b;
}

.signal-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
}

.signal-title {
  font-weight: 600;
  color: #0f172a;
  margin-bottom: 4px;
}

.signal-subtitle,
.signal-next-step,
.signal-reason {
  margin: 0 0 12px;
  color: #64748b;
}

.signal-kpis {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.signal-kpi-label {
  color: #64748b;
  margin-bottom: 8px;
}

.comparison-card {
  margin-top: 16px;
  padding: 16px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #f8fafc;
}

.comparison-header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
  color: #334155;
}

.comparison-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.delta-better {
  color: #15803d;
  margin-top: 6px;
}

.delta-worse {
  color: #b91c1c;
  margin-top: 6px;
}

.delta-neutral {
  color: #64748b;
  margin-top: 6px;
}

.phase-change-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  color: #334155;
}
</style>
