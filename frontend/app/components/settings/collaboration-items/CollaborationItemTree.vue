<script setup lang="ts">
import type { CollaborationItem } from '~/types/collaborationItems'
import CollaborationItemCard from './CollaborationItemCard.vue'
import draggable from 'vuedraggable'
import type { DragChangeEvent } from '~/types/dragEvents'

interface Props {
  /** 項目列表（樹狀結構） */
  items: CollaborationItem[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'add-item': [parentId?: string | null]
  'edit-item': [item: CollaborationItem]
  'delete-item': [item: CollaborationItem]
  'reorder': [itemIds: string[], parentId: string | null]
}>()

// 展開狀態追蹤
const expandedItems = ref<Record<string, boolean>>({})

// 本地頂層項目列表（用於拖曳）
const localItems = computed({
  get: () => props.items,
  set: async (newItems) => {
    // 拖曳後更新順序
    const itemIds = newItems.map(item => item.id)
    emit('reorder', itemIds, null)
  }
})

// 拖曳選項 - 禁止跨層級拖曳
const dragOptions = {
  animation: 200,
  group: {
    name: 'items-level-0',
    pull: false,
    put: false
  },
  disabled: false,
  ghostClass: 'drag-ghost',
  chosenClass: 'drag-chosen',
  dragClass: 'drag-dragging',
  handle: '.drag-handle'
}

// 處理拖曳變更
const handleChange = (evt: DragChangeEvent<CollaborationItem>) => {
  if (evt.moved || evt.added) {
    const itemIds = localItems.value.map(item => item.id)
    emit('reorder', itemIds, null)
  }
}

// 處理展開/收起
const handleToggleExpand = (item: CollaborationItem) => {
  expandedItems.value[item.id] = !expandedItems.value[item.id]
}
</script>

<template>
  <div class="collaboration-item-tree space-y-2">
    <draggable
      v-model="localItems"
      v-bind="dragOptions"
      item-key="id"
      @change="handleChange"
    >
      <template #item="{ element }">
        <CollaborationItemCard
          :item="element"
          :level="0"
          :expanded="expandedItems[element.id] || false"
          @add-item="emit('add-item', $event)"
          @edit-item="emit('edit-item', $event)"
          @delete-item="emit('delete-item', $event)"
          @reorder="emit('reorder', $event[0], $event[1])"
          @toggle-expand="handleToggleExpand"
        />
      </template>
    </draggable>
  </div>
</template>

<style scoped>
.collaboration-item-tree {
  min-height: 100px;
}
</style>

