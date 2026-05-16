<script setup lang="ts">
import { ref, computed } from 'vue'
import { useChatStore } from '@/stores/chat'
import type { TemplateCategory, QuickTemplate } from '@/types'

const chatStore = useChatStore()

const activeCategory = ref('structure')

const categories = computed(() => chatStore.templateCategories)
const currentCategory = computed(() => categories.value.find((c) => c.key === activeCategory.value))
const currentTemplates = computed(() => currentCategory.value?.templates ?? [])

const emit = defineEmits<{
  select: [content: string]
}>()

function selectTemplate(tmpl: QuickTemplate) {
  emit('select', tmpl.content)
}
</script>

<template>
  <div class="template-panel">
    <div class="tmpl-tabs">
      <button
        v-for="cat in categories"
        :key="cat.key"
        class="tmpl-tab"
        :class="{ active: activeCategory === cat.key }"
        @click="activeCategory = cat.key"
      >
        {{ cat.label }}
      </button>
    </div>
    <div class="tmpl-list">
      <div
        v-for="tmpl in currentTemplates"
        :key="tmpl.id"
        class="tmpl-item"
        @click="selectTemplate(tmpl)"
      >
        <div class="tmpl-label">{{ tmpl.label }}</div>
        <div v-if="tmpl.hint" class="tmpl-hint">{{ tmpl.hint }}</div>
      </div>
      <div v-if="currentTemplates.length === 0" class="tmpl-empty">
        暂无模板
      </div>
    </div>
  </div>
</template>

<style scoped>
.template-panel {
  border-bottom: 1px solid var(--color-border-light);
  background: rgba(0,0,0,0.2);
  max-height: 220px;
  display: flex;
  flex-direction: column;
}

.tmpl-tabs {
  display: flex;
  gap: 2px;
  padding: 6px 8px;
  border-bottom: 1px solid var(--color-border-light);
  overflow-x: auto;
}

.tmpl-tab {
  font-size: 11px;
  font-weight: 500;
  padding: 4px 12px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s;
}

.tmpl-tab:hover {
  background: rgba(255,255,255,0.05);
  color: var(--color-text);
}

.tmpl-tab.active {
  background: var(--color-accent-soft);
  color: var(--color-accent);
}

.tmpl-list {
  flex: 1;
  overflow-y: auto;
  padding: 6px 8px;
}

.tmpl-item {
  padding: 7px 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
  margin-bottom: 2px;
}

.tmpl-item:hover {
  background: rgba(255,255,255,0.05);
}

.tmpl-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--color-text);
}

.tmpl-hint {
  font-size: 11px;
  color: var(--color-text-muted);
  margin-top: 2px;
}

.tmpl-empty {
  padding: 16px;
  text-align: center;
  font-size: 12px;
  color: var(--color-text-muted);
}
</style>
