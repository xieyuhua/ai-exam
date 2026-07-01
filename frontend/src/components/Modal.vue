<template>
  <Teleport to="body">
    <div :class="['modal-overlay', { show }]" @click.self="$emit('close')">
      <div class="modal" :style="{ maxWidth: maxWidth }">
        <h2 v-if="title">{{ title }}</h2>
        <slot />
      </div>
    </div>
  </Teleport>
</template>

<script setup>
defineProps({
  show: Boolean,
  title: String,
  maxWidth: { type: String, default: '650px' }
})
defineEmits(['close'])
</script>

<style scoped>
.modal-overlay {
  display: none; position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(15,23,42,0.5); backdrop-filter: blur(2px);
  z-index: 1000; align-items: center; justify-content: center;
}
.modal-overlay.show { display: flex; }
.modal {
  background: var(--white); border-radius: 16px; padding: 32px;
  width: 90%; max-height: 90vh; overflow-y: auto;
  animation: modalIn 0.25s ease;
  box-shadow: var(--shadow-lg);
}
@keyframes modalIn {
  from { transform: scale(0.95) translateY(10px); opacity: 0; }
  to { transform: scale(1) translateY(0); opacity: 1; }
}
.modal h2 { color: var(--text); margin-bottom: 20px; font-size: 18px; font-weight: 700; }
</style>
