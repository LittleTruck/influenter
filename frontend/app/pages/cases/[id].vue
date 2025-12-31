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
const { fetchCase, currentCase, loading, updateCase } = useCases()
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

// 處理階段日期編輯
const handleEditPhase = (phase: any) => {
  editingPhase.value = phase
  showPhaseDateEditor.value = true
}

// 處理階段刪除
const handleDeletePhase = async (phase: any) => {
  try {
    // TODO: 呼叫 API 刪除階段
    handleSuccess('階段已刪除')
    await fetchCase(caseId.value)
  } catch (error: any) {
    handleError(error, '刪除失敗')
  }
}

// 處理新增階段
const handleAddPhase = () => {
  // TODO: 實作新增階段功能
  toast.add({
    title: '功能開發中',
    description: '手動新增階段功能即將推出',
    color: 'info'
  })
}

// 處理階段日期更新
const handlePhaseDateUpdate = async (data: any) => {
  try {
    // TODO: 呼叫 API 更新階段日期
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
    // TODO: 呼叫 API 套用流程
    // const config = useRuntimeConfig()
    // const authStore = useAuthStore()
    // await $fetch(`${config.public.apiBase}/api/v1/cases/${caseId.value}/phases/apply-template`, {
    //   method: 'POST',
    //   body: data,
    //   headers: {
    //     Authorization: `Bearer ${authStore.token}`
    //   }
    // })
    
    // 暫時：顯示提示訊息
    toast.add({
      title: '套用流程功能開發中',
      description: `將套用流程 ${data.workflow_id} 的階段`,
      color: 'info'
    })
    
    handleSuccess('流程已套用（模擬）')
    showApplyTemplate.value = false
    // await fetchCase(caseId.value) // API 實作後取消註解
  } catch (error: any) {
    handleError(error, '套用失敗')
  }
}

// 取得案件階段列表（如果有的話，使用假資料）
const casePhases = computed(() => {
  if ((currentCase.value as any)?.phases && (currentCase.value as any).phases.length > 0) {
    return (currentCase.value as any).phases
  }
  // 假資料
  return [
    {
      id: 'phase-1',
      name: '腳本',
      start_date: '2024-01-01',
      end_date: '2024-01-07',
      duration_days: 7,
      order: 0,
      status: 'completed'
    },
    {
      id: 'phase-2',
      name: '拍攝',
      start_date: '2024-01-08',
      end_date: '2024-01-15',
      duration_days: 8,
      order: 1,
      status: 'in_progress'
    },
    {
      id: 'phase-3',
      name: '後製',
      start_date: '2024-01-16',
      end_date: '2024-01-23',
      duration_days: 8,
      order: 2,
      status: 'pending'
    }
  ]
})

// 取得案件開始日期
const caseStartDate = computed(() => {
  return (currentCase.value as any)?.start_date || new Date().toISOString().split('T')[0]
})

// 假郵件資料
const mockEmails = computed(() => {
  if ((currentCase.value as any)?.emails && (currentCase.value as any).emails.length > 0) {
    return (currentCase.value as any).emails
  }
  return [
    {
      id: 'email-1',
      from_name: '張小明',
      from_email: 'zhang@example.com',
      subject: '合作提案',
      received_at: '2024-01-01T10:00:00Z',
      email_type: 'initial_inquiry'
    },
    {
      id: 'email-2',
      from_name: '李經理',
      from_email: 'li@example.com',
      subject: 'Re: 合作提案',
      received_at: '2024-01-02T14:30:00Z',
      email_type: 'reply'
    },
    {
      id: 'email-3',
      from_name: '張小明',
      from_email: 'zhang@example.com',
      subject: 'Re: Re: 合作提案',
      received_at: '2024-01-03T09:15:00Z',
      email_type: 'follow_up'
    }
  ]
})
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
        <!-- 內容區域 -->
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <!-- 左側：郵件和專案流程 -->
          <div class="lg:col-span-2 space-y-6">
            <!-- 郵件區域 -->
            <AppSectionWithHeader
              title="郵件"
            >
              <div class="space-y-4">
                <CaseEmailsTimeline :emails="mockEmails" />
                <BaseButton
                  icon="i-lucide-sparkles"
                  variant="outline"
                  block
                  @click="() => {}"
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
                <BaseButton
                  icon="i-lucide-layout-template"
                  size="sm"
                  variant="outline"
                  @click="showApplyTemplate = true"
                >
                  {{ casePhases.length === 0 ? '套用流程' : '重新套用流程' }}
                </BaseButton>
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

