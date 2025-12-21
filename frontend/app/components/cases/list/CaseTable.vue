<script setup lang="ts">
import type { Case } from '~/types/cases'
import type { TableColumn } from '@nuxt/ui'
import { BaseTable, BaseButton, BaseIcon } from '~/components/base'
import CaseStatusBadge from '~/components/cases/common/CaseStatusBadge.vue'
import { formatAmount, formatRelativeDate, isDeadlineUrgent } from '~/utils/formatters'

interface Props {
  /** 案件列表 */
  cases: Case[]
  /** 是否載入中 */
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false
})

const emit = defineEmits<{
  'case-click': [caseId: string]
}>()

const handleRowClick = (caseId: string) => {
  emit('case-click', caseId)
}

// 處理表格點擊事件（事件委派）
const handleTableClick = (event: MouseEvent) => {
  try {
    const target = event.target as HTMLElement
    
    // 如果點擊的是按鈕、輸入框、連結或其他交互元素，不處理
    if (target.closest('button, input, select, textarea, a, [role="button"]')) {
      return
    }
    
    // 找到被點擊的行（只處理 tbody 中的行，不包括 thead）
    const row = target.closest('tbody tr')
    if (!row) return
    
    // 找到 tbody 中所有行
    const tbody = row.parentElement
    if (!tbody || tbody.tagName !== 'TBODY') return
    
    // 計算行索引（tbody 中的索引，不包括 thead）
    const rows = Array.from(tbody.children)
    const rowIndex = rows.indexOf(row as Element)
    
    // 確保索引有效且 cases 數組有對應的數據
    if (rowIndex >= 0 && rowIndex < props.cases.length && props.cases[rowIndex]?.id) {
      handleRowClick(props.cases[rowIndex].id)
    }
  } catch (error) {
    // 靜默處理錯誤，避免影響用戶體驗
    console.warn('Table click handler error:', error)
  }
}

const columns: TableColumn<Case>[] = [
  {
    accessorKey: 'title',
    header: '案件標題'
  },
  {
    accessorKey: 'status',
    header: '狀態'
  },
  {
    accessorKey: 'quoted_amount',
    header: '報價金額'
  },
  {
    accessorKey: 'deadline_date',
    header: '截止日期'
  },
  {
    accessorKey: 'email_count',
    header: '郵件'
  },
  {
    accessorKey: 'task_count',
    header: '任務'
  },
  {
    accessorKey: 'updated_at',
    header: '更新時間'
  },
  {
    id: 'actions',
    header: ''
  }
]
</script>

<template>
  <div class="case-table" @click="handleTableClick">
    <BaseTable
      :data="cases"
      :columns="columns"
      :loading="loading"
      :ui="{
        tbody: '[&>tr]:cursor-pointer [&>tr]:transition-colors [&>tr:hover]:bg-gray-50 dark:[&>tr:hover]:bg-gray-800/50',
        td: 'cursor-pointer'
      }"
    >
      <!-- 案件標題列 -->
      <template #title-data="{ row }">
        <div class="flex flex-col">
          <p class="font-medium text-gray-900 dark:text-white">{{ row.title }}</p>
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ row.brand_name }}</p>
        </div>
      </template>

      <!-- 狀態列 -->
      <template #status-data="{ row }">
        <CaseStatusBadge :status="row.status" />
      </template>

      <!-- 報價金額列 -->
      <template #quoted_amount-data="{ row }">
        {{ row.quoted_amount ? formatAmount(row.quoted_amount, row.currency) : '-' }}
      </template>

      <!-- 截止日期列 -->
      <template #deadline_date-data="{ row }">
        <div v-if="row.deadline_date" class="flex items-center gap-1">
          <BaseIcon
            name="i-lucide-calendar"
            :class="isDeadlineUrgent(row.deadline_date) ? 'w-4 h-4 text-red-500' : 'w-4 h-4 text-gray-400'"
          />
          <span :class="isDeadlineUrgent(row.deadline_date) ? 'text-red-600 dark:text-red-400 font-semibold' : 'text-gray-600 dark:text-gray-400'">
            {{ formatRelativeDate(row.deadline_date) }}
          </span>
        </div>
        <span v-else>-</span>
      </template>

      <!-- 郵件列 -->
      <template #email_count-data="{ row }">
        {{ row.email_count || 0 }}
      </template>

      <!-- 任務列 -->
      <template #task_count-data="{ row }">
        <span v-if="(row.task_count || 0) > 0">
          {{ row.completed_task_count || 0 }}/{{ row.task_count || 0 }}
        </span>
        <span v-else>-</span>
      </template>

      <!-- 更新時間列 -->
      <template #updated_at-data="{ row }">
        {{ formatRelativeDate(row.updated_at) }}
      </template>

      <!-- 操作列 -->
      <template #actions-data="{ row }">
        <BaseButton
          icon="i-lucide-chevron-right"
          variant="ghost"
          size="sm"
          @click.stop="handleRowClick(row.id)"
        />
      </template>
    </BaseTable>
  </div>
</template>

<style scoped>
.case-table :deep(tbody tr) {
  transition: background-color 0.2s;
}

.case-table :deep(tbody tr:hover) {
  background-color: rgba(0, 0, 0, 0.02);
}

.dark .case-table :deep(tbody tr:hover) {
  background-color: rgba(255, 255, 255, 0.05);
}
</style>
