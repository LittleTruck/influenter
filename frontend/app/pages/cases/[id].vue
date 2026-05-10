<script setup lang="ts">
import { useCases } from '~/composables/useCases'
import { useCaseFields } from '~/composables/useCaseFields'
import { useErrorHandler } from '~/composables/useErrorHandler'
import { formatAmount } from '~/utils/formatters'
import { getStatusColor, getStatusLabel } from '~/utils/caseStatus'
import { useCollaborationItems } from '~/composables/useCollaborationItems'
import { BaseDashboardPanel, BaseDashboardNavbar, BaseButton, BaseCard, BaseBadge, BaseIcon, BaseInput, BaseRichTextEditor } from '~/components/base'
import AppSectionWithHeader from '~/components/ui/AppSectionWithHeader.vue'
import CasePropertiesPanel from '~/components/cases/detail/CasePropertiesPanel.vue'
import CaseEmailsTimeline from '~/components/cases/detail/CaseEmailsTimeline.vue'
import CasePhaseTimeline from '~/components/cases/detail/CasePhaseTimeline.vue'
import CasePhaseStepper from '~/components/cases/detail/CasePhaseStepper.vue'
import DraftReplySlideover from '~/components/cases/detail/DraftReplySlideover.vue'
import EmailDetailSlideover from '~/components/cases/detail/EmailDetailSlideover.vue'
import PhaseManagerModal from '~/components/cases/detail/PhaseManagerModal.vue'
import LoadingState from '~/components/common/LoadingState.vue'
import ErrorState from '~/components/common/ErrorState.vue'
import { format, differenceInDays } from 'date-fns'

definePageMeta({
  middleware: 'auth'
})

const route = useRoute()
const { fetchCase, fetchCaseEmails, currentCase, loading, updateCase, addCaseCollaborationItem, removeCaseCollaborationItem, updateCaseCollaborationItem } = useCases()
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
    deadline_date: currentCase.value.deadline_date || '',
    contact_name: currentCase.value.contact_name || '',
    contact_email: currentCase.value.contact_email || '',
    agency_name: currentCase.value.agency_name || '',
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
const showPhaseManager = ref(false)
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

const config = useRuntimeConfig()
const authStore = useAuthStore()
const apiHeaders = computed(() => ({
  Authorization: `Bearer ${authStore.token}`
}))

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

