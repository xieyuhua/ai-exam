import { reactive } from 'vue'

export const state = reactive({ show: false, message: '', type: 'info' })

let timer = null

export function showToast(message, type = 'info') {
  state.message = message
  state.type = type
  state.show = true
  clearTimeout(timer)
  timer = setTimeout(() => { state.show = false }, 3000)
}
