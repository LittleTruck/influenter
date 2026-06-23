<script setup lang="ts">
import type { CollaborationItem } from '~/types/collaborationItems'
import { useWorkflowTemplates } from '~/composables/useWorkflowTemplates'
import { formatAmount } from '~/utils/formatters'
import { BaseIcon } from '~/components/base'

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

const isBundle = computed(() => props.item.type === 'bundle')

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
    class="spec-card"
    :class="isBundle ? 'flex-col items-start' : 'items-center'"
  >
    <!-- 主列：拖曳把手 / 名稱 / 類型 Badge / 報價 / 操作 -->
    <div class="flex items-center gap-2.5 w-full">
      <!-- 拖曳把手 -->
      <BaseIcon
        name="i-lucide-grip-vertical"
        class="drag-handle shrink-0 cursor-grab text-[#D3D1C7] dark:text-gray-600 size-4"
        @click.stop
      />

      <!-- 項目名稱 -->
      <h4 class="flex-1 min-w-0 truncate text-[12px] font-medium text-[#26215C] dark:text-gray-100">
        {{ item.title }}
      </h4>

      <!-- 類型 Badge -->
      <span
        v-if="isBundle"
        class="shrink-0 rounded-lg px-2 py-0.5 text-[10px] font-medium bg-[#E6F1FB] text-[#185FA5] dark:bg-[#185FA5]/20 dark:text-[#8FC1EA]"
      >
        組合
      </span>
      <span
        v-else
        class="shrink-0 rounded-lg px-2 py-0.5 text-[10px] font-medium bg-iris-100 text-iris-600 dark:bg-iris-500/15 dark:text-iris-300"
      >
        單項
      </span>

      <!-- 流程標籤（僅 individual） -->
      <span
        v-if="workflowName"
        class="shrink-0 max-w-[120px] truncate rounded-md px-2 py-0.5 text-[10px] bg-[#F4F3FD] text-[#5F5E5A] dark:bg-gray-800 dark:text-gray-400"
      >
        {{ workflowName }}
      </span>

      <!-- 報價金額 -->
      <span class="shrink-0 mr-2 text-[12px] font-medium text-[#534AB7] dark:text-iris-300">
        {{ formatAmount(item.price) }}
      </span>

      <!-- 操作按鈕 -->
      <div class="flex items-center gap-1.5 shrink-0" @click.stop>
        <button
          type="button"
          class="icon-btn bg-[#F4F3FD] text-[#7F77DD] hover:bg-iris-100 dark:bg-iris-500/15 dark:text-iris-300"
          aria-label="編輯"
          @click="emit('edit-item', item)"
        >
          <BaseIcon name="i-lucide-pencil" class="size-3.5" />
        </button>
        <button
          type="button"
          class="icon-btn bg-[#FCEBEB] text-[#E24B4A] hover:bg-[#F8DCDC] dark:bg-[#E24B4A]/15 dark:text-[#F08886]"
          aria-label="刪除"
          @click="emit('delete-item', item)"
        >
          <BaseIcon name="i-lucide-trash-2" class="size-3.5" />
        </button>
      </div>
    </div>

    <!-- Bundle 內含項目標籤列 -->
    <div
      v-if="isBundle && bundleItemNames.length > 0"
      class="flex flex-wrap gap-1 mt-1.5 pl-[22px]"
    >
      <span
        v-for="(name, idx) in bundleItemNames"
        :key="idx"
        class="rounded-md px-2 py-0.5 text-[10px] bg-[#F4F3FD] text-[#5F5E5A] dark:bg-gray-800 dark:text-gray-400"
      >
        {{ name }}
      </span>
    </div>

    <!-- 注意事項預覽 -->
    <div
      v-if="item.notes"
      class="flex items-start gap-1 mt-1.5 text-[11px] text-[#888780] dark:text-gray-400"
      :class="isBundle ? 'pl-[22px]' : ''"
    >
      <BaseIcon name="i-lucide-info" class="size-3.5 shrink-0 mt-0.5" />
      <p class="whitespace-pre-line line-clamp-2">{{ item.notes }}</p>
    </div>
  </div>
</template>

<style scoped>
/* 列表卡片 — 依 redesign.md：無框線、以品牌紫系柔和陰影呈現 */
.spec-card {
  display: flex;
  background: #ffffff;
  border-radius: 10px;
  padding: 11px 14px;
  box-shadow:
    0 1px 4px rgba(83, 74, 183, 0.08),
    0 0 0 0.5px rgba(83, 74, 183, 0.1);
  transition: box-shadow 0.15s ease, transform 0.15s ease;
}

.spec-card:hover {
  box-shadow:
    0 4px 12px rgba(83, 74, 183, 0.14),
    0 0 0 0.5px rgba(83, 74, 183, 0.18);
}

:global(.dark) .spec-card {
  background: #1a1730;
  box-shadow:
    0 1px 4px rgba(0, 0, 0, 0.3),
    0 0 0 0.5px rgba(127, 119, 221, 0.18);
}

:global(.dark) .spec-card:hover {
  box-shadow:
    0 4px 12px rgba(0, 0, 0, 0.4),
    0 0 0 0.5px rgba(127, 119, 221, 0.3);
}

/* 操作按鈕 — 26×26，圓角 6px */
.icon-btn {
  width: 26px;
  height: 26px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background-color 0.15s ease;
}
</style>
