// api/simulation.ts
import axios from 'axios' 
const  configHttp = axios.create({
  baseURL: '/config',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

configHttp.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

configHttp.interceptors.response.use(
  res => res.data,
  err => {
    console.error('仿真请求错误:', err)
    return Promise.reject(err)
  }
)


export function getGeojson(netfile :string): Promise<any> {
  return configHttp.get(`/getGeojson?netfile=${netfile}`)
}
