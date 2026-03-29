import { getUserInfo } from '@/api/user'

export async function checkTokenAndGetUserInfo() {
  const token = localStorage.getItem('token')
  if (!token) return null

  try {
    const res = await getUserInfo() // 假设你有一个获取当前用户信息的接口
    return res // 返回用户信息
  } catch (error) {
    localStorage.removeItem('token')
    return null
  }
}
