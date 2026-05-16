<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  show: boolean
  title: string
  message?: string
  placeholder?: string
  confirmText?: string
  cancelText?: string
  type?: 'info' | 'danger' | 'input'  // info=仅确定, danger=确认取消, input=输入框
  defaultValue?: string
}>()

const emit = defineEmits<{
  close: []
  confirm: [value?: string]
  cancel: []
}>()

const inputValue = ref(props.defaultValue || '')

watch(() => props.show, (val) => {
  if (val) {
    inputValue.value = props.defaultValue || ''
    // focus input after mount
    setTimeout(() => {
      if (props.type === 'input') {
        const el = document.querySelector('.modal-input') as HTMLInputElement
        el?.focus()
      }
    }, 100)
  }
})

function handleConfirm() {
  emit('confirm', props.type === 'input' ? inputValue.value : undefined)
}

function handleCancel() {
  emit('cancel')
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter') handleConfirm()
  if (e.key === 'Escape') handleCancel()
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show" class="modal-overlay" @click.self="handleCancel" @keydown="onKeydown">
        <div class="modal-box" :class="type">
          <div class="modal-header">
            <span class="modal-title">{{ title }}</span>
          </div>

          <div v-if="message" class="modal-message">{{ message }}</div>

          <input
            v-if="type === 'input'"
            v-model="inputValue"
            class="modal-input"
            :placeholder="placeholder || '请输入...'"
            @keydown.enter="handleConfirm"
            @keydown.esc="handleCancel"
          />

          <div class="modal-actions">
            <button
              v-if="type !== 'info'"
              class="modal-btn modal-btn-cancel"
              @click="handleCancel"
            >
              {{ cancelText || '取消' }}
            </button>
            <button
              class="modal-btn modal-btn-confirm"
              :class="{ danger: type === 'danger' }"
              @click="handleConfirm"
            >
              {{ confirmText || '确定' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.55);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
}

.modal-box {
  background: #18181B;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  padding: 24px;
  min-width: 360px;
  max-width: 440px;
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.03),
    0 24px 64px rgba(0, 0, 0, 0.6);
}

.modal-header {
  margin-bottom: 16px;
}

.modal-title {
  font-size: 16px;
  font-weight: 600;
  color: #EDEDEF;
  letter-spacing: 0.01em;
}

.modal-message {
  font-size: 14px;
  color: #8B8B96;
  line-height: 1.6;
  margin-bottom: 20px;
}

.modal-input {
  width: 100%;
  padding: 10px 14px;
  font-size: 14px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  color: #EDEDEF;
  outline: none;
  font-family: inherit;
  margin-bottom: 20px;
  transition: border-color 0.2s;
}

.modal-input:focus {
  border-color: var(--color-accent);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.12);
}

.modal-input::placeholder {
  color: #5C5C66;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.modal-btn {
  padding: 8px 20px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  border: none;
  transition: all 0.15s ease;
}

.modal-btn-cancel {
  background: rgba(255, 255, 255, 0.06);
  color: #A1A1AA;
}

.modal-btn-cancel:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #EDEDEF;
}

.modal-btn-confirm {
  background: var(--color-accent);
  color: #fff;
}

.modal-btn-confirm:hover {
  filter: brightness(1.15);
}

.modal-btn-confirm.danger {
  background: #DC2626;
}

.modal-btn-confirm.danger:hover {
  background: #EF4444;
}

/* Transition */
.modal-enter-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.modal-leave-active {
  transition: all 0.15s cubic-bezier(0.4, 0, 0.2, 1);
}

.modal-enter-from {
  opacity: 0;
}
.modal-enter-from .modal-box {
  transform: scale(0.95) translateY(8px);
  opacity: 0;
}

.modal-leave-to {
  opacity: 0;
}
.modal-leave-to .modal-box {
  transform: scale(0.95) translateY(8px);
  opacity: 0;
}

.modal-box {
  transition: transform 0.2s cubic-bezier(0.4, 0, 0.2, 1), opacity 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
</style>
