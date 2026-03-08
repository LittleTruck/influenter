<script setup lang="ts">
import type { ApplyTemplateRequest } from '~/types/cases'
import { useCases } from '~/composables/useCases'
import { useCaseFields } from '~/composables/useCaseFields'
import { useErrorHandler } from '~/composables/useErrorHandler'
import { formatAmount } from '~/utils/formatters'
import { getStatusColor, getStatusLabel } from '~/utils/caseStatus'
import { BaseDashboardPanel, BaseDashboardNavbar, BaseDashboardSidebarCollapse, BaseButton, BaseCard, BaseBadge, BaseIcon, BaseInput } from '~/components/base'
import AppSectionWithHeader from '~/components/ui/AppSectionWithHeader.vue'
import CasePropertiesPanel from '~/components/cases/detail/CasePropertiesPanel.vue'
import CaseEmailsTimeline from '~/components/cases/detail/CaseEmailsTimeline.vue'
import CasePhaseTimeline from '~/components/cases/detail/CasePhaseTimeline.vue'
import CasePhaseStepper from '~/components/cases/detail/CasePhaseStepper.vue'
import DraftReplySlideover from '~/components/cases/detail/DraftReplySlideover.vue'
import EmailDetailSlideover from '~/components/cases/detail/EmailDetailSlideover.vue'
import PhaseDateEditor from '~/components/cases/detail/PhaseDateEditor.vue'
import ApplyTemplateModal from '~/components/cases/detail/ApplyTemplateModal.vue'
import LoadingState from '~/components/common/LoadingState.vue'
import ErrorState from '~/components/common/ErrorState.vue'
import { format, differenceInDays } from 'date-fns'

definePageMeta({
  middleware: 'auth'
})

const route = useRoute()
const { fetchCase, fetchCaseEmails, currentCase, loading, updateCase } = useCases()
const { allFields, fetchFields } = useCaseFields()

const { handleError, handleSuccess } = useErrorHandler()
const toast = useToast()

const caseId = computed(() => route.params.id as string)

// 頁面載入守衛：避免在 onMounted 前渲染殘留的舊資料
const pageReady = ref(false)

// 載入案件詳情和屬性
onMounted(async () => {
  try {
    await Promise.all([
      fetchCase(caseId.value).catch(() => {}),
      fetchFields().catch(() => {})
    ])
    const caseData = currentCase.value
    if (caseData && caseData.status !== 'other') {
      await fetchCaseEmails(caseId.value).catch(() => {})
    }
  } finally {
    pageReady.value = true
  }
})

// ── 屬性 inline 編輯 ──
const isEditingProperties = ref(false)

const handleFieldUpdate = async (fieldName: string, value: unknown): Promise<void> => {
  try {
    await updateCase(caseId.value, { [fieldName]: value })
    handleSuccess('已更新')
    await fetchCase(caseId.value)
  } catch (error: any) {
    handleError(error, '更新失敗')
  }
}

const handleFieldDelete = async (fieldId: string) => {
  try {
    const { deleteField } = useCaseFields()
    await deleteField(fieldId)
    await fetchFields()
    await fetchCase(caseId.value)
    handleSuccess('屬性已刪除')
  } catch (error: any) {
    handleError(error, '刪除失敗')
  }
}

// 用於 inline 編輯的暫存值
const editValues = ref<Record<string, string | number>>({})

const startEditing = () => {
  if (!currentCase.value) return
  editValues.value = {
    brand_name: currentCase.value.brand_name || '',
    collaboration_type: currentCase.value.collaboration_type || '',
    quoted_amount: currentCase.value.quoted_amount || 0,
    deadline_date: currentCase.value.deadline_date || '',
    contact_name: currentCase.value.contact_name || ''
  }
  isEditingProperties.value = true
}

const cancelEditing = () => {
  isEditingProperties.value = false
  editValues.value = {}
}

const saveEditing = async () => {
  try {
    await updateCase(caseId.value, editValues.value as any)
    handleSuccess('屬性已更新')
    isEditingProperties.value = false
    editValues.value = {}
    await fetchCase(caseId.value)
  } catch (error: any) {
    handleError(error, '更新失敗')
  }
}

// ── 階段管理 ──
const showPhaseDateEditor = ref(false)
const editingPhase = ref<any>(null)
const showApplyTemplate = ref(false)
const showPhaseDetail = ref(false)

