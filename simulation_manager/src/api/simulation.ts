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

export interface ImportProjectPayload {
  projectName: string
  simulationTime?: number
  speed?: number
  isNowRun?: boolean
  configFile: { file: File; relativePath: string }
  projectFiles: Array<{ file: File; relativePath: string }>
}

export function listProjects(): Promise<ProjectListResponse> {
  return simulationHttp.get('/project/projects')
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
