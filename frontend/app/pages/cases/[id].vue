<script setup lang="ts">
import type { CaseDetail, ApplyTemplateRequest } from '~/types/cases'
import { useCases } from '~/composables/useCases'
import { useCaseFields } from '~/composables/useCaseFields'
import { useErrorHandler } from '~/composables/useErrorHandler'
import { BaseDashboardPanel, BaseDashboardNavbar, BaseDashboardSidebarCollapse, BaseButton, BaseCard } from '~/components/base'
import AppSectionWithHeader from '~/components/ui/AppSectionWithHeader.vue'
import CasePropertiesPanel from '~/components/cases/detail/CasePropertiesPanel.vue'
import CaseEmailsTimeline from '~/components/cases/detail/CaseEmailsTimeline.vue'
import CaseCollaborationItems from '~/components/cases/detail/CaseCollaborationItems.vue'
import CasePhaseTimeline from '~/components/cases/detail/CasePhaseTimeline.vue'
import PhaseDateEditor from '~/components/cases/detail/PhaseDateEditor.vue'
import ApplyTemplateModal from '~/components/cases/detail/ApplyTemplateModal.vue'
import LoadingState from '~/components/common/LoadingState.vue'
import ErrorState from '~/components/common/ErrorState.vue'

definePageMeta({
  middleware: 'auth'
})

const route = useRoute()
const router = useRouter()
const { fetchCase, fetchCaseEmails, currentCase, loading, updateCase } = useCases()
const { allFields, fetchFields } = useCaseFields()

const caseId = computed(() => route.params.id as string)

