import axios from 'axios'
import { resolveApiBase } from './base'

const http = axios.create({
  baseURL: resolveApiBase('/api'),
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = token.startsWith('Bearer ') ? token : `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(
  (res) => res.data,
  (err) => {
    console.error('request error:', err)
    return Promise.reject(err)
  }
)

export default http
