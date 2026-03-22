<script setup lang="ts">
import type { ApplyTemplateRequest } from '~/types/cases'
import { useCases } from '~/composables/useCases'
import { useCaseFields } from '~/composables/useCaseFields'
import { useErrorHandler } from '~/composables/useErrorHandler'
import { formatAmount } from '~/utils/formatters'
import { getStatusColor, getStatusLabel } from '~/utils/caseStatus'
import { useCollaborationItems } from '~/composables/useCollaborationItems'
import { BaseDashboardPanel, BaseDashboardNavbar, BaseButton, BaseCard, BaseBadge, BaseIcon, BaseInput } from '~/components/base'
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
const { fetchCase, fetchCaseEmails, currentCase, loading, updateCase, addCaseCollaborationItem, removeCaseCollaborationItem, updateFlowLayout } = useCases()
const { items: allCollaborationItems, individualItems: allIndividualItems, bundleItems: allBundleItems, fetchItems: fetchCollaborationItems } = useCollaborationItems()
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
      fetchFields().catch(() => {}),
      fetchCollaborationItems().catch(() => {})
    ])
    // 永遠載入郵件往來
    await fetchCaseEmails(caseId.value).catch(() => {})
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

// 未開案 = to_confirm，專案流程預設收合
const isCaseConfirmed = computed(() => currentCase.value?.status !== 'to_confirm')
const isFlowSectionExpanded = ref(false)

// 當案件載入後，根據狀態決定是否展開
watch(() => currentCase.value?.status, (status) => {
  if (status && status !== 'to_confirm') {
    isFlowSectionExpanded.value = true
  }
}, { immediate: true })

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

