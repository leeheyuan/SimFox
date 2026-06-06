import axios from 'axios'
import { resolveApiBase } from './base'

const simulationHttp = axios.create({
  baseURL: resolveApiBase('/simulation-api'),
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

simulationHttp.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = token.startsWith('Bearer ') ? token : `Bearer ${token}`
  }
  return config
})

simulationHttp.interceptors.response.use(
  (res) => res.data,
  (err) => {
    console.error('simulation request error:', err)
    return Promise.reject(err)
  }
)

export default simulationHttp

export interface ProjectListItem {
  id: number
  name: string
  status: string
  createdAt: string
  updatedAt: string
}

export interface ProjectListResponse {
  projects: ProjectListItem[]
}

export interface ProjectFileItem {
  name: string
  relativePath: string
  absolutePath: string
  extension: string
  size: number
  modifiedAt: string
}

export interface ProjectFilesResponse {
  projectId: number
  projectName: string
  projectRoot: string
  configPath: string
  files: ProjectFileItem[]
}

export interface TaskListItem {
  id: number
  projectId: number
  projectName: string
  status: string
  progress: number
  workerId?: number | null
  queueName: string
  durationSeconds: number
  speed: number
  monitorPort: number
  traCIPort: number
  logUrl: string
  lastError: string
  createdAt: string
  submittedAt?: string | null
  scheduledAt?: string | null
  startedAt?: string | null
  endedAt?: string | null
}

export interface TaskListResponse {
  tasks: TaskListItem[]
}

export interface ResultListItem {
  taskId: number
  projectId: number
  projectName: string
  status: string
  progress: number
  configPath: string
  logUrl: string
  outputArtifactId?: number | null
  lastError: string
  generatedAt?: string | null
}

export interface ResultListResponse {
  results: ResultListItem[]
}

export interface SignalOptimizationMetrics {
  tripCount: number
  averageDuration: number
  averageWaitingTime: number
  averageTimeLoss: number
  candidateSignals: Array<{
    id: string
    programId: string
    sourceFile: string
    phaseCount: number
    cycleLength: number
  }>
}

export interface SignalOptimizationAnalysis {
  mode: string
  congestedJunctions: string[]
  reasons: string[]
  recommendationLevel: string
}

export interface SignalOptimizationProposal {
  adjustments: Array<{
    junctionId: string
    programId: string
    sourceFile: string
    beforeCycle: number
    afterCycle: number
    phaseChanges: Array<{
      index: number
      state: string
      oldDuration: number
      newDuration: number
      comment: string
    }>
  }>
  nextStep: string
}

export interface SignalOptimizationSuggestion {
  taskId: number
  projectId: number
  rerunTaskId?: number | null
  rerunTaskStatus?: string
  status: string
  engine: string
  summary: string
  metrics: SignalOptimizationMetrics
  analysis: SignalOptimizationAnalysis
  proposal: SignalOptimizationProposal
  comparison?: {
    baselineTaskId: number
    optimizedTaskId: number
    optimizedTaskStatus: string
    baselineMetrics: SignalOptimizationMetrics
    optimizedMetrics?: SignalOptimizationMetrics
    delta?: {
      tripCountDelta: number
      averageDurationDelta: number
      averageDurationDeltaPercent: number
      averageWaitingTimeDelta: number
      averageWaitingTimeDeltaPercent: number
      averageTimeLossDelta: number
      averageTimeLossDeltaPercent: number
    }
  } | null
  proposedConfigPath: string
  appliedConfigPath: string
  appliedSignalFile: string
  rejectedReason: string
  createdAt: string
  updatedAt: string
  reviewedAt?: string | null
}

export interface WorkerListItem {
  id: number
  name: string
  status: string
  address: string
  queueName: string
  maxConcurrency: number
  runningTasks: number
  lastHeartbeat?: string | null
  updatedAt: string
}

export interface WorkerListResponse {
  workers: WorkerListItem[]
}

export interface RegisterWorkerPayload {
  name: string
  address: string
  queueName: string
  labelsJson: string
  maxConcurrency: number
}

export interface RegisterWorkerResponse {
  workerId: number
  workerSecret: string
}

export interface ImportProjectPayload {
  projectName: string
  simulationTime?: number
  speed?: number
  isNowRun?: boolean
  configFile: { file: File; relativePath: string }
  projectFiles: Array<{ file: File; relativePath: string }>
}

export interface EnqueueTaskPayload {
  simulationTime: number
  speed: number
}

export function listProjects(): Promise<ProjectListResponse> {
  return simulationHttp.get('/project/projects')
}

export function listProjectFiles(projectId: number): Promise<ProjectFilesResponse> {
  return simulationHttp.get(`/project/${projectId}/files`)
}

export function deleteProject(projectId: number): Promise<{ projectId: number }> {
  return simulationHttp.delete(`/project/${projectId}`)
}

export function listTasks(): Promise<TaskListResponse> {
  return simulationHttp.get('/project/tasks')
}

export function listResults(): Promise<ResultListResponse> {
  return simulationHttp.get('/project/results')
}

export function getSignalOptimization(taskId: number): Promise<SignalOptimizationSuggestion> {
  return simulationHttp.get(`/project/results/${taskId}/signal-optimization`)
}

export function acceptSignalOptimization(taskId: number): Promise<SignalOptimizationSuggestion> {
  return simulationHttp.post(`/project/results/${taskId}/signal-optimization/accept`)
}

export function rejectSignalOptimization(taskId: number, reason = ''): Promise<SignalOptimizationSuggestion> {
  return simulationHttp.post(`/project/results/${taskId}/signal-optimization/reject`, { reason })
}

export function cancelTask(taskId: number): Promise<{ taskId: number; status: string }> {
  return simulationHttp.post(`/project/tasks/${taskId}/cancel`)
}

export function retryTask(taskId: number): Promise<{ taskId: number }> {
  return simulationHttp.post(`/project/tasks/${taskId}/retry`)
}

export function listWorkers(): Promise<WorkerListResponse> {
  return simulationHttp.get('/worker/list')
}

export function registerWorker(payload: RegisterWorkerPayload): Promise<RegisterWorkerResponse> {
  return simulationHttp.post('/worker/register', payload)
}

export function enqueueProjectTask(projectId: number, payload: EnqueueTaskPayload): Promise<{ taskId: number }> {
  return simulationHttp.post(`/project/${projectId}/run`, payload)
}

export async function importProjectFromConfig(payload: ImportProjectPayload): Promise<any> {
  const formData = new FormData()
  formData.append('name', payload.projectName)
  formData.append('simulationTime', String(payload.simulationTime ?? 60))
  formData.append('speed', String(payload.speed ?? 1))
  formData.append('isNowRun', String(payload.isNowRun ?? false))
  formData.append(
    'manifest',
    JSON.stringify({
      configPath: payload.configFile.relativePath,
      projectFiles: payload.projectFiles.map(({ relativePath }) => relativePath),
    }),
  )
  formData.append('configFile', payload.configFile.file, payload.configFile.relativePath)

  payload.projectFiles.forEach(({ file, relativePath }) => {
    formData.append('projectFiles', file, relativePath)
  })

  return simulationHttp.post('/project/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}
