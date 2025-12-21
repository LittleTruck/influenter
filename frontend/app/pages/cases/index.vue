<script setup lang="ts">
import type { Case, ViewType } from '~/types/cases'
import { useCases } from '~/composables/useCases'
import { useErrorHandler } from '~/composables/useErrorHandler'
import { BaseDashboardPanel, BaseDashboardNavbar, BaseDashboardSidebarCollapse, BaseButton } from '~/components/base'
import CaseBoardView from '~/components/cases/board/CaseBoardView.vue'
import CaseTable from '~/components/cases/list/CaseTable.vue'
import CaseFormModal from '~/components/cases/forms/CaseFormModal.vue'
import ViewToggle from '~/components/cases/common/ViewToggle.vue'
import ErrorState from '~/components/common/ErrorState.vue'
import EmptyState from '~/components/common/EmptyState.vue'

definePageMeta({
  middleware: 'auth'
})

const casesStore = useCasesStore()
const { cases, loading, error, fetchCases, updateCaseStatus } = useCases()
const { handleError, handleSuccess } = useErrorHandler()

// 直接使用 store 的 currentView，確保響應性
const currentView = computed(() => casesStore.currentView)

// 表單狀態
const showCaseForm = ref(false)
const editingCase = ref<Case | null>(null)

// 載入案件列表
onMounted(async () => {
  await fetchCases()
})

// 處理視圖切換
const handleViewChange = (newView: ViewType) => {
  casesStore.currentView = newView
}

// 處理案件點擊
const handleCaseClick = (caseId: string) => {
  navigateTo(`/cases/${caseId}`)
}

// 處理案件移動（拖曳）
const handleCaseMove = async (caseId: string, newStatus: Case['status']) => {
  try {
    await updateCaseStatus(caseId, newStatus)
    handleSuccess('案件狀態已更新')
  } catch (error: any) {
    handleError(error, '更新狀態失敗')
  }
}

// 處理新增案件
const handleAddCase = () => {
  editingCase.value = null
  showCaseForm.value = true
}

// 處理表單提交
const handleFormSubmit = () => {
  fetchCases()
}
</script>

<template>
  <BaseDashboardPanel>
    <template #header>
      <BaseDashboardNavbar title="案件管理">
        <template #leading>
          <BaseDashboardSidebarCollapse />
        </template>

        <template #trailing>
          <div class="flex items-center gap-3">
            <ViewToggle :model-value="currentView" @update:model-value="handleViewChange" />
            <BaseButton
              icon="i-lucide-plus"
              @click="handleAddCase"
            >
              建立案件
            </BaseButton>
          </div>
        </template>
      </BaseDashboardNavbar>
    </template>

    <template #body>
    <!-- Board 視圖 -->
    <CaseBoardView
      v-if="currentView === 'board'"
      :cases="cases"
      :is-loading="loading"
      :error="error"
      @case-click="handleCaseClick"
      @case-move="handleCaseMove"
    />

    <!-- List 視圖 -->
    <template v-else>
      <EmptyState
        v-if="cases.length === 0 && !loading"
        icon="i-lucide-briefcase"
        title="還沒有案件"
        action-label="建立第一個案件"
        :show-icon-background="false"
        @action="handleAddCase"
      />
      <CaseTable
        v-else
        :cases="cases"
        :loading="loading"
        @case-click="handleCaseClick"
      />
    </template>

      <!-- 案件表單 Modal -->
      <CaseFormModal
        v-model="showCaseForm"
        :case="editingCase"
        @submit="handleFormSubmit"
      />
    </template>
  </BaseDashboardPanel>
</template>

