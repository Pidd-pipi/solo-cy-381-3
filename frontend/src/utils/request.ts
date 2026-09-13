// request.ts 统一 axios 实例：拦截器处理 JWT、错误码、全局消息（横切关注点 3 前端触达层）
import axios, { type AxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

export interface PageData<T = unknown> {
  list: T[]
  total: number
  page: number
  page_size: number
}

const instance = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

instance.interceptors.request.use((config) => {
  const token = localStorage.getItem('aasplit_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

instance.interceptors.response.use(
  (response) => {
    const body = response.data as ApiResponse
    if (body && typeof body.code === 'number' && body.code !== 0) {
      ElMessage.error(body.message || '请求失败')
      return Promise.reject(new Error(body.message))
    }
    return response
  },
  (error) => {
    const status = error.response?.status
    const body = error.response?.data as ApiResponse | undefined
    const msg = body?.message || (status === 401 ? '登录已失效，请重新登录' : '网络异常，请稍后重试')
    if (status === 401) {
      localStorage.removeItem('aasplit_token')
      localStorage.removeItem('aasplit_user')
      router.push('/login')
    }
    ElMessage.error(msg)
    return Promise.reject(error)
  },
)

export async function get<T = unknown>(url: string, params?: object): Promise<T> {
  const res = await instance.get<ApiResponse<T>>(url, { params })
  return res.data.data as T
}

export async function post<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
  const res = await instance.post<ApiResponse<T>>(url, data, config)
  return res.data.data as T
}

export async function put<T = unknown>(url: string, data?: unknown): Promise<T> {
  const res = await instance.put<ApiResponse<T>>(url, data)
  return res.data.data as T
}

export async function del<T = unknown>(url: string): Promise<T> {
  const res = await instance.delete<ApiResponse<T>>(url)
  return res.data.data as T
}

export default instance
