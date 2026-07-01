<template>
  <Modal :show="show" title="🔒 修改密码" :maxWidth="'440px'" @close="$emit('close')">
    <div class="form-group"><label>当前密码</label><input type="password" v-model="oldPwd" placeholder="请输入当前密码" /></div>
    <div class="form-group"><label>新密码</label><input type="password" v-model="newPwd" placeholder="请输入新密码" /></div>
    <div :class="['error-msg', { show: errorMsg }]">{{ errorMsg }}</div>
    <div class="modal-buttons">
      <button class="modal-btn cancel" @click="$emit('close')">取消</button>
      <button class="modal-btn confirm" :disabled="loading" @click="changePwd">{{ loading ? '修改中...' : '确认修改' }}</button>
    </div>
  </Modal>
</template>

<script setup>
import { ref } from 'vue'
import { api, showToast } from './shared'
import Modal from '@/components/Modal.vue'

defineProps({ show: Boolean })
const emit = defineEmits(['close'])

const oldPwd = ref('')
const newPwd = ref('')
const errorMsg = ref('')
const loading = ref(false)

async function changePwd() {
  if (!oldPwd.value.trim() || !newPwd.value.trim()) {
    errorMsg.value = '请填写完整信息'; return
  }
  loading.value = true; errorMsg.value = ''
  try {
    const res = await api('/auth/password', {
      method: 'PUT',
      body: JSON.stringify({ oldPassword: oldPwd.value, newPassword: newPwd.value })
    })
    if (res.code === 200) {
      showToast('密码修改成功', 'success')
      emit('close')
    } else {
      errorMsg.value = res.message || '修改失败'
    }
  } catch (e) {
    errorMsg.value = '网络错误'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.error-msg { color: #ef4444; font-size: 13px; text-align: center; margin: 10px 0; opacity: 0; max-height: 0; overflow: hidden; transition: all 0.3s; }
.error-msg.show { opacity: 1; max-height: 40px; }
.modal-buttons { display: flex; gap: 10px; justify-content: flex-end; margin-top: 20px; }
.modal-btn {
  padding: 10px 24px; border: none; border-radius: var(--radius-sm);
  font-size: 14px; font-weight: 600; cursor: pointer; transition: all var(--transition);
}
.modal-btn.cancel { background: #f1f5f9; color: var(--text-secondary); }
.modal-btn.cancel:hover { background: #e2e8f0; }
.modal-btn.confirm { background: linear-gradient(135deg, #6366f1, #8b5cf6); color: white; }
.modal-btn.confirm:hover { box-shadow: 0 4px 16px rgba(99,102,241,0.35); }
.modal-btn:disabled { opacity: 0.6; cursor: not-allowed; }
</style>
