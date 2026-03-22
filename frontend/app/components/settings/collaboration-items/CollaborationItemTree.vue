<script setup lang="ts">
import type { CollaborationItem } from '~/types/collaborationItems'
import CollaborationItemCard from './CollaborationItemCard.vue'
import DraggableList from '~/components/base/DraggableList.vue'

interface Props {
  /** 單項列表 */
  individualItems: CollaborationItem[]
  /** 組合列表 */
  bundleItems: CollaborationItem[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'add-item': [type: 'individual' | 'bundle']
  'edit-item': [item: CollaborationItem]
  'delete-item': [item: CollaborationItem]
  'reorder': [itemIds: string[]]
}>()

// 本地列表（用於拖曳）
const localIndividualItems = ref([...props.individualItems])
const localBundleItems = ref([...props.bundleItems])

// 同步外部 items 變化
watch(() => props.individualItems, (newItems) => {
  localIndividualItems.value = [...newItems]
}, { deep: true, immediate: true })

watch(() => props.bundleItems, (newItems) => {
  localBundleItems.value = [...newItems]
}, { deep: true, immediate: true })

// 處理拖曳重排序（合併兩個列表的順序）
const handleReorderIndividual = (itemIds: string[]) => {
  const bundleIds = localBundleItems.value.map(i => i.id)
  emit('reorder', [...itemIds, ...bundleIds])
}

const handleReorderBundle = (itemIds: string[]) => {
  const individualIds = localIndividualItems.value.map(i => i.id)
  emit('reorder', [...individualIds, ...itemIds])
}
</script>

<template>
  <div class="collaboration-item-tree space-y-6">
    <!-- 單項區塊 -->
    <div>
      <div class="flex items-center justify-between mb-2">
        <h3 class="text-sm font-semibold text-highlighted">單項</h3>
      </div>
      <div v-if="localIndividualItems.length === 0" class="text-sm text-muted p-4 text-center border border-dashed border-default rounded-lg">
        尚未建立單項
      </div>
      <DraggableList
        v-else
        v-model:items="localIndividualItems"
        group-name="individual-items"
        @reorder="handleReorderIndividual"
      >
        <template #item="{ element }">
          <CollaborationItemCard
            :item="element"
            @edit-item="emit('edit-item', $event)"
            @delete-item="emit('delete-item', $event)"
          />
        </template>
      </DraggableList>
    </div>

    <!-- 組合區塊 -->
    <div>
      <div class="flex items-center justify-between mb-2">
        <h3 class="text-sm font-semibold text-highlighted">組合</h3>
      </div>
      <div v-if="localBundleItems.length === 0" class="text-sm text-muted p-4 text-center border border-dashed border-default rounded-lg">
        尚未建立組合
      </div>
      <DraggableList
        v-else
        v-model:items="localBundleItems"
        group-name="bundle-items"
        @reorder="handleReorderBundle"
      >
        <template #item="{ element }">
          <CollaborationItemCard
            :item="element"
            @edit-item="emit('edit-item', $event)"
            @delete-item="emit('delete-item', $event)"
          />
        </template>
      </DraggableList>
    </div>
  </div>
</template>

<style scoped>
.collaboration-item-tree {
  min-height: 100px;
}
</style>
