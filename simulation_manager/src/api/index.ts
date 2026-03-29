// api/index.ts
import axios from 'axios'

const http = axios.create({
  baseURL: '/api', // 可在 vite.config.ts 里配置代理
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
http.interceptors.request.use((config) => {
  // 例如添加 token
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `${token}`
  }
  return config
})

// 响应拦截器
http.interceptors.response.use(
  (res) => res.data,
  (err) => {
    console.error('请求错误:', err)
    return Promise.reject(err)
  }
)

export default http
