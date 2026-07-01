<template>
  <header class="topbar">
    <h1>{{ title }}</h1>
    <div class="topbar-right">
      <div class="user-info">
        <div class="avatar">{{ initial }}</div>
        <span>{{ displayName }}</span>
      </div>
      <slot name="actions" />
      <button class="btn-topbar" @click="$emit('changePassword')" v-if="showChangePassword">修改密码</button>
      <button class="btn-topbar danger" @click="$emit('logout')">退出登录</button>
    </div>
  </header>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  title: { type: String, default: '' },
  userName: { type: String, default: '' },
  showChangePassword: { type: Boolean, default: false }
})

defineEmits(['logout', 'changePassword'])

const displayName = computed(() => props.userName || '--')
const initial = computed(() => (props.userName || '?')[0].toUpperCase())
</script>

<style scoped>
.topbar {
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  color: white; padding: 16px 32px;
  display: flex; justify-content: space-between; align-items: center;
  box-shadow: 0 2px 12px rgba(99,102,241,0.2);
  position: sticky; top: 0; z-index: 100;
}
.topbar h1 { font-size: 20px; font-weight: 700; letter-spacing: 0.5px; }
.topbar-right { display: flex; align-items: center; gap: 12px; }
.user-info {
  display: flex; align-items: center; gap: 8px;
  padding: 6px 14px; background: rgba(255,255,255,0.15);
  border-radius: 20px; font-size: 13px; font-weight: 500;
}
.avatar {
  width: 28px; height: 28px; border-radius: 50%;
  background: rgba(255,255,255,0.25);
  display: flex; align-items: center; justify-content: center; font-size: 14px;
}
.btn-topbar {
  background: rgba(255,255,255,0.15); color: white; border: 1px solid rgba(255,255,255,0.2);
  padding: 7px 16px; border-radius: 8px; cursor: pointer; font-size: 13px;
  font-weight: 500; transition: all var(--transition);
}
.btn-topbar:hover { background: rgba(255,255,255,0.25); }
.btn-topbar.danger { background: rgba(239,68,68,0.3); border-color: rgba(239,68,68,0.4); }
.btn-topbar.danger:hover { background: rgba(239,68,68,0.5); }

@media (max-width: 768px) {
  .topbar { padding: 12px 14px; }
  .topbar h1 { font-size: 16px; }
  .topbar-right { gap: 6px; }
  .user-info { padding: 4px 10px; font-size: 12px; }
  .btn-topbar { padding: 5px 10px; font-size: 11px; }
}
</style>