// 清空流程 double check
const showClearPhasesConfirm = ref(false)
const handleClearPhases = async () => {
  try {
    await $fetch(
      `${config.public.apiBase}/api/v1/cases/${caseId.value}/phases`,
      { method: 'DELETE', headers: apiHeaders.value }
    )
    handleSuccess('所有流程階段已清空')
    showClearPhasesConfirm.value = false
    await fetchCase(caseId.value)
  } catch (error: any) {
    handleError(error, '清空失敗')
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

const handleAddPhaseForItem = async (itemId: string | null) => {
  try {
    await $fetch(
      `${config.public.apiBase}/api/v1/cases/${caseId.value}/phases`,
      {
        method: 'POST',
        body: {
          name: '新階段',
          start_date: new Date().toISOString().split('T')[0],
          duration_days: 7,
          collaboration_item_id: itemId
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

// ── AI 自動匹配合作項目 ──
const autoMatchingItems = ref(false)
const handleAutoMatchItems = async () => {
  autoMatchingItems.value = true
  try {
    const result = await $fetch<{ matched: boolean; message?: string; reason?: string; matched_names?: string[] }>(
      `${config.public.apiBase}/api/v1/cases/${caseId.value}/auto-match-items`,
      { method: 'POST', headers: apiHeaders.value }
    )
    if (result.matched) {
      handleSuccess(result.message || 'AI 已自動匹配合作項目')
      await fetchCase(caseId.value)
    } else {
      toast.add({ title: 'AI 無法匹配', description: result.reason || '找不到適合的合作項目', color: 'warning' })
    }
  } catch (error: any) {
    const serverMsg = error?.data?.message || error?.response?._data?.message
    if (serverMsg) {
      toast.add({ title: 'AI 匹配失敗', description: serverMsg, color: 'error' })
    } else {
      handleError(error, 'AI 匹配失敗')
    }
  } finally {
    autoMatchingItems.value = false
  }
}

// ── 合作項目管理 ──
const showAddItemDropdown = ref(false)
const addItemBtnRef = ref<HTMLElement | null>(null)
const addItemDropdownRef = ref<HTMLElement | null>(null)

// 計算下拉選單位置（基於按鈕位置）
const addItemDropdownStyle = computed(() => {
  if (!addItemBtnRef.value) return {}
  const rect = addItemBtnRef.value.getBoundingClientRect()
  return {
    top: `${rect.bottom + 4}px`,
    right: `${window.innerWidth - rect.right}px`
  }
})

// 可新增的項目（排除已關聯的）
const availableItems = computed(() => {
  const linkedIds = new Set(caseCollaborationItems.value.map((c: any) => c.collaboration_item_id))
  return allCollaborationItems.value.filter(item => !linkedIds.has(item.id))
})

const handleAddCollaborationItem = async (itemId: string) => {
  try {
    await addCaseCollaborationItem(caseId.value, itemId)
    handleSuccess('已新增合作項目')
    showAddItemDropdown.value = false
  } catch (error: any) {
    handleError(error, '新增失敗')
  }
}

const handleRemoveCollaborationItem = async (itemId: string) => {
  try {
    await removeCaseCollaborationItem(caseId.value, itemId)
    handleSuccess('已移除合作項目')
  } catch (error: any) {
    handleError(error, '移除失敗')
  }
}

const handleToggleFlowLayout = async (layout: 'parallel' | 'sequential') => {
  if (currentCase.value?.flow_layout === layout) return
  try {
    await updateFlowLayout(caseId.value, layout)
    handleSuccess(layout === 'parallel' ? '已切換為並聯模式' : '已切換為串聯模式')
  } catch (error: any) {
    handleError(error, '切換失敗')
  }
}

const autoApplying = ref(false)
const handleAutoApplyTemplate = async () => {
  autoApplying.value = true
  try {
    const result = await $fetch<{ matched: boolean; message?: string; reason?: string; skipped_items?: string[] }>(
      `${config.public.apiBase}/api/v1/cases/${caseId.value}/phases/auto-apply`,
      { method: 'POST', headers: apiHeaders.value }
    )
    if (result.matched) {
      handleSuccess(result.message || '已套用流程')
      // 如果有跳過的項目，提示使用者
      if (result.skipped_items && result.skipped_items.length > 0) {
        toast.add({
          title: '部分項目未設定流程',
          description: `${result.skipped_items.join('、')} 尚未綁定流程範本`,
          color: 'warning'
        })
      }
      isFlowSectionExpanded.value = true
      await fetchCase(caseId.value)
    } else {
      toast.add({ title: '無法套用流程', description: result.reason || '找不到可套用的流程', color: 'warning' })
    }
  } catch (error: any) {
    const serverMsg = error?.data?.message || error?.response?._data?.message
    if (serverMsg) {
      toast.add({ title: '套用失敗', description: serverMsg, color: 'error' })
    } else {
      handleError(error, '套用失敗')
    }
  } finally {
    autoApplying.value = false
  }
}

const casePhases = computed(() => (currentCase.value as any)?.phases ?? [])
const caseStartDate = computed(() => (currentCase.value as any)?.start_date || new Date().toISOString().split('T')[0])
const caseEmails = computed(() => currentCase.value?.emails ?? [])

// ── 合作項目 ──
const caseCollaborationItems = computed(() => currentCase.value?.case_collaboration_items ?? [])
const hasCollaborationItems = computed(() => caseCollaborationItems.value.length > 0)

// 合作項目總價
const collaborationItemsTotal = computed(() => {
  return caseCollaborationItems.value.reduce((sum: number, cci: any) => {
    return sum + (cci.collaboration_item?.price || 0)
  }, 0)
})

// 取得合作項目名稱 by ID（用於流程階段標註）
// 需要同時查直接關聯的項目和 bundle 內含的項目
const getItemNameById = (itemId: string | undefined): string | null => {
  if (!itemId) return null
  // 先查直接關聯的
  const cci = caseCollaborationItems.value.find((c: any) => c.collaboration_item_id === itemId)
  if (cci?.collaboration_item?.title) return cci.collaboration_item.title
  // 再查 bundle 內含的 individual 項目
  for (const c of caseCollaborationItems.value) {
    const item = (c as any).collaboration_item
    if (item?.type === 'bundle' && item?.bundle_items) {
      const bi = item.bundle_items.find((b: any) => b.item_id === itemId)
      if (bi?.item?.title) return bi.item.title
    }
  }
  // 最後從全域項目列表查
  const globalItem = allCollaborationItems.value.find(i => i.id === itemId)
  return globalItem?.title || null
}

// 按合作項目分組的流程階段（含起始編號）
const phasesGroupedByItem = computed(() => {
  const groups: Array<{ itemId: string | null; itemName: string; phases: any[]; startIndex: number }> = []
  const phasesByItem = new Map<string, any[]>()

  for (const phase of casePhases.value) {
    const key = phase.collaboration_item_id || '__none__'
    if (!phasesByItem.has(key)) {
      phasesByItem.set(key, [])
    }
    phasesByItem.get(key)!.push(phase)
  }

  const isSequential = currentCase.value?.flow_layout === 'sequential'
  let runningIndex = 1

  for (const [key, phases] of phasesByItem) {
    const sorted = phases.sort((a: any, b: any) => a.order - b.order)
    const itemName = key === '__none__' ? '未分類' : (getItemNameById(key) || '未知項目')
    groups.push({
      itemId: key === '__none__' ? null : key,
      itemName,
      phases: sorted,
      startIndex: isSequential ? runningIndex : 1
    })
    runningIndex += sorted.length
  }

  return groups
})

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
  // 優先使用合作項目合計
  if (collaborationItemsTotal.value > 0) return formatAmount(collaborationItemsTotal.value)
  const amount = c.quoted_amount
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
        <template #right>
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

          <!-- ② 合作項目 -->
          <BaseCard>
            <template #header>
              <div class="flex items-center justify-between w-full">
                <h2 class="text-lg font-semibold">合作項目</h2>
                <div class="flex items-center gap-3">
                  <span v-if="hasCollaborationItems" class="text-sm font-semibold text-primary-600 dark:text-primary-400">
                    合計 {{ formatAmount(collaborationItemsTotal) }}
                  </span>
                  <!-- AI 自動匹配 -->
                  <BaseButton
                    icon="i-lucide-sparkles"
                    size="sm"
                    variant="outline"
                    :loading="autoMatchingItems"
                    @click="handleAutoMatchItems"
                  >
                    AI 匹配
                  </BaseButton>
                  <!-- 新增項目 -->
                  <div ref="addItemBtnRef">
                    <BaseButton
                      icon="i-lucide-plus"
                      size="sm"
                      variant="outline"
                      :disabled="availableItems.length === 0"
                      @click="showAddItemDropdown = !showAddItemDropdown"
                    >
                      新增項目
                    </BaseButton>
                  </div>
                </div>
              </div>
            </template>

            <!-- 已關聯的項目列表 -->
            <div v-if="hasCollaborationItems" class="space-y-1">
              <div
                v-for="cci in caseCollaborationItems"
                :key="cci.id"
                class="flex items-center justify-between gap-4 py-2 px-3 rounded-lg hover:bg-subtle transition-colors group"
              >
                <div class="flex items-center gap-2 min-w-0 flex-1">
                  <BaseIcon
                    :name="cci.collaboration_item?.type === 'bundle' ? 'i-lucide-package' : 'i-lucide-file-text'"
                    class="w-4 h-4 text-muted flex-shrink-0"
                  />
                  <div class="min-w-0 flex-1">
                    <div class="flex items-center gap-2">
                      <span class="font-medium text-highlighted truncate">
                        {{ cci.collaboration_item?.title || '未知項目' }}
                      </span>
                      <BaseBadge
                        v-if="cci.collaboration_item?.type === 'bundle'"
                        color="info"
                        variant="subtle"
                        size="xs"
                      >
                        組合
                      </BaseBadge>
                      <BaseBadge
                        v-if="cci.collaboration_item?.workflow"
                        color="neutral"
                        variant="subtle"
                        size="xs"
                      >
                        {{ cci.collaboration_item.workflow.name }}
                      </BaseBadge>
                    </div>
                    <!-- Bundle 內含項目 -->
                    <p
                      v-if="cci.collaboration_item?.type === 'bundle' && cci.collaboration_item?.bundle_items?.length"
                      class="text-xs text-muted mt-0.5"
                    >
                      包含：{{ cci.collaboration_item.bundle_items.map((bi: any) => bi.item?.title).filter(Boolean).join('、') }}
                    </p>
                  </div>
                </div>
                <div class="flex items-center gap-2">
                  <span class="text-sm font-semibold text-primary-600 dark:text-primary-400 whitespace-nowrap">
                    {{ formatAmount(cci.collaboration_item?.price || 0) }}
                  </span>
                  <BaseButton
                    icon="i-lucide-x"
                    size="xs"
                    variant="ghost"
                    color="neutral"
                    class="opacity-0 group-hover:opacity-100 transition-opacity"
                    @click="handleRemoveCollaborationItem(cci.collaboration_item_id)"
                  />
                </div>
              </div>
            </div>

            <!-- 空狀態 -->
            <div v-else class="text-center py-6 text-muted">
              <BaseIcon name="i-lucide-package" class="w-8 h-8 mx-auto mb-2 opacity-40" />
              <p class="text-sm">尚未選擇合作項目</p>
              <p class="text-xs text-dimmed mt-0.5">點擊「新增項目」加入合作項目</p>
            </div>
          </BaseCard>

          <!-- ③ 專案流程 -->
          <BaseCard>
            <template #header>
              <div class="flex items-center justify-between w-full">
                <div class="flex items-center gap-3">
                  <!-- 點擊標題可展開/收合 -->
                  <button
                    class="flex items-center gap-2 hover:opacity-80 transition-opacity"
                    @click="isFlowSectionExpanded = !isFlowSectionExpanded"
                  >
                    <UIcon
                      :name="isFlowSectionExpanded ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right'"
                      class="w-4 h-4 text-muted"
                    />
                    <h2 class="text-lg font-semibold">專案流程</h2>
                  </button>
                  <!-- 未開案標籤 -->
                  <BaseBadge
                    v-if="!isCaseConfirmed"
                    color="warning"
                    variant="subtle"
                    size="xs"
                  >
                    未開案
                  </BaseBadge>
                  <!-- 階段數量（收合時顯示） -->
                  <span v-if="!isFlowSectionExpanded && casePhases.length > 0" class="text-xs text-muted">
                    {{ casePhases.length }} 個階段
                  </span>
                  <!-- 並聯/串聯切換 -->
                  <div v-if="isFlowSectionExpanded && casePhases.length > 0 && phasesGroupedByItem.length > 1" class="flex items-center gap-1 bg-subtle rounded-lg p-0.5">
                    <button
                      class="text-xs px-2 py-1 rounded-md transition-colors"
                      :class="currentCase?.flow_layout === 'parallel' ? 'bg-default text-highlighted shadow-sm' : 'text-muted hover:text-highlighted'"
                      @click.stop="handleToggleFlowLayout('parallel')"
                    >
                      並聯
                    </button>
                    <button
                      class="text-xs px-2 py-1 rounded-md transition-colors"
                      :class="currentCase?.flow_layout === 'sequential' ? 'bg-default text-highlighted shadow-sm' : 'text-muted hover:text-highlighted'"
                      @click.stop="handleToggleFlowLayout('sequential')"
                    >
                      串聯
                    </button>
                  </div>
                </div>
                <div v-if="isFlowSectionExpanded" class="flex gap-2">
                  <BaseButton icon="i-lucide-zap" size="sm" variant="outline" :loading="autoApplying" @click="handleAutoApplyTemplate">
                    自動套用
                  </BaseButton>
                  <BaseButton icon="i-lucide-layout-template" size="sm" variant="outline" @click="showApplyTemplate = true">
                    手動套用
                  </BaseButton>
                  <BaseButton icon="i-lucide-plus" size="sm" variant="ghost" @click="handleAddPhase">
                    新增階段
                  </BaseButton>
                  <BaseButton
                    v-if="casePhases.length > 0"
                    icon="i-lucide-trash-2"
                    size="sm"
                    variant="ghost"
                    color="error"
                    @click="showClearPhasesConfirm = true"
                  >
                    清空
                  </BaseButton>
                </div>
                <!-- 收合狀態下的快捷操作 -->
                <div v-else class="flex gap-2">
                  <BaseButton icon="i-lucide-zap" size="xs" variant="ghost" :loading="autoApplying" @click.stop="handleAutoApplyTemplate">
                    自動套用
                  </BaseButton>
                  <BaseButton
                    icon="i-lucide-chevron-down"
                    size="xs"
                    variant="ghost"
                    @click="isFlowSectionExpanded = true"
                  >
                    展開
                  </BaseButton>
                </div>
              </div>
            </template>

            <!-- 展開的流程內容 -->
            <template v-if="isFlowSectionExpanded">
              <!-- 按合作項目分組顯示流程 -->
              <template v-if="phasesGroupedByItem.length > 1">
                <div v-for="group in phasesGroupedByItem" :key="group.itemId || '__none__'" class="mb-5 last:mb-0">
                  <div class="flex items-center justify-between mb-2">
                    <div class="flex items-center gap-2">
                      <BaseBadge color="primary" variant="subtle" size="xs">
                        {{ group.itemName }}
                      </BaseBadge>
                      <span class="text-xs text-muted">
                        {{ group.phases.length }} 個階段（{{ group.startIndex }}-{{ group.startIndex + group.phases.length - 1 }}）
                      </span>
                    </div>
                    <BaseButton
                      icon="i-lucide-plus"
                      size="xs"
                      variant="ghost"
                      @click="handleAddPhaseForItem(group.itemId)"
                    >
                      新增
                    </BaseButton>
                  </div>
                  <CasePhaseStepper :phases="group.phases" :start-index="group.startIndex" :editable="true" @edit-phase="handleEditPhase" @delete-phase="handleDeletePhase" />
                </div>
              </template>
              <template v-else>
                <CasePhaseStepper :phases="casePhases" :editable="true" @edit-phase="handleEditPhase" @delete-phase="handleDeletePhase" />
              </template>

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
            </template>
          </BaseCard>

        </template>

        <!-- 郵件往來（所有案件類型都顯示） -->
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
      </div>

      <ErrorState v-else title="無法載入案件詳情" message="請重新整理頁面或返回列表" />

      <!-- Modals & Slideovers -->
      <PhaseDateEditor v-model="showPhaseDateEditor" :phase="editingPhase" @submit="handlePhaseDateUpdate" />
      <ApplyTemplateModal v-model="showApplyTemplate" :case-start-date="caseStartDate" :case-id="caseId" @submit="handleApplyTemplate" />
      <DraftReplySlideover v-model="showDraftReply" :case-id="caseId" :case="currentCase" :emails="caseEmails" @sent="fetchCaseEmails(caseId)" />
      <EmailDetailSlideover v-model="showEmailDetail" :email-id="viewingEmailId" />

      <!-- 清空流程確認對話框 -->
      <UModal v-model:open="showClearPhasesConfirm">
        <template #content>
          <div class="p-6">
            <div class="flex items-center gap-3 mb-4">
              <div class="w-10 h-10 rounded-full bg-error/10 flex items-center justify-center">
                <BaseIcon name="i-lucide-alert-triangle" class="w-5 h-5 text-error" />
              </div>
              <div>
                <h3 class="text-lg font-semibold text-highlighted">確認清空所有流程階段？</h3>
                <p class="text-sm text-muted mt-0.5">此操作將刪除案件中的所有 {{ casePhases.length }} 個流程階段，無法復原。</p>
              </div>
            </div>
            <div class="flex justify-end gap-2 mt-6">
              <BaseButton variant="outline" @click="showClearPhasesConfirm = false">
                取消
              </BaseButton>
              <BaseButton color="error" @click="handleClearPhases">
                確認清空
              </BaseButton>
            </div>
          </div>
        </template>
      </UModal>

      <!-- 新增合作項目下拉選單（Teleport 到 body 避免被 overflow 截斷） -->
      <Teleport to="body">
        <div v-if="showAddItemDropdown" class="fixed inset-0 z-[9998]" @click="showAddItemDropdown = false" />
        <div
          v-if="showAddItemDropdown"
          ref="addItemDropdownRef"
          class="fixed z-[9999] w-72 max-h-64 overflow-y-auto rounded-lg border border-default bg-default shadow-lg"
          :style="addItemDropdownStyle"
        >
          <div v-if="availableItems.length === 0" class="p-3 text-sm text-muted text-center">
            所有項目都已新增
          </div>
          <template v-else>
            <!-- 單項 -->
            <div v-if="availableItems.filter(i => i.type === 'individual').length > 0" class="px-3 pt-2 pb-1">
              <span class="text-xs font-medium text-dimmed uppercase">單項</span>
            </div>
            <button
              v-for="item in availableItems.filter(i => i.type === 'individual')"
              :key="item.id"
              class="w-full flex items-center justify-between gap-2 px-3 py-2 text-left hover:bg-subtle transition-colors"
              @click="handleAddCollaborationItem(item.id)"
            >
              <div class="min-w-0 flex-1">
                <div class="text-sm font-medium text-highlighted truncate">{{ item.title }}</div>
                <div v-if="item.workflow" class="text-xs text-muted">{{ item.workflow.name }}</div>
              </div>
              <span class="text-xs font-semibold text-primary-600 dark:text-primary-400 flex-shrink-0">
                {{ formatAmount(item.price) }}
              </span>
            </button>
            <!-- 組合 -->
            <div v-if="availableItems.filter(i => i.type === 'bundle').length > 0" class="px-3 pt-2 pb-1 border-t border-default">
              <span class="text-xs font-medium text-dimmed uppercase">組合</span>
            </div>
            <button
              v-for="item in availableItems.filter(i => i.type === 'bundle')"
              :key="item.id"
              class="w-full flex items-center justify-between gap-2 px-3 py-2 text-left hover:bg-subtle transition-colors"
              @click="handleAddCollaborationItem(item.id)"
            >
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-1.5">
                  <span class="text-sm font-medium text-highlighted truncate">{{ item.title }}</span>
                  <BaseBadge color="info" variant="subtle" size="xs">組合</BaseBadge>
                </div>
                <div v-if="item.bundle_items?.length" class="text-xs text-muted mt-0.5">
                  {{ item.bundle_items.map((bi: any) => bi.item?.title).filter(Boolean).join('、') }}
                </div>
              </div>
              <span class="text-xs font-semibold text-primary-600 dark:text-primary-400 flex-shrink-0">
                {{ formatAmount(item.price) }}
              </span>
            </button>
          </template>
        </div>
      </Teleport>
    </template>
  </BaseDashboardPanel>
</template>
