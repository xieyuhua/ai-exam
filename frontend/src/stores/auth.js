import { ref } from 'vue'
import { useRouter } from 'vue-router'

const API_BASE = '/api'

// 全局状态
export const currentUser = ref(null)
export const isAuthenticated = ref(false)

// Token 管理
export function getToken() {
  const match = document.cookie.match(/(?:^|;\s*)exam_token=([^;]*)/)
  if (match) return match[1]
  return localStorage.getItem('exam_token')
}

export function setToken(token) {
  document.cookie = `exam_token=${token};path=/;max-age=86400;SameSite=Lax`
  localStorage.setItem('exam_token', token)
}

export function clearToken() {
  localStorage.removeItem('exam_token')
  document.cookie = 'exam_token=;path=/;max-age=0'
}

// API 封装
export async function api(url, options = {}) {
  const headers = { 'Content-Type': 'application/json', ...options.headers }
  const token = getToken()
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(API_BASE + url, { headers, ...options })
  if (res.status === 401) {
    clearToken()
    isAuthenticated.value = false
    currentUser.value = null
    window.location.href = '/'
    return { code: 401, message: '未登录' }
  }
  return res.json()
}

// 检查登录状态
export async function checkAuth() {
  const token = getToken()
  if (!token) return false
  try {
    const res = await api('/auth/me')
    if (res.code === 200 && res.data) {
      currentUser.value = res.data
      isAuthenticated.value = true
      return true
    }
  } catch (e) { /* ignore */ }
  return false
}

// 登出
export async function logout() {
  try { await api('/auth/logout', { method: 'POST' }) } catch (e) { /* ignore */ }
  clearToken()
  isAuthenticated.value = false
  currentUser.value = null
}
