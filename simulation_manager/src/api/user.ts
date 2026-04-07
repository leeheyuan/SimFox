import type { AxiosProgressEvent } from 'axios'
import http from './index'
import simulationHttp from './simulation'

export interface OverviewProjectItem {
  id?: number
  name: string
  updated?: string
  status: string
}

export interface OverviewResponse {
  map_count: number
  project_count: number
  simulation_count: number
  recent_projects: OverviewProjectItem[]
}

export function login(username: string, password: string): Promise<boolean> {
  return http.post('/login', {
  username: username,
  password: password
})
    .then((res: any) => {
      const token = res.token;
      if (token) {
        localStorage.setItem('token', token); // 持久化 token
        return true;  // 表示登录成功
      } else {
        return false; // 登录失败
      }
    })
    .catch(err => {
      console.error('登录失败：', err.response?.data?.error || err.message);
      return false;
    });
}


export function register(username: string,password: string) {
  return http.post('/register', { username ,password})
}

export function getUserInfo(): Promise<OverviewResponse> {
  return simulationHttp.get('/overview')
}


export async function generate(
  formData: FormData,
  onProgress?: (percent: number) => void
): Promise<boolean> {
  try {
    await simulationHttp.post('/project/generate', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (progressEvent: AxiosProgressEvent) => {
        if (progressEvent.lengthComputable && onProgress) {
          const percent = (progressEvent.loaded / (progressEvent.total || 1)) * 100
          onProgress(percent)
        }
      },
    })
    return true
  } catch (error) {
    return false
  }
}



export async function uploadMap(
  formData: FormData,
  onProgress?: (percent: number) => void
): Promise<boolean> {
  try {
    await simulationHttp.post('/uploadmap', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (progressEvent: AxiosProgressEvent) => {
        if (progressEvent.lengthComputable && onProgress) {
          const percent = (progressEvent.loaded / (progressEvent.total || 1)) * 100
          onProgress(percent)
        }
      },
    })
    return true
  } catch (error) {
    return false
  }
}


export function getOverview(): Promise<OverviewResponse> {
  return simulationHttp.get('/overview')
}

export function importOsmBounds(bounds: {
  south: number
  west: number
  north: number
  east: number
}): Promise<any> {
  return simulationHttp.post('/map/import-bounds', bounds)
}
