<template>
  <Teleport to="body">
    <div :class="['toast', state.type, { show: state.show }]">{{ state.message }}</div>
  </Teleport>
</template>

<script setup>
import { state } from './toastState.js'
</script>

<style scoped>
.toast {
  position: fixed; top: 24px; left: 50%;
  transform: translateX(-50%) translateY(-120px);
  padding: 12px 28px; border-radius: 10px;
  font-size: 14px; z-index: 3000;
  transition: transform 0.35s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  box-shadow: 0 8px 24px rgba(0,0,0,0.18); white-space: nowrap;
  pointer-events: none; font-weight: 500;
}
.toast.show { transform: translateX(-50%) translateY(0); }
.toast.success { background: #10b981; color: white; }
.toast.error { background: #ef4444; color: white; }
.toast.info { background: #334155; color: white; }
</style>
