import http from './index'

export function login(username: string, password: string): Promise<boolean> {
  return http.post('/login', {
  username: username,
  password: password
})
    .then(res => {
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

export function getUserInfo(): Promise<boolean> {
  return http.get('/overview')
}


export async function generate(
  formData: FormData,
  onProgress?: (percent: number) => void
): Promise<boolean> {
  try {
    await http.post('/project/generate', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (progressEvent: ProgressEvent) => {
        if (progressEvent.lengthComputable && onProgress) {
          const percent = (progressEvent.loaded / progressEvent.total) * 100
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
    await http.post('/uploadmap', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (progressEvent: ProgressEvent) => {
        if (progressEvent.lengthComputable && onProgress) {
          const percent = (progressEvent.loaded / progressEvent.total) * 100
          onProgress(percent)
        }
      },
    })
    return true
  } catch (error) {
    return false
  }
}


export function getOverview(): Promise<boolean> {
  return http.get('/overview')
}