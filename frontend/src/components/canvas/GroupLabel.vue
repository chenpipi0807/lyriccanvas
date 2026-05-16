<script setup lang="ts">
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'

const props = defineProps<{
  id: string
  data: {
    id: string
    type: 'structure'
    label: string
    category: string
    color: string
    nodeIds: string[]
  }
}>()

const categoryColors: Record<string, string> = {
  intro: '#9CA3AF',
  verse: '#60A5FA',
  'pre-chorus': '#818CF8',
  chorus: '#F472B6',
  bridge: '#FBBF24',
  outro: '#9CA3AF',
  interlude: '#34D399',
}

const bgColor = computed(() => props.data.color || categoryColors[props.data.category] || '#4A90D9')
</script>

<template>
  <div class="group-label" :style="{ borderColor: bgColor }">
    <div class="label-dot" :style="{ background: bgColor }" />
    <span class="label-text">{{ data.label }}</span>
    <span class="label-count">{{ data.nodeIds.length }} 句</span>

    <Handle type="source" :position="Position.Bottom" :style="{ visibility: 'hidden' }" />
    <Handle type="target" :position="Position.Top" :style="{ visibility: 'hidden' }" />
  </div>
</template>

<style scoped>
.group-label {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 16px;
  border-radius: 10px;
  border: 1.5px solid;
  background: linear-gradient(135deg, rgba(255,255,255,0.04) 0%, rgba(255,255,255,0.01) 100%);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  font-size: 13px;
  font-weight: 500;
  cursor: grab;
  user-select: none;
  min-width: 130px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.2);
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.group-label:hover {
  box-shadow: 0 4px 20px rgba(0,0,0,0.35);
  transform: translateY(-1px);
}

.label-dot {
  width: 9px;
  height: 9px;
  border-radius: 50%;
  flex-shrink: 0;
  box-shadow: 0 0 8px currentColor;
}

.label-text {
  font-weight: 600;
  color: var(--color-text);
  letter-spacing: 0.02em;
}

.label-count {
  font-size: 11px;
  color: var(--color-text-muted);
  margin-left: auto;
  background: rgba(255,255,255,0.05);
  padding: 2px 8px;
  border-radius: 10px;
}
</style>