const handlePhasesSaved = async () => {
  await fetchCase(caseId.value)
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
const caseEmails = computed(() => currentCase.value?.emails ?? [])

// ── 合作項目 ──
const caseCollaborationItems = computed(() => currentCase.value?.case_collaboration_items ?? [])
const hasCollaborationItems = computed(() => caseCollaborationItems.value.length > 0)

// 取得案件中某項目的實際價格（案件個別價格優先）
const getItemPrice = (cci: any): number => {
  if (cci.price != null) return cci.price
  return cci.collaboration_item?.price || 0
}

// 判斷是否為自訂價格
const isCustomPrice = (cci: any): boolean => {
  return cci.price != null && cci.price !== cci.collaboration_item?.price
}

// 合作項目總價（使用案件個別價格）
const collaborationItemsTotal = computed(() => {
  return caseCollaborationItems.value.reduce((sum: number, cci: any) => {
    return sum + getItemPrice(cci)
  }, 0)
})

// ── 單項價格 inline 編輯 ──
const editingPriceItemId = ref<string | null>(null)
const editingPriceValue = ref('')
const savingPriceItemId = ref<string | null>(null)

const startEditItemPrice = (cci: any) => {
  editingPriceItemId.value = cci.collaboration_item_id
  editingPriceValue.value = String(getItemPrice(cci))
}

const cancelEditItemPrice = () => {
  editingPriceItemId.value = null
}

const saveItemPrice = async (cci: any) => {
  const newPrice = parseFloat(editingPriceValue.value)
  if (isNaN(newPrice) || newPrice < 0) {
    editingPriceItemId.value = null
    return
  }
  if (newPrice === getItemPrice(cci)) {
    editingPriceItemId.value = null
    return
  }

  const itemId = cci.collaboration_item_id
  savingPriceItemId.value = itemId
  try {
    const defaultPrice = cci.collaboration_item?.price || 0
    const priceToSave = newPrice === defaultPrice ? null : newPrice
    await updateCaseCollaborationItem(caseId.value, itemId, { price: priceToSave })
    handleSuccess('價格已更新')
  } catch (error: unknown) {
    handleError(error, '更新價格失敗')
  } finally {
    editingPriceItemId.value = null
    savingPriceItemId.value = null
  }
}

const handlePriceKeydown = (event: KeyboardEvent, cci: any) => {
  if (event.key === 'Enter') saveItemPrice(cci)
  else if (event.key === 'Escape') cancelEditItemPrice()
}

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

// 合作項目 ID → 名稱 的對照表（供 PhaseManagerModal 使用）
// 包含所有案件關聯的合作項目（含 bundle 內含），以便新增範本流程時也能顯示 badge
const itemNameMap = computed(() => {
  const map: Record<string, string> = {}
  // 從現有階段收集
  for (const phase of casePhases.value) {
    const id = phase.collaboration_item_id
    if (id && !map[id]) {
      map[id] = getItemNameById(id) || '未知項目'
    }
  }
  // 從案件關聯的合作項目補齊（含 bundle 內含的 individual）
  for (const cci of caseCollaborationItems.value) {
    const item = (cci as any).collaboration_item
    if (item?.id && !map[item.id]) {
      map[item.id] = item.title || '未知項目'
    }
    if (item?.type === 'bundle' && item?.bundle_items) {
      for (const bi of item.bundle_items) {
        if (bi.item?.id && !map[bi.item.id]) {
          map[bi.item.id] = bi.item.title || '未知項目'
        }
      }
    }
  }
  return map
})

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

// ── 備註 ──
const isEditingNotes = ref(false)
const notesContent = ref('')

const startEditingNotes = () => {
  notesContent.value = currentCase.value?.notes || ''
  isEditingNotes.value = true
}

const cancelEditingNotes = () => {
  isEditingNotes.value = false
  notesContent.value = ''
}

const notesPlainText = computed(() => {
  const html = currentCase.value?.notes || ''
  if (!html) return ''
  return html.replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim()
})

const notesTruncatedHtml = computed(() => {
  const plain = notesPlainText.value
  if (!plain) return ''
  return plain.length > 30 ? plain.slice(0, 30) + '…' : plain
})

const saveNotes = async () => {
  try {
    await updateCase(caseId.value, { notes: notesContent.value } as any)
    handleSuccess('備註已更新')
    isEditingNotes.value = false
    await fetchCase(caseId.value)
  } catch (error: any) {
    handleError(error, '更新備註失敗')
  }
}

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
      <BaseDashboardNavbar :title="currentCase ? (currentCase.agency_name ? `${currentCase.agency_name} - ${currentCase.title}` : currentCase.title) : '案件詳情'">
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
              <!-- 代理商 -->
              <div class="min-w-0" style="max-width: 140px;">
                <div class="text-xs text-dimmed mb-0.5">代理商</div>
                <template v-if="isEditingProperties">
                  <BaseInput v-model="editValues.agency_name" placeholder="代理商名稱" class="w-28" size="sm" />
                </template>
                <div v-else class="text-sm font-semibold text-highlighted truncate">
                  {{ currentCase.agency_name || '-' }}
                </div>
              </div>

              <div class="w-px h-8 bg-gray-200 dark:bg-gray-700 hidden sm:block" />

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

              <!-- 總價（由合作項目加總，不可手動編輯） -->
              <div>
                <div class="text-xs text-dimmed mb-0.5">總價</div>
                <div class="text-sm font-semibold text-highlighted">{{ displayTotal }}</div>
              </div>

              <div class="w-px h-8 bg-gray-200 dark:bg-gray-700 hidden sm:block" />

              <!-- 預計上線日 -->
              <div>
                <div class="text-xs text-dimmed mb-0.5">預計上線日</div>
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

              <div class="w-px h-8 bg-gray-200 dark:bg-gray-700 hidden sm:block" />

              <!-- 聯絡人信箱 -->
              <div class="min-w-0" style="max-width: 200px;">
                <div class="text-xs text-dimmed mb-0.5">聯絡人信箱</div>
                <template v-if="isEditingProperties">
                  <BaseInput v-model="editValues.contact_email" type="email" placeholder="email@example.com" class="w-44" size="sm" />
                </template>
                <div v-else class="text-sm font-semibold text-highlighted flex items-center gap-1.5 truncate">
                  <BaseIcon name="i-lucide-mail" class="w-4 h-4 text-muted flex-shrink-0" />
                  <span class="truncate" :title="currentCase.contact_email || ''">{{ currentCase.contact_email || '-' }}</span>
                </div>
              </div>

              <div class="w-px h-8 bg-gray-200 dark:bg-gray-700 hidden sm:block" />

              <!-- 備註 -->
              <div class="min-w-0" style="max-width: 200px;">
                <div class="text-xs text-dimmed mb-0.5">備註</div>
                <div
                  v-if="currentCase.notes"
                  class="text-sm font-semibold text-highlighted truncate cursor-pointer"
                  :title="notesPlainText"
                  @click="startEditingNotes"
                >{{ notesTruncatedHtml }}</div>
                <span
                  v-else
                  class="text-sm text-muted cursor-pointer hover:text-highlighted transition-colors"
                  @click="startEditingNotes"
                >
                  點擊新增...
                </span>
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
                  <!-- 正在編輯此項目的價格 -->
                  <template v-if="editingPriceItemId === cci.collaboration_item_id">
                    <div class="flex items-center gap-1">
                      <span class="text-sm text-muted">$</span>
                      <input
                        v-model="editingPriceValue"
                        type="number"
                        min="0"
                        step="100"
                        autofocus
                        class="w-24 text-right text-sm font-semibold border border-primary-300 dark:border-primary-600 rounded px-2 py-1 bg-transparent focus:outline-none focus:ring-1 focus:ring-primary-500"
                        @keydown="handlePriceKeydown($event, cci)"
                      />
                    </div>
                    <BaseButton
                      icon="i-lucide-check"
                      size="xs"
                      variant="solid"
                      :loading="savingPriceItemId === cci.collaboration_item_id"
                      @click="saveItemPrice(cci)"
                    />
                    <BaseButton
                      icon="i-lucide-x"
                      size="xs"
                      variant="ghost"
                      @click="cancelEditItemPrice"
                    />
                  </template>
                  <!-- 一般顯示 -->
                  <template v-else>
                    <span class="text-sm font-semibold text-primary-600 dark:text-primary-400 whitespace-nowrap">
                      {{ formatAmount(getItemPrice(cci)) }}
                    </span>
                    <BaseBadge
                      v-if="isCustomPrice(cci)"
                      color="warning"
                      variant="subtle"
                      size="xs"
                    >
                      自訂
                    </BaseBadge>
                    <BaseButton
                      icon="i-lucide-pencil"
                      size="xs"
                      variant="ghost"
                      color="neutral"
                      title="編輯價格"
                      @click="startEditItemPrice(cci)"
                    />
                    <BaseButton
                      icon="i-lucide-x"
                      size="xs"
                      variant="ghost"
                      color="neutral"
                      class="opacity-0 group-hover:opacity-100 transition-opacity"
                      @click="handleRemoveCollaborationItem(cci.collaboration_item_id)"
                    />
                  </template>
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
                  <!-- 並聯/串聯標示 -->
                  <BaseBadge
                    v-if="isFlowSectionExpanded && casePhases.length > 0"
                    color="neutral"
                    variant="subtle"
                    size="xs"
                  >
                    {{ currentCase?.flow_layout === 'sequential' ? '串聯' : '並聯' }}
                  </BaseBadge>
                </div>
                <div v-if="isFlowSectionExpanded" class="flex gap-2">
                  <BaseButton icon="i-lucide-zap" size="sm" variant="outline" :loading="autoApplying" @click="handleAutoApplyTemplate">
                    自動套用
                  </BaseButton>
                  <BaseButton icon="i-lucide-edit" size="sm" variant="outline" @click="showPhaseManager = true">
                    編輯流程
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
                  </div>
                  <CasePhaseStepper :phases="group.phases" :start-index="group.startIndex" />
                </div>
              </template>
              <template v-else>
                <CasePhaseStepper :phases="casePhases" />
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
                  <CasePhaseTimeline :phases="casePhases" />
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

      <!-- 備註編輯 Modal -->
      <UModal v-model:open="isEditingNotes">
        <template #content>
          <div class="p-5 space-y-4">
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold text-highlighted">編輯備註</h3>
              <BaseButton icon="i-lucide-x" size="xs" variant="ghost" @click="cancelEditingNotes" />
            </div>
            <BaseRichTextEditor
              v-model="notesContent"
              placeholder="自由輸入備註內容..."
              min-height="200px"
              :toolbar="true"
            />
            <div class="flex justify-end gap-2">
              <BaseButton variant="outline" @click="cancelEditingNotes">取消</BaseButton>
              <BaseButton @click="saveNotes">儲存</BaseButton>
            </div>
          </div>
        </template>
      </UModal>

      <!-- Modals & Slideovers -->
      <PhaseManagerModal v-model="showPhaseManager" :phases="casePhases" :case-id="caseId" :item-name-map="itemNameMap" :flow-layout="currentCase?.flow_layout" @saved="handlePhasesSaved" />
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

<style scoped>
/* 隱藏 number input 的上下箭頭 */
input[type="number"]::-webkit-inner-spin-button,
input[type="number"]::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
input[type="number"] {
  -moz-appearance: textfield;
}
</style>
