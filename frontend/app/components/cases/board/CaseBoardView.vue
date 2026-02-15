<script setup lang="ts">
import type { Case, CaseStatus } from '~/types/cases'
import { getStatusLabel, getStatusColorHex } from '~/utils/caseStatus'
import CaseBoardColumn from './CaseBoardColumn.vue'
import LoadingState from '~/components/common/LoadingState.vue'
import ErrorState from '~/components/common/ErrorState.vue'
import EmptyState from '~/components/common/EmptyState.vue'

interface Props {
  /** 案件列表 */
  cases: Case[]
  /** 是否載入中 */
  isLoading?: boolean
  /** 錯誤訊息 */
  error?: string | null
}

const props = withDefaults(defineProps<Props>(), {
  isLoading: false,
  error: null
})

const emit = defineEmits<{
  'case-click': [caseId: string]
  'case-move': [caseId: string, newStatus: CaseStatus]
}>()

// 狀態欄位定義
const statusColumns = computed(() => {
  const statuses: CaseStatus[] = ['to_confirm', 'in_progress', 'completed', 'cancelled', 'other']
  return statuses.map(status => ({
    status,
    label: getStatusLabel(status),
    color: getStatusColorHex(status)
  }))
})

// 根據狀態分組案件
const casesByStatus = computed(() => {
  const grouped: Record<CaseStatus, Case[]> = {
    to_confirm: [],
    in_progress: [],
    completed: [],
    cancelled: [],
    other: []
  }

  props.cases.forEach(c => {
    if (grouped[c.status]) {
      grouped[c.status].push(c)
    }
  })

  return grouped
})

// 處理卡片移動
const handleCaseMove = (caseId: string, newStatus: CaseStatus) => {
  emit('case-move', caseId, newStatus)
}

// 處理卡片點擊
const handleCaseClick = (caseId: string) => {
  emit('case-click', caseId)
}

// 處理欄位案件更新（拖曳後）
const handleColumnUpdate = (status: CaseStatus, updatedCases: Case[]) => {
  // 這裡可以觸發更新，但實際的狀態變更應該由父組件處理
  // 因為拖曳可能跨欄位，需要更新整個列表
}
</script>

<template>
  <!-- Loading State -->
  <LoadingState v-if="isLoading" />

  <!-- Empty State（包含後端尚未就緒或讀取失敗的情況） -->
  <EmptyState
    v-else-if="!cases || cases.length === 0"
    icon="i-lucide-briefcase"
    title="還沒有案件"
    description="開始建立你的第一個案件，讓 AI 幫你管理所有合作邀約"
  />

  <!-- Board View -->
  <div
    v-else
    class="flex h-full gap-4 overflow-x-auto -m-4 sm:-m-6 p-3 sm:p-4"
  >
    <div class="flex gap-4 h-full" style="min-width: fit-content;">
      <CaseBoardColumn
        v-for="column in statusColumns"
        :key="column.status"
        :status="column.status"
        :label="column.label"
        :color="column.color"
        :cases="casesByStatus[column.status]"
        @card-click="handleCaseClick"
        @card-move="handleCaseMove"
        @update:cases="(updatedCases) => handleColumnUpdate(column.status, updatedCases)"
      />
    </div>
  </div>
</template>

<style scoped>
/* 確保看板可以水平滾動 */
:deep(.case-board-view) {
  overflow-x: auto;
  overflow-y: hidden;
}
</style>

