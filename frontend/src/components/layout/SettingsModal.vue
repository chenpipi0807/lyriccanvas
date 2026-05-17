<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getConfig, setApiKey } from '@/api/config'

const emit = defineEmits<{
  close: []
}>()

const hasKey = ref(false)
const keyPreview = ref('')
const keyInput = ref('')
const saving = ref(false)
const message = ref('')
const messageType = ref<'success' | 'error'>('success')
const showKey = ref(false)

async function loadConfig() {
  try {
    const cfg = await getConfig()
    hasKey.value = cfg.hasApiKey
    keyPreview.value = cfg.apiKeyPreview
  } catch {
    // ignore
  }
}

async function handleSave() {
  const val = keyInput.value.trim()
  if (!val) {
    message.value = '请输入 API Key'
    messageType.value = 'error'
    return
  }
  if (!val.startsWith('sk-')) {
    message.value = 'API Key 格式不正确，应以 sk- 开头'
    messageType.value = 'error'
    return
  }

  saving.value = true
  message.value = ''
  try {
    await setApiKey(val)
    message.value = '✅ API Key 已保存并生效'
    messageType.value = 'success'
    hasKey.value = true
    keyPreview.value = val.length > 12 ? val.slice(0, 5) + '****' + val.slice(-2) : val.slice(0, 3) + '****'
    keyInput.value = ''
  } catch (e: any) {
    message.value = '❌ 保存失败: ' + (e?.response?.data?.error || e.message || '未知错误')
    messageType.value = 'error'
  } finally {
    saving.value = false
  }
}

onMounted(loadConfig)
</script>

<template>
  <div class="settings-overlay" @click.self="emit('close')">
    <div class="settings-modal">
      <div class="settings-header">
        <h2>⚙️ 设置</h2>
        <button class="close-btn" @click="emit('close')">✕</button>
      </div>

      <div class="settings-body">
        <!-- API Key 状态 -->
        <div class="setting-section">
          <label class="setting-label">🤖 DeepSeek API Key</label>
          <p class="setting-hint" v-if="hasKey">
            当前 Key: <code>{{ keyPreview }}</code>
          </p>
          <p class="setting-hint" v-else>
            ⚠️ 未设置 API Key，AI 功能不可用。请在下方填入你的 Key。
          </p>

          <div class="key-input-row">
            <input
              v-model="keyInput"
              :type="showKey ? 'text' : 'password'"
              class="key-input"
              placeholder="sk-..."
              @keyup.enter="handleSave"
            />
            <button class="toggle-vis" @click="showKey = !showKey" type="button">
              {{ showKey ? '🙈' : '👁️' }}
            </button>
          </div>

          <button class="btn-save" @click="handleSave" :disabled="saving">
            {{ saving ? '⏳ 保存中...' : '💾 保存 Key' }}
          </button>

          <p class="setting-msg" :class="messageType" v-if="message">{{ message }}</p>
        </div>

        <div class="setting-section">
          <p class="setting-hint-secondary">
            💡 前往 <a href="https://platform.deepseek.com/" target="_blank">platform.deepseek.com</a> 获取 API Key（新用户有免费额度）
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.settings-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 80px;
}

.settings-modal {
  background: #1a1a2e;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  width: 480px;
  max-width: 90vw;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  overflow: hidden;
}

.settings-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.settings-header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #e0e0e0;
}

.close-btn {
  background: none;
  border: none;
  color: #888;
  font-size: 18px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

.settings-body {
  padding: 24px;
}

.setting-section {
  margin-bottom: 20px;
}

.setting-label {
  display: block;
  font-size: 14px;
  font-weight: 600;
  color: #ccc;
  margin-bottom: 6px;
}

.setting-hint {
  font-size: 13px;
  color: #888;
  margin: 0 0 10px;
}

.setting-hint code {
  background: rgba(255, 255, 255, 0.08);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 12px;
  color: #a0d2ff;
}

.setting-hint-secondary {
  font-size: 13px;
  color: #666;
  margin: 0;
}

.setting-hint-secondary a {
  color: #3b82f6;
  text-decoration: none;
}

.setting-hint-secondary a:hover {
  text-decoration: underline;
}

.key-input-row {
  display: flex;
  gap: 6px;
  margin-bottom: 10px;
}

.key-input {
  flex: 1;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 8px;
  padding: 10px 14px;
  font-size: 14px;
  color: #e0e0e0;
  font-family: monospace;
  outline: none;
  transition: border-color 0.2s;
}

.key-input:focus {
  border-color: #3b82f6;
}

.key-input::placeholder {
  color: #555;
}

.toggle-vis {
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 8px;
  padding: 8px 10px;
  cursor: pointer;
  font-size: 16px;
  color: #888;
  transition: all 0.2s;
}

.toggle-vis:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.btn-save {
  width: 100%;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  border: none;
  border-radius: 8px;
  padding: 10px 16px;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-save:hover:not(:disabled) {
  background: linear-gradient(135deg, #4f8dff, #3864f0);
  transform: translateY(-1px);
}

.btn-save:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.setting-msg {
  margin-top: 10px;
  font-size: 13px;
  font-weight: 500;
}

.setting-msg.success {
  color: #34d399;
}

.setting-msg.error {
  color: #f87171;
}
</style>