// 載入案件詳情和屬性（忽略 404 錯誤，因為後端還沒實作）
onMounted(async () => {
  try {
    await Promise.all([
      fetchCase(caseId.value).catch((err: any) => {
        if (err?.statusCode !== 404) {
          console.error('載入案件失敗:', err)
        }
      }),
      fetchFields().catch((err: any) => {
        if (err?.statusCode !== 404) {
          console.error('載入案件屬性失敗:', err)
        }
      })
    ])
    // 合作案件：載入關聯郵件
    const caseData = currentCase.value
    if (caseData && caseData.status !== 'other') {
      await fetchCaseEmails(caseId.value).catch((err: any) => {
        if (err?.statusCode !== 404) {
          console.error('載入案件郵件失敗:', err)
        }
      })
    }
  } catch (err) {
    // 忽略錯誤，讓頁面正常顯示
    console.error('載入失敗:', err)
  }
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

// 階段管理相關
const showPhaseDateEditor = ref(false)
const editingPhase = ref<any>(null)
const showApplyTemplate = ref(false)

// 點擊 AI 擬信：導向該案件最新一封郵件的回覆頁
const goToReplyPage = () => {
  const list = currentCase.value?.emails ?? []
  if (list.length === 0) return
  const lastEmailId = list[list.length - 1].id
  router.push(`/emails/${lastEmailId}?reply=1&from_case=${caseId.value}`)
}

// 處理階段日期編輯
const handleEditPhase = (phase: any) => {
  editingPhase.value = phase
  showPhaseDateEditor.value = true
}

// API helpers
const config = useRuntimeConfig()
const authStore = useAuthStore()
const apiHeaders = computed(() => ({
  Authorization: `Bearer ${authStore.token}`
}))

// 處理階段刪除
const handleDeletePhase = async (phase: any) => {
  try {
    await $fetch(
      `${config.public.apiBase}/api/v1/cases/${caseId.value}/phases/${phase.id}`,
      {
        method: 'DELETE',
        headers: apiHeaders.value
      }
    )
    handleSuccess('階段已刪除')
    await fetchCase(caseId.value)
  } catch (error: any) {
    handleError(error, '刪除失敗')
  }
}

// 處理新增階段
const handleAddPhase = async () => {
  try {
    await $fetch(
      `${config.public.apiBase}/api/v1/cases/${caseId.value}/phases`,
      {
        method: 'POST',
        body: {
          name: '新階段',
          start_date: new Date().toISOString().split('T')[0],
          duration_days: 7
        },
        headers: apiHeaders.value
      }
    )
    handleSuccess('階段已新增')
    await fetchCase(caseId.value)
  } catch (error: any) {
    handleError(error, '新增失敗')
  }
}

// 處理階段日期更新
const handlePhaseDateUpdate = async (data: any) => {
  try {
    const phaseId = editingPhase.value?.id
    if (!phaseId) return

    await $fetch(
      `${config.public.apiBase}/api/v1/cases/${caseId.value}/phases/${phaseId}`,
      {
        method: 'PATCH',
        body: data,
        headers: apiHeaders.value
      }
    )
    handleSuccess('階段日期已更新')
    showPhaseDateEditor.value = false
    editingPhase.value = null
    await fetchCase(caseId.value)
  } catch (error: any) {
    handleError(error, '更新失敗')
  }
}

// 處理套用流程
const handleApplyTemplate = async (data: ApplyTemplateRequest) => {
  try {
    await $fetch(
      `${config.public.apiBase}/api/v1/cases/${caseId.value}/phases/apply-template`,
      {
        method: 'POST',
        body: data,
        headers: apiHeaders.value
      }
    )
    handleSuccess('流程已套用')
    showApplyTemplate.value = false
    await fetchCase(caseId.value)
  } catch (error: any) {
    handleError(error, '套用失敗')
  }
}

// AI 自動套用流程
const autoApplying = ref(false)
const handleAutoApplyTemplate = async () => {
  autoApplying.value = true
  try {
    const result = await $fetch<{ matched: boolean; message: string; reason: string; template_name?: string }>(
      `${config.public.apiBase}/api/v1/cases/${caseId.value}/phases/auto-apply`,
      {
        method: 'POST',
        headers: apiHeaders.value
      }
    )
    if (result.matched) {
      handleSuccess(result.message || 'AI 已自動套用流程')
      await fetchCase(caseId.value)
    } else {
      toast.add({
        title: 'AI 無法自動套用',
        description: result.reason || '找不到適合的流程範本',
        color: 'warning'
      })
    }
  } catch (error: any) {
    handleError(error, 'AI 自動套用失敗')
  } finally {
    autoApplying.value = false
  }
}

// 取得案件階段列表（從 API 回傳的 phases）
const casePhases = computed(() => {
  return (currentCase.value as any)?.phases ?? []
})

// 取得案件開始日期
const caseStartDate = computed(() => {
  return (currentCase.value as any)?.start_date || new Date().toISOString().split('T')[0]
})

// 案件關聯郵件（由 fetchCaseEmails 載入，無則為空陣列）
const caseEmails = computed(() => currentCase.value?.emails ?? [])
</script>

<template>
  <BaseDashboardPanel>
    <template #header>
      <BaseDashboardNavbar :title="currentCase?.title || '案件詳情'">
        <template #leading>
          <BaseDashboardSidebarCollapse />
        </template>

        <template #trailing>
          <BaseButton
            icon="i-lucide-arrow-left"
            variant="ghost"
            @click="navigateTo('/cases')"
          >
            返回列表
          </BaseButton>
        </template>
      </BaseDashboardNavbar>
    </template>

    <template #body>
      <!-- 載入中 -->
      <LoadingState v-if="loading" />

      <!-- 案件詳情 -->
      <div v-else-if="currentCase" class="space-y-6">
        <!-- 非合作案件：僅顯示屬性面板 -->
        <div v-if="currentCase.status === 'other'" class="max-w-2xl">
          <CasePropertiesPanel
            :case="currentCase"
            :fields="allFields"
            :editable="true"
            @field-update="handleFieldUpdate"
            @field-delete="handleFieldDelete"
          />
        </div>

        <!-- 合作案件：完整版面（郵件、專案流程、屬性、合作項目） -->
        <div v-else class="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <!-- 左側：郵件和專案流程 -->
          <div class="lg:col-span-2 space-y-6">
            <!-- 郵件區域 -->
            <AppSectionWithHeader
              title="郵件"
            >
              <div class="space-y-4">
                <CaseEmailsTimeline :emails="caseEmails" :case-id="caseId" />
                <BaseButton
                  icon="i-lucide-sparkles"
                  variant="outline"
                  block
                  :disabled="caseEmails.length === 0"
                  @click="goToReplyPage"
                >
                  AI 擬信
                </BaseButton>
              </div>
            </AppSectionWithHeader>

            <!-- 專案流程 -->
            <AppSectionWithHeader
              title="專案流程"
              description="管理案件的執行階段"
            >
              <template #actions>
                <div class="flex gap-2">
                  <BaseButton
                    icon="i-lucide-sparkles"
                    size="sm"
                    variant="outline"
                    :loading="autoApplying"
                    @click="handleAutoApplyTemplate"
                  >
                    AI 自動套用
                  </BaseButton>
                  <BaseButton
                    icon="i-lucide-layout-template"
                    size="sm"
                    variant="outline"
                    @click="showApplyTemplate = true"
                  >
                    {{ casePhases.length === 0 ? '手動套用' : '重新套用' }}
                  </BaseButton>
                </div>
              </template>
              <CasePhaseTimeline
                :phases="casePhases"
                :editable="true"
                @edit-phase="handleEditPhase"
                @delete-phase="handleDeletePhase"
                @add-phase="handleAddPhase"
              />
            </AppSectionWithHeader>
          </div>

          <!-- 右側：屬性面板和合作項目 -->
          <div class="space-y-6">
            <!-- 屬性面板 -->
            <CasePropertiesPanel
              :case="currentCase"
              :fields="allFields"
              :editable="true"
              @field-update="handleFieldUpdate"
              @field-delete="handleFieldDelete"
            />

            <!-- 合作項目 -->
            <CaseCollaborationItems
              :case="currentCase"
              :editable="true"
              @update="fetchCase(caseId)"
            />
          </div>
        </div>
      </div>

      <!-- 錯誤狀態 -->
      <ErrorState v-else title="無法載入案件詳情" message="請重新整理頁面或返回列表" />

      <!-- 階段日期編輯 Modal -->
      <PhaseDateEditor
        v-model="showPhaseDateEditor"
        :phase="editingPhase"
        @submit="handlePhaseDateUpdate"
      />

      <!-- 套用流程 Modal -->
      <ApplyTemplateModal
        v-model="showApplyTemplate"
        :case-start-date="caseStartDate"
        :case-id="caseId"
        @submit="handleApplyTemplate"
      />
    </template>
  </BaseDashboardPanel>
</template>

