<script setup lang="ts">
import type { CollaborationItem } from '~/types/collaborationItems'
import CollaborationItemCard from './CollaborationItemCard.vue'
import DraggableList from '~/components/base/DraggableList.vue'

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
const localItems = ref([...props.items])

// 同步外部 items 變化
watch(() => props.items, (newItems) => {
  localItems.value = [...newItems]
}, { deep: true, immediate: true })

// 處理拖曳重排序
const handleReorder = (itemIds: string[]) => {
  emit('reorder', itemIds, null)
}

// 處理展開/收起
const handleToggleExpand = (item: CollaborationItem) => {
  expandedItems.value[item.id] = !expandedItems.value[item.id]
}
</script>

<template>
  <div class="collaboration-item-tree">
    <DraggableList
      v-model:items="localItems"
      group-name="collaboration-items"
      @reorder="handleReorder"
    >
      <template #item="{ element }">
        <CollaborationItemCard
          :item="element"
          :level="0"
          :expanded="expandedItems[element.id] || false"
          :parent-item="null"
          @add-item="emit('add-item', $event)"
          @edit-item="emit('edit-item', $event)"
          @delete-item="emit('delete-item', $event)"
          @reorder="emit('reorder', $event[0], $event[1])"
          @toggle-expand="handleToggleExpand"
        />
      </template>
    </DraggableList>
  </div>
</template>

<style scoped>
.collaboration-item-tree {
  min-height: 100px;
}
</style>

