<script setup lang="ts">
import type { ChatMessage } from '@/types'

const props = defineProps<{
  message: ChatMessage
}>()

const emit = defineEmits<{
  'addToCanvas': [content: string]
}>()

const isUser = props.message.role === 'user'
</script>

<template>
  <div class="chat-message" :class="{ user: isUser, assistant: !isUser }">
    <div class="msg-role">{{ isUser ? '👤 你' : '🤖 AI' }}</div>
    <div class="msg-content">{{ message.content }}</div>
    <div v-if="!isUser && message.content.length > 20" class="msg-actions">
      <button class="btn-ghost" @click="emit('addToCanvas', message.content)">📝 拆句到画布</button>
    </div>
  </div>
</template>

<style scoped>
.chat-message {
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  margin-bottom: 2px;
  transition: background 0.2s;
}

.chat-message.user {
  background: var(--color-accent-soft);
  border: 1px solid rgba(129,140,248,0.12);
}

.chat-message.assistant {
  background: rgba(255,255,255,0.02);
  border: 1px solid transparent;
}

.chat-message:hover {
  background: var(--color-bg-hover);
}

.msg-role {
  font-size: 11px;
  color: var(--color-text-muted);
  margin-bottom: 5px;
  font-weight: 600;
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.msg-content {
  font-size: 13px;
  line-height: 1.8;
  white-space: pre-wrap;
  word-break: break-word;
}

.msg-actions {
  margin-top: 8px;
  opacity: 0;
  transform: translateY(-2px);
  transition: all 0.2s ease;
}

.chat-message:hover .msg-actions {
  opacity: 1;
  transform: translateY(0);
}

.msg-actions button {
  font-size: 11px;
  padding: 4px 10px;
  border-radius: 6px;
  font-weight: 500;
  border: 1px solid var(--color-border);
}
</style>
