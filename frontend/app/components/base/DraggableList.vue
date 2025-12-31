<script setup lang="ts" generic="T extends { id: string }">
import draggable from 'vuedraggable'
import type { DragChangeEvent } from '~/types/dragEvents'
import { useDraggableList } from '~/composables/useDraggableList'

interface Props {
  /** 列表項目 */
  items: T[]
  /** 拖曳組名稱 */
  groupName: string
  /** 是否禁用拖曳 */
  disabled?: boolean
  /** 動畫時長（毫秒） */
  animation?: number
  /** item-key 屬性 */
  itemKey?: string
}

interface Emits {
  (e: 'update:items', items: T[]): void
  (e: 'reorder', itemIds: string[]): void
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  animation: 200,
  itemKey: 'id'
})

const emit = defineEmits<Emits>()

// 本地列表（用於拖曳）
const localItems = computed({
  get: () => props.items,
  set: (newItems) => {
    emit('update:items', newItems)
  }
})

const { dragOptions, handleChange } = useDraggableList<T>(localItems, {
  groupName: props.groupName,
  disabled: props.disabled,
  animation: props.animation
})

// 處理拖曳變更
const onChange = (evt: DragChangeEvent<T>) => {
  handleChange(evt, (itemIds) => {
    emit('reorder', itemIds)
  })
}
</script>

<template>
  <draggable
    v-model="localItems"
    v-bind="dragOptions"
    :item-key="itemKey"
    @change="onChange"
    class="space-y-2"
  >
    <template #item="{ element, index }">
      <slot name="item" :element="element" :index="index" />
    </template>
  </draggable>
</template>

<style scoped>
/* 統一的拖曳效果樣式 */
:deep(.drag-ghost) {
  opacity: 0.5;
  background: rgba(var(--color-primary-500) / 0.1);
}

:deep(.drag-chosen) {
  transform: scale(1.02);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  z-index: 10;
}

:deep(.drag-dragging) {
  cursor: grabbing !important;
  z-index: 1000;
}

:deep(.drag-handle) {
  cursor: grab;
}

:deep(.drag-handle:active) {
  cursor: grabbing;
}
</style>