const handleEditPhase = (phase: any) => {
  editingPhase.value = phase
  showPhaseDateEditor.value = true
}

const config = useRuntimeConfig()
const authStore = useAuthStore()
const apiHeaders = computed(() => ({
  Authorization: `Bearer ${authStore.token}`
}))

const handleDeletePhase = async (phase: any) => {
  try {
    await $fetch(
      `${config.public.apiBase}/api/v1/cases/${caseId.value}/phases/${phase.id}`,
      { method: 'DELETE', headers: apiHeaders.value }
    )
    handleSuccess('階段已刪除')
    await fetchCase(caseId.value)
  } catch (error: any) {
    handleError(error, '刪除失敗')
  }
}

const handleAddPhase = async () => {
  try {
    await $fetch(
      `${config.public.apiBase}/api/v1/cases/${caseId.value}/phases`,
      {
        method: 'POST',
        body: { name: '新階段', start_date: new Date().toISOString().split('T')[0], duration_days: 7 },
        headers: apiHeaders.value
      }
    )
    handleSuccess('階段已新增')
    await fetchCase(caseId.value)
  } catch (error: any) {
    handleError(error, '新增失敗')
  }
}

const handlePhaseDateUpdate = async (data: any) => {
  try {
    const phaseId = editingPhase.value?.id
    if (!phaseId) return
    await $fetch(
      `${config.public.apiBase}/api/v1/cases/${caseId.value}/phases/${phaseId}`,
      { method: 'PATCH', body: data, headers: apiHeaders.value }
    )
    handleSuccess('階段日期已更新')
    showPhaseDateEditor.value = false
    editingPhase.value = null
    await fetchCase(caseId.value)
  } catch (error: any) {
    handleError(error, '更新失敗')
  }
}

const handleApplyTemplate = async (data: ApplyTemplateRequest) => {
  try {
    await $fetch(
      `${config.public.apiBase}/api/v1/cases/${caseId.value}/phases/apply-template`,
      { method: 'POST', body: data, headers: apiHeaders.value }
    )
    handleSuccess('流程已套用')
    showApplyTemplate.value = false
    await fetchCase(caseId.value)
  } catch (error: any) {
    handleError(error, '套用失敗')
  }
}

const autoApplying = ref(false)
const handleAutoApplyTemplate = async () => {
  autoApplying.value = true
  try {
    const result = await $fetch<{ matched: boolean; message: string; reason: string; template_name?: string }>(
      `${config.public.apiBase}/api/v1/cases/${caseId.value}/phases/auto-apply`,
      { method: 'POST', headers: apiHeaders.value }
    )
    if (result.matched) {
      handleSuccess(result.message || 'AI 已自動套用流程')
      await fetchCase(caseId.value)
    } else {
      toast.add({ title: 'AI 無法自動套用', description: result.reason || '找不到適合的流程範本', color: 'warning' })
    }
  } catch (error: any) {
    const serverMsg = error?.data?.message || error?.response?._data?.message
    if (serverMsg) {
      toast.add({ title: 'AI 自動套用失敗', description: serverMsg, color: 'error' })
    } else {
      handleError(error, 'AI 自動套用失敗')
    }
  } finally {
    autoApplying.value = false
  }
}

const casePhases = computed(() => (currentCase.value as any)?.phases ?? [])
const caseStartDate = computed(() => (currentCase.value as any)?.start_date || new Date().toISOString().split('T')[0])
const caseEmails = computed(() => currentCase.value?.emails ?? [])

// ── 摘要列 computed ──
const deadlineDaysText = computed(() => {
  if (!currentCase.value?.deadline_date) return null
  const days = differenceInDays(new Date(currentCase.value.deadline_date), new Date())
  if (days < 0) return `已過期 ${Math.abs(days)} 天`
  if (days === 0) return '今天截止'
  return `剩 ${days} 天`
})

const formattedDeadline = computed(() => {
  if (!currentCase.value?.deadline_date) return '-'
  return format(new Date(currentCase.value.deadline_date), 'yyyy/MM/dd')
})

const displayTotal = computed(() => {
  const c = currentCase.value
  if (!c) return '-'
  const amount = c.collaboration_items_total || c.quoted_amount
  return amount ? formatAmount(amount) : '-'
})

// ── AI 擬信 Slideover ──
const showDraftReply = ref(false)

// ── 郵件詳情 Slideover ──
const showEmailDetail = ref(false)
const viewingEmailId = ref<string | null>(null)

