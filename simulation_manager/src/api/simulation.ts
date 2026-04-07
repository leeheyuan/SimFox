import axios from 'axios'

const simulationHttp = axios.create({
  baseURL: '/simulation-api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

simulationHttp.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = token
  }
  return config
})

simulationHttp.interceptors.response.use(
  (res) => res.data,
  (err) => {
    console.error('仿真服务请求错误:', err)
    return Promise.reject(err)
  }
)

export default simulationHttp
