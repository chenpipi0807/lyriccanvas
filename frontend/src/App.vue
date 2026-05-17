<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useProjectStore } from './stores/project'
import { useChatStore } from './stores/chat'
import { getConfig } from './api/config'
import AppLayout from './components/layout/AppLayout.vue'
import SettingsModal from './components/layout/SettingsModal.vue'

const projectStore = useProjectStore()
const chatStore = useChatStore()
const showSettings = ref(false)

onMounted(async () => {
  // 检测 API Key 是否已设置
  try {
    const cfg = await getConfig()
    if (!cfg.hasApiKey) {
      showSettings.value = true
    }
  } catch { /* ignore */ }

  await projectStore.loadList()
  await chatStore.loadSystemPrompt()
  if (projectStore.list.length > 0) {
    await projectStore.loadProject(projectStore.list[0].id)
    await chatStore.loadHistory()
  }
})
</script>

<template>
  <AppLayout />
  <SettingsModal v-if="showSettings" @close="showSettings = false" />
</template>
