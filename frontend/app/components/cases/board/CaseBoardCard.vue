<script setup lang="ts">
import type { Case } from '~/types/cases'
import { formatAmount, formatRelativeDate, isDeadlineUrgent } from '~/utils/formatters'
import CaseStatusBadge from '~/components/cases/common/CaseStatusBadge.vue'

interface Props {
  /** 案件資料 */
  caseData: Case
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'card-click': [caseId: string]
}>()

const handleClick = () => {
  emit('card-click', props.caseData.id)
}

// 狀態漸層背景類別
const statusGradientClass = computed(() => {
  const gradients = {
    to_confirm: 'from-yellow-50 to-yellow-100 dark:from-yellow-900/20 dark:to-yellow-800/20',
    in_progress: 'from-blue-50 to-blue-100 dark:from-blue-900/20 dark:to-blue-800/20',
    completed: 'from-green-50 to-green-100 dark:from-green-900/20 dark:to-green-800/20',
    cancelled: 'from-gray-50 to-gray-100 dark:from-gray-900/20 dark:to-gray-800/20'
  }
  return gradients[props.caseData.status] || gradients.to_confirm
})

// 取得品牌名稱首字（用於頭像）
const avatarText = computed(() => {
  return props.caseData.brand_name?.charAt(0) || '?'
})
</script>

<template>
  <div
    :class="[
      'case-board-card rounded-lg p-4 transition-all duration-200 cursor-pointer',
      `bg-gradient-to-br ${statusGradientClass}`,
      'hover:shadow-lg hover:-translate-y-1 border border-gray-200 dark:border-gray-700'
    ]"
    @click="handleClick"
  >
    <!-- 標題和狀態 -->
    <div class="flex items-start justify-between gap-2 mb-3">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-white flex-1 line-clamp-2">
        {{ caseData.title }}
      </h3>
      <CaseStatusBadge :status="caseData.status" size="xs" />
    </div>

    <!-- 品牌名稱 -->
    <div class="flex items-center gap-2 mb-2">
      <div
        class="w-6 h-6 rounded-full bg-primary-500 text-white text-xs font-semibold flex items-center justify-center flex-shrink-0"
      >
        {{ avatarText }}
      </div>
      <span class="text-xs text-gray-600 dark:text-gray-400 truncate">
        {{ caseData.brand_name }}
      </span>
    </div>

    <!-- 金額 -->
    <div v-if="caseData.quoted_amount" class="mb-2">
      <div class="text-xs font-semibold text-gray-900 dark:text-white">
        {{ formatAmount(caseData.quoted_amount, caseData.currency) }}
      </div>
    </div>

    <!-- 截止日期 -->
    <div v-if="caseData.deadline_date" class="mb-2 flex items-center gap-1">
      <UIcon
        name="i-lucide-calendar"
        :class="[
          'w-3.5 h-3.5',
          isDeadlineUrgent(caseData.deadline_date)
            ? 'text-red-500'
            : 'text-gray-400 dark:text-gray-500'
        ]"
      />
      <span
        :class="[
          'text-xs',
          isDeadlineUrgent(caseData.deadline_date)
            ? 'text-red-600 dark:text-red-400 font-semibold'
            : 'text-gray-500 dark:text-gray-400'
        ]"
      >
        {{ formatRelativeDate(caseData.deadline_date) }}
      </span>
    </div>

    <!-- 統計資訊 -->
    <div class="flex items-center gap-3 mt-3 pt-3 border-t border-gray-200 dark:border-gray-700">
      <div v-if="caseData.email_count" class="flex items-center gap-1 text-xs text-gray-500 dark:text-gray-400">
        <UIcon name="i-lucide-mail" class="w-3.5 h-3.5" />
        <span>{{ caseData.email_count }}</span>
      </div>
      <div
        v-if="caseData.task_count !== undefined"
        class="flex items-center gap-1 text-xs text-gray-500 dark:text-gray-400"
      >
        <UIcon name="i-lucide-check-square" class="w-3.5 h-3.5" />
        <span>
          {{ caseData.completed_task_count || 0 }}/{{ caseData.task_count }}
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.case-board-card {
  position: relative;
  will-change: transform, box-shadow;
}

/* 拖曳時的視覺效果 */
.case-board-card.sortable-ghost {
  opacity: 0.5;
  transform: rotate(2deg);
}

.case-board-card.sortable-chosen {
  transform: scale(1.02);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.15);
  z-index: 1000;
}

.case-board-card.sortable-drag {
  cursor: grabbing !important;
  z-index: 1000;
}
</style>

