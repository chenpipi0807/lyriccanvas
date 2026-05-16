<script setup lang="ts">
import { onMounted } from 'vue'
import { useProjectStore } from './stores/project'
import { useChatStore } from './stores/chat'
import AppLayout from './components/layout/AppLayout.vue'

const projectStore = useProjectStore()
const chatStore = useChatStore()

onMounted(async () => {
  await projectStore.loadList()
  await chatStore.loadSystemPrompt()
  // 如果有项目，自动加载第一个
  if (projectStore.list.length > 0) {
    await projectStore.loadProject(projectStore.list[0].id)
    await chatStore.loadHistory()
  }
})
</script>

<template>
  <AppLayout />
</template>
