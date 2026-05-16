<script setup lang="ts">
import type { ProjectListItem } from '@/types'

const props = defineProps<{
  project: ProjectListItem
  active: boolean
}>()

const emit = defineEmits<{
  select: []
  delete: []
  rename: []
}>()

function formatDate(iso: string): string {
  const d = new Date(iso)
  return `${d.getMonth() + 1}/${d.getDate()}`
}
</script>

<template>
  <div
    class="project-card"
    :class="{ active }"
    @click="emit('select')"
    @contextmenu.prevent
  >
    <div class="card-main">
      <span class="card-icon">🎵</span>
      <div class="card-info">
        <div class="card-name" :title="project.name">{{ project.name }}</div>
        <div class="card-meta">
          {{ formatDate(project.updatedAt) }} · {{ project.nodeCount }} 句
        </div>
      </div>
    </div>
    <div class="card-actions">
      <button class="btn-ghost" title="重命名" @click.stop="emit('rename')">✏️</button>
      <button class="btn-danger" title="删除" @click.stop="emit('delete')">🗑️</button>
    </div>
  </div>
</template>

<style scoped>
.project-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.18s ease;
  margin-bottom: 3px;
  border: 1px solid transparent;
}

.project-card:hover {
  background: var(--color-bg-hover);
}

.project-card.active {
  background: var(--color-accent-soft);
  border-color: rgba(59, 130, 246, 0.2);
}

.card-main {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  flex: 1;
}

.card-icon {
  font-size: 20px;
  flex-shrink: 0;
  opacity: 0.8;
}

.card-info {
  min-width: 0;
}

.card-name {
  font-size: 13px;
  font-weight: 550;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  letter-spacing: 0.01em;
}

.card-meta {
  font-size: 11px;
  color: var(--color-text-muted);
  margin-top: 3px;
  font-weight: 400;
}

.card-actions {
  display: flex;
  gap: 2px;
  flex-shrink: 0;
  opacity: 0;
  transition: opacity 0.18s;
}

.project-card:hover .card-actions {
  opacity: 1;
}

.card-actions button {
  padding: 3px 7px;
  font-size: 12px;
}
</style>
