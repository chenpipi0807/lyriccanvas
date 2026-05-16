<script setup lang="ts">
import ProjectSidebar from '../sidebar/ProjectSidebar.vue'
import LyricCanvas from '../canvas/LyricCanvas.vue'
import ChatPanel from '../chat/ChatPanel.vue'
</script>

<template>
  <div class="app-layout">
    <ProjectSidebar />
    <main class="canvas-area">
      <LyricCanvas />
    </main>
    <ChatPanel />
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  width: 100%;
  height: 100%;
  position: relative;
}

.canvas-area {
  flex: 1;
  position: relative;
  overflow: hidden;
  background:
    radial-gradient(ellipse at 60% 40%, rgba(59,130,246,0.02) 0%, transparent 70%),
    radial-gradient(ellipse at 30% 70%, rgba(59,130,246,0.01) 0%, transparent 60%),
    var(--color-bg);
}
</style>
