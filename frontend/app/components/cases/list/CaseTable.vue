<script setup lang="ts">
import type { Case } from '~/types/cases'
import type { TableColumn } from '@nuxt/ui'
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

// 解析組件以便在 h() 中使用
const UIcon = resolveComponent('UIcon')
const UButton = resolveComponent('UButton')

const columns: TableColumn<Case>[] = [
  {
    accessorKey: 'title',
    header: '案件標題',
    cell: ({ row }) => {
      return h('div', { class: 'flex flex-col' }, [
        h('p', { class: 'font-medium text-gray-900 dark:text-white' }, row.original.title),
        h('p', { class: 'text-sm text-gray-500 dark:text-gray-400' }, row.original.brand_name)
      ])
    }
  },
  {
    accessorKey: 'status',
    header: '狀態',
    cell: ({ row }) => {
      return h(CaseStatusBadge, { status: row.original.status })
    }
  },
  {
    accessorKey: 'quoted_amount',
    header: '報價金額',
    cell: ({ row }) => {
      return row.original.quoted_amount
        ? formatAmount(row.original.quoted_amount, row.original.currency)
        : '-'
    }
  },
  {
    accessorKey: 'deadline_date',
    header: '截止日期',
    cell: ({ row }) => {
      if (!row.original.deadline_date) return '-'
      
      const isUrgent = isDeadlineUrgent(row.original.deadline_date)
      return h('div', { class: 'flex items-center gap-1' }, [
        h(UIcon as any, {
          name: 'i-lucide-calendar',
          class: isUrgent ? 'w-4 h-4 text-red-500' : 'w-4 h-4 text-gray-400'
        }),
        h('span', {
          class: isUrgent ? 'text-red-600 dark:text-red-400 font-semibold' : 'text-gray-600 dark:text-gray-400'
        }, formatRelativeDate(row.original.deadline_date))
      ])
    }
  },
  {
    accessorKey: 'email_count',
    header: '郵件',
    cell: ({ row }) => row.original.email_count || 0
  },
  {
    accessorKey: 'task_count',
    header: '任務',
    cell: ({ row }) => {
      const completed = row.original.completed_task_count || 0
      const total = row.original.task_count || 0
      return total > 0 ? `${completed}/${total}` : '-'
    }
  },
  {
    accessorKey: 'updated_at',
    header: '更新時間',
    cell: ({ row }) => formatRelativeDate(row.original.updated_at)
  },
  {
    id: 'actions',
    header: '',
    cell: ({ row }) => {
      return h(UButton as any, {
        icon: 'i-lucide-chevron-right',
        variant: 'ghost',
        size: 'sm',
        onClick: (e: MouseEvent) => {
          e.stopPropagation()
          handleRowClick(row.original.id)
        }
      })
    }
  }
]
</script>

<template>
  <div class="case-table" @click="handleTableClick">
    <UTable
      :data="cases"
      :columns="columns"
      :loading="loading"
      :ui="{
        tbody: '[&>tr]:cursor-pointer [&>tr]:transition-colors [&>tr:hover]:bg-gray-50 dark:[&>tr:hover]:bg-gray-800/50',
        td: 'cursor-pointer'
      }"
    />
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