const handleViewEmail = (emailId: string) => {
  viewingEmailId.value = emailId
  showEmailDetail.value = true
}
</script>

<template>
  <BaseDashboardPanel>
    <template #header>
      <BaseDashboardNavbar :title="currentCase?.title || '案件詳情'">
        <template #leading>
          <BaseDashboardSidebarCollapse />
        </template>
        <template #trailing>
          <div class="flex items-center gap-2">
            <BaseBadge
              v-if="currentCase"
              :color="getStatusColor(currentCase.status)"
              variant="subtle"
              size="lg"
            >
              {{ getStatusLabel(currentCase.status) }}
            </BaseBadge>
            <BaseButton icon="i-lucide-arrow-left" variant="ghost" @click="navigateTo('/cases')">
              返回列表
            </BaseButton>
          </div>
        </template>
      </BaseDashboardNavbar>
    </template>

    <template #body>
      <LoadingState v-if="loading || !pageReady" />

      <div v-else-if="currentCase" class="space-y-6">
        <!-- 非合作案件 -->
        <div v-if="currentCase.status === 'other'" class="max-w-2xl">
          <CasePropertiesPanel
            :case="currentCase"
            :fields="allFields"
            :editable="true"
            @field-update="handleFieldUpdate"
            @field-delete="handleFieldDelete"
          />
        </div>

        <!-- 合作案件：全寬（屬性 → 流程 → 郵件） -->
        <template v-else>
          <!-- ① 屬性摘要列 -->
          <div class="bg-subtle rounded-xl p-4">
            <div class="flex flex-wrap items-center gap-6">
              <!-- 品牌 -->
              <div class="min-w-0">
                <div class="text-xs text-dimmed mb-0.5">品牌</div>
                <template v-if="isEditingProperties">
                  <BaseInput v-model="editValues.brand_name" placeholder="品牌名稱" class="w-32" size="sm" />
                </template>
                <div v-else class="text-sm font-semibold text-highlighted truncate flex items-center gap-1.5">
                  <BaseIcon name="i-lucide-building-2" class="w-4 h-4 text-muted flex-shrink-0" />
                  {{ currentCase.brand_name || '-' }}
                </div>
              </div>

              <div class="w-px h-8 bg-gray-200 dark:bg-gray-700 hidden sm:block" />

              <!-- 合作類型 -->
              <div>
                <div class="text-xs text-dimmed mb-0.5">合作類型</div>
                <template v-if="isEditingProperties">
                  <BaseInput v-model="editValues.collaboration_type" placeholder="合作類型" class="w-28" size="sm" />
                </template>
                <template v-else>
                  <BaseBadge v-if="currentCase.collaboration_type" color="primary" variant="subtle" size="sm">
                    {{ currentCase.collaboration_type }}
                  </BaseBadge>
                  <span v-else class="text-sm text-muted">-</span>
                </template>
              </div>

              <div class="w-px h-8 bg-gray-200 dark:bg-gray-700 hidden sm:block" />

              <!-- 總價 -->
              <div>
                <div class="text-xs text-dimmed mb-0.5">總價</div>
                <template v-if="isEditingProperties">
                  <BaseInput v-model.number="editValues.quoted_amount" type="number" placeholder="報價金額" class="w-28" size="sm" />
                </template>
                <div v-else class="text-sm font-semibold text-highlighted">{{ displayTotal }}</div>
              </div>

              <div class="w-px h-8 bg-gray-200 dark:bg-gray-700 hidden sm:block" />

              <!-- 截止日 -->
              <div>
                <div class="text-xs text-dimmed mb-0.5">截止日</div>
                <template v-if="isEditingProperties">
                  <input v-model="editValues.deadline_date" type="date" class="w-36 px-2 py-1 text-sm rounded-md border border-default bg-default text-highlighted" />
                </template>
                <div v-else class="flex items-center gap-1.5">
                  <span class="text-sm font-semibold text-highlighted">{{ formattedDeadline }}</span>
                  <BaseBadge
                    v-if="deadlineDaysText"
                    :color="deadlineDaysText.includes('過期') ? 'error' : deadlineDaysText.includes('今天') ? 'warning' : 'neutral'"
                    variant="subtle"
                    size="xs"
                  >
                    {{ deadlineDaysText }}
                  </BaseBadge>
                </div>
              </div>

              <div class="w-px h-8 bg-gray-200 dark:bg-gray-700 hidden sm:block" />

              <!-- 聯絡人 -->
              <div>
                <div class="text-xs text-dimmed mb-0.5">聯絡人</div>
                <template v-if="isEditingProperties">
                  <BaseInput v-model="editValues.contact_name" placeholder="聯絡人" class="w-28" size="sm" />
                </template>
                <div v-else class="text-sm font-semibold text-highlighted flex items-center gap-1.5">
                  <BaseIcon name="i-lucide-user" class="w-4 h-4 text-muted flex-shrink-0" />
                  {{ currentCase.contact_name || '-' }}
                </div>
              </div>

              <!-- 分隔 + 編輯按鈕 -->
              <div class="w-px h-8 bg-gray-200 dark:bg-gray-700 hidden sm:block" />

              <div class="flex items-center gap-2 ml-auto">
                <template v-if="isEditingProperties">
                  <BaseButton size="sm" variant="outline" @click="cancelEditing">取消</BaseButton>
                  <BaseButton size="sm" @click="saveEditing">儲存</BaseButton>
                </template>
                <BaseButton
                  v-else
                  icon="i-lucide-edit"
                  size="sm"
                  variant="ghost"
                  @click="startEditing"
                >
                  編輯
                </BaseButton>
              </div>
            </div>
          </div>

          <!-- ② 專案流程 -->
          <BaseCard>
            <template #header>
              <div class="flex items-center justify-between w-full">
                <h2 class="text-lg font-semibold">專案流程</h2>
                <div class="flex gap-2">
                  <BaseButton icon="i-lucide-sparkles" size="sm" variant="outline" :loading="autoApplying" @click="handleAutoApplyTemplate">
                    AI 自動套用
                  </BaseButton>
                  <BaseButton icon="i-lucide-layout-template" size="sm" variant="outline" @click="showApplyTemplate = true">
                    {{ casePhases.length === 0 ? '手動套用' : '重新套用' }}
                  </BaseButton>
                  <BaseButton icon="i-lucide-plus" size="sm" variant="ghost" @click="handleAddPhase">
                    新增階段
                  </BaseButton>
                </div>
              </div>
            </template>

            <CasePhaseStepper :phases="casePhases" :editable="true" @edit-phase="handleEditPhase" @delete-phase="handleDeletePhase" />

            <div v-if="casePhases.length > 0" class="border-t border-default pt-3 mt-3">
              <button
                class="flex items-center gap-1.5 text-sm text-muted hover:text-highlighted transition-colors"
                @click="showPhaseDetail = !showPhaseDetail"
              >
                <UIcon :name="showPhaseDetail ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" class="w-4 h-4" />
                {{ showPhaseDetail ? '收合詳情' : '展開階段詳情' }}
              </button>
              <div v-if="showPhaseDetail" class="mt-3">
                <CasePhaseTimeline :phases="casePhases" :editable="true" @edit-phase="handleEditPhase" @delete-phase="handleDeletePhase" @add-phase="handleAddPhase" />
              </div>
            </div>
          </BaseCard>

          <!-- ③ 郵件往來 -->
          <AppSectionWithHeader title="郵件往來">
            <div class="space-y-4">
              <CaseEmailsTimeline :emails="caseEmails" :case-id="caseId" @view-email="handleViewEmail" />
              <BaseButton
                icon="i-lucide-reply"
                variant="outline"
                :disabled="caseEmails.length === 0"
                @click="showDraftReply = true"
              >
                回覆
              </BaseButton>
            </div>
          </AppSectionWithHeader>
        </template>
      </div>

      <ErrorState v-else title="無法載入案件詳情" message="請重新整理頁面或返回列表" />

      <!-- Modals & Slideovers -->
      <PhaseDateEditor v-model="showPhaseDateEditor" :phase="editingPhase" @submit="handlePhaseDateUpdate" />
      <ApplyTemplateModal v-model="showApplyTemplate" :case-start-date="caseStartDate" :case-id="caseId" @submit="handleApplyTemplate" />
      <DraftReplySlideover v-model="showDraftReply" :case-id="caseId" :case="currentCase" :emails="caseEmails" @sent="fetchCaseEmails(caseId)" />
      <EmailDetailSlideover v-model="showEmailDetail" :email-id="viewingEmailId" />
    </template>
  </BaseDashboardPanel>
</template>
