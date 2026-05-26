import axios from 'axios'
import { resolveApiBase } from './base'

const configHttp = axios.create({
  baseURL: resolveApiBase('/config'),
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

configHttp.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = token.startsWith('Bearer ') ? token : `Bearer ${token}`
  }
  return config
})

configHttp.interceptors.response.use(
  (res) => res.data,
  (err) => {
    console.error('config request error:', err)
    return Promise.reject(err)
  }
)

export function getGeojson(netfile: string): Promise<any> {
  return configHttp.get(`/getGeojson?netfile=${netfile}`)
}
