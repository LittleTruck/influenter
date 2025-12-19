<script setup lang="ts">
import type { CaseDetail } from '~/types/cases'
import { useCases } from '~/composables/useCases'
import { useCaseFields } from '~/composables/useCaseFields'
import { useErrorHandler } from '~/composables/useErrorHandler'
import CasePropertiesPanel from '~/components/cases/detail/CasePropertiesPanel.vue'
import CaseTasksList from '~/components/cases/detail/CaseTasksList.vue'
import CaseEmailsTimeline from '~/components/cases/detail/CaseEmailsTimeline.vue'
import CaseCollaborationItems from '~/components/cases/detail/CaseCollaborationItems.vue'
import LoadingState from '~/components/common/LoadingState.vue'
import ErrorState from '~/components/common/ErrorState.vue'

definePageMeta({
  middleware: 'auth'
})

const route = useRoute()
const { fetchCase, currentCase, loading, updateCase } = useCases()
const { allFields, fetchFields } = useCaseFields()
const { handleError, handleSuccess } = useErrorHandler()

const caseId = computed(() => route.params.id as string)

// 載入案件詳情和屬性
onMounted(async () => {
  await Promise.all([
    fetchCase(caseId.value),
    fetchFields()
  ])
})

// 處理屬性更新
const handleFieldUpdate = async (fieldName: string, value: unknown): Promise<void> => {
  try {
    await updateCase(caseId.value, { [fieldName]: value })
    handleSuccess('屬性已更新')
  } catch (error: any) {
    handleError(error, '更新失敗')
  }
}

// 處理屬性刪除
const handleFieldDelete = async (fieldId: string) => {
  try {
    const { deleteField } = useCaseFields()
    await deleteField(fieldId)
    await fetchFields() // 重新載入屬性列表
    await fetchCase(caseId.value) // 重新載入案件以更新屬性值
    handleSuccess('屬性已刪除')
  } catch (error: any) {
    handleError(error, '刪除失敗')
  }
}

// 處理任務更新
const handleTaskUpdate = () => {
  fetchCase(caseId.value) // 重新載入案件以更新任務列表
}
</script>

<template>
  <UDashboardPanel>
    <template #header>
      <UDashboardNavbar :title="currentCase?.title || '案件詳情'">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>

        <template #trailing>
          <UButton
            icon="i-lucide-arrow-left"
            variant="ghost"
            @click="navigateTo('/cases')"
          >
            返回列表
          </UButton>
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <!-- 載入中 -->
      <LoadingState v-if="loading" />

      <!-- 案件詳情 -->
      <div v-else-if="currentCase" class="space-y-6">
        <!-- 品牌名稱 -->
        <div v-if="currentCase.brand_name" class="text-sm text-gray-500 dark:text-gray-400">
          {{ currentCase.brand_name }}
        </div>

        <!-- 內容區域 -->
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <!-- 左側：屬性面板和合作項目 -->
          <div class="lg:col-span-2 space-y-6">
            <UCard>
              <template #header>
                <h2 class="text-lg font-semibold">屬性</h2>
              </template>
              <CasePropertiesPanel
                :case="currentCase"
                :fields="allFields"
                :editable="true"
                @field-update="handleFieldUpdate"
                @field-delete="handleFieldDelete"
              />
            </UCard>

            <!-- 合作項目 -->
            <UCard>
              <template #header>
                <h2 class="text-lg font-semibold">合作項目</h2>
              </template>
              <CaseCollaborationItems
                :case="currentCase"
                :editable="true"
                @update="fetchCase(caseId)"
              />
            </UCard>
          </div>

          <!-- 右側：任務和郵件 -->
          <div class="space-y-6">
            <!-- 任務列表 -->
            <UCard>
              <template #header>
                <h2 class="text-lg font-semibold">任務</h2>
              </template>
              <CaseTasksList
                :case-id="caseId"
                :tasks="currentCase.tasks || []"
                @task-update="handleTaskUpdate"
              />
            </UCard>

            <!-- 郵件時間軸 -->
            <UCard>
              <template #header>
                <h2 class="text-lg font-semibold">郵件</h2>
              </template>
              <CaseEmailsTimeline :emails="currentCase.emails || []" />
            </UCard>
          </div>
        </div>
      </div>

      <!-- 錯誤狀態 -->
      <ErrorState v-else title="無法載入案件詳情" message="請重新整理頁面或返回列表" />
    </template>
  </UDashboardPanel>
</template>

