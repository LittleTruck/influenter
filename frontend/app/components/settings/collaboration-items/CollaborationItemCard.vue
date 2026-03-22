<script setup lang="ts">
import type { CollaborationItem } from '~/types/collaborationItems'
import { useWorkflowTemplates } from '~/composables/useWorkflowTemplates'
import { formatAmount } from '~/utils/formatters'
import { BaseButton, BaseIcon, BaseBadge } from '~/components/base'

interface Props {
  /** 項目資料 */
  item: CollaborationItem
}

const props = defineProps<Props>()

const { findWorkflowById } = useWorkflowTemplates()

const emit = defineEmits<{
  'edit-item': [item: CollaborationItem]
  'delete-item': [item: CollaborationItem]
}>()

// 取得流程名稱（僅 individual 有流程）
const workflowName = computed(() => {
  if (props.item.type !== 'individual') return null
  if (props.item.workflow) return props.item.workflow.name
  if (props.item.workflow_id) {
    const wf = findWorkflowById(props.item.workflow_id)
    return wf?.name || null
  }
  return null
})

// 取得 bundle 內含的項目名稱
const bundleItemNames = computed(() => {
  if (props.item.type !== 'bundle' || !props.item.bundle_items) return []
  return props.item.bundle_items
    .sort((a, b) => a.order - b.order)
    .map(ref => ref.item?.title || '未知項目')
})
</script>

<template>
  <div
    class="flex items-center gap-3 p-3 border border-default rounded-lg hover:bg-white dark:hover:bg-gray-700/50 transition-colors bg-elevated"
  >
    <!-- 拖曳手柄 -->
    <BaseIcon
      name="i-lucide-grip-vertical"
      class="w-5 h-5 text-dimmed drag-handle cursor-grab flex-shrink-0"
      @click.stop
    />

    <!-- 項目資訊 -->
    <div class="flex-1 min-w-0">
      <div class="flex items-center gap-2">
        <h4 class="font-medium text-highlighted truncate">
          {{ item.title }}
        </h4>
        <BaseBadge color="primary" variant="subtle" size="xs">
          {{ formatAmount(item.price) }}
        </BaseBadge>
        <!-- 類型標籤 -->
        <BaseBadge
          v-if="item.type === 'bundle'"
          color="info"
          variant="subtle"
          size="xs"
        >
          組合
        </BaseBadge>
        <!-- 流程標籤（僅 individual） -->
        <BaseBadge
          v-if="workflowName"
          color="neutral"
          variant="subtle"
          size="xs"
        >
          {{ workflowName }}
        </BaseBadge>
      </div>
      <!-- Bundle 內含項目 -->
      <div v-if="item.type === 'bundle' && bundleItemNames.length > 0" class="flex flex-wrap gap-1 mt-1">
        <span
          v-for="(name, idx) in bundleItemNames"
          :key="idx"
          class="text-xs px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 text-muted rounded"
        >
          {{ name }}
        </span>
      </div>
    </div>

    <!-- 操作按鈕 -->
    <div class="flex items-center gap-1 flex-shrink-0 ml-2" @click.stop>
      <BaseButton
        icon="i-lucide-edit"
        variant="ghost"
        size="xs"
        @click="emit('edit-item', item)"
      />
      <BaseButton
        icon="i-lucide-trash-2"
        variant="ghost"
        size="xs"
        color="error"
        @click="emit('delete-item', item)"
      />
    </div>
  </div>
</template>

<style scoped>
/* 拖曳樣式已統一在 DraggableList 組件中 */
</style>
