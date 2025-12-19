import type { Case, CaseDetail, CreateCaseRequest, UpdateCaseRequest } from '~/types/cases'
import { formatAmount, formatRelativeDate, formatFullDate, isDeadlineUrgent } from '~/utils/formatters'
import { getStatusColor, getStatusLabel } from '~/utils/caseStatus'

/**
 * 案件相關 composable
 * 提供便捷的案件操作方法
 * 
 * 職責：
 * - 封裝案件相關的 store 操作
 * - 提供業務邏輯方法
 * - 不包含格式化工具函數（已移至 utils）
 */
export const useCases = () => {
  // 確保在正確的上下文中使用 store
  let casesStore: ReturnType<typeof useCasesStore>
  try {
    casesStore = useCasesStore()
  } catch (error) {
    console.error('Failed to initialize cases store:', error)
    throw error
  }

  return {
    // Store 方法
    fetchCases: casesStore.fetchCases,
    fetchCase: casesStore.fetchCase,
    createCase: casesStore.createCase,
    updateCase: casesStore.updateCase,
    updateCaseStatus: casesStore.updateCaseStatus,
    deleteCase: casesStore.deleteCase,
    linkEmailToCase: casesStore.linkEmailToCase,
    unlinkEmailFromCase: casesStore.unlinkEmailFromCase,
    fetchCaseEmails: casesStore.fetchCaseEmails,
    fetchCaseTasks: casesStore.fetchCaseTasks,
    createTask: casesStore.createTask,
    updateTask: casesStore.updateTask,
    deleteTask: casesStore.deleteTask,
    completeTask: casesStore.completeTask,
    reorderTasks: casesStore.reorderTasks,

    // 格式化函數（從 utils 重新匯出，保持向後兼容）
    formatAmount,
    formatRelativeDate,
    formatFullDate,
    isDeadlineUrgent,
    getStatusColor,
    getStatusLabel,

    // Computed 值
    cases: computed(() => casesStore.cases),
    currentCase: computed(() => casesStore.currentCase),
    loading: computed(() => casesStore.loading),
    error: computed(() => casesStore.error),
    pagination: computed(() => casesStore.pagination),
    filters: computed(() => casesStore.filters),
    currentView: computed(() => casesStore.currentView),
    hasCases: computed(() => casesStore.hasCases),
    casesByStatus: computed(() => casesStore.casesByStatus)
  }
}

