<script setup lang="ts">
import type { CaseDetail, CaseCollaborationItem } from '~/types/cases'
import type { CollaborationItem } from '~/types/collaborationItems'
import { useCases } from '~/composables/useCases'
import { useErrorHandler } from '~/composables/useErrorHandler'
import { formatAmount } from '~/utils/formatters'
import { BaseButton, BaseIcon } from '~/components/base'
import CollaborationItemsEditor from '~/components/cases/fields/CollaborationItemsEditor.vue'
import AppSectionWithHeader from '~/components/ui/AppSectionWithHeader.vue'

interface Props {
  /** 案件詳情 */
  case: CaseDetail
  /** 是否可編輯 */
  editable?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  editable: true
})

const emit = defineEmits<{
  'update': []
}>()

const { updateCase, updateCaseCollaborationItem } = useCases()
const { handleError, handleSuccess } = useErrorHandler()

// 編輯合作項目（新增/移除）
const isEditing = ref(false)
// 編輯價格模式
const isEditingPrices = ref(false)
// 各項目暫存價格（collaboration_item_id → 價格字串）
const editingPrices = ref<Record<string, string>>({})
// 儲存中
const isSavingPrices = ref(false)

// 處理更新（合作項目新增/移除）
const handleUpdate = async (items: Array<{ id?: string; title: string; description?: string; price: number }>) => {
  try {
    const itemIds = items.map(item => item.id || `custom_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`)

    await updateCase(props.case.id, {
      collaboration_items: itemIds,
      collaboration_items_custom: items.filter(item => !item.id || item.id.startsWith('custom_')).map(item => ({
        id: item.id || `custom_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        title: item.title,
        description: item.description,
        price: item.price
      }))
    } as any)

    handleSuccess('合作項目已更新')
    isEditing.value = false
    emit('update')
  } catch (error: unknown) {
    handleError(error, '更新失敗')
  }
}

// 排序後的 case_collaboration_items
const sortedCCIs = computed(() => {
  const cciList = props.case.case_collaboration_items
  if (!cciList || cciList.length === 0) return []
  return [...cciList].sort((a, b) => a.order - b.order)
})

// 從 case_collaboration_items 多對多關聯取得選中的項目（用於編輯器）
const selectedItems = computed(() => {
  return sortedCCIs.value
    .map(cci => cci.collaboration_item)
    .filter((item): item is CollaborationItem => !!item)
})

// 取得案件中某項目的實際價格（優先用案件個別價格）
const getItemPrice = (cci: CaseCollaborationItem): number => {
  if (cci.price != null) return cci.price
  return cci.collaboration_item?.price || 0
}

// 判斷是否為自訂價格（與預設價格不同）
const isCustomPrice = (cci: CaseCollaborationItem): boolean => {
  return cci.price != null && cci.price !== cci.collaboration_item?.price
}

// 計算總價（使用案件個別價格）
const totalPrice = computed(() => {
  return sortedCCIs.value.reduce((sum, cci) => sum + getItemPrice(cci), 0)
})

// 進入價格編輯模式：將目前各項目價格填入暫存
const enterPriceEditMode = () => {
  const prices: Record<string, string> = {}
  for (const cci of sortedCCIs.value) {
    prices[cci.collaboration_item_id] = String(getItemPrice(cci))
  }
  editingPrices.value = prices
  isEditingPrices.value = true
}

// 取消價格編輯
const cancelPriceEdit = () => {
  isEditingPrices.value = false
  editingPrices.value = {}
}

// 儲存所有價格變更
const savePrices = async () => {
  isSavingPrices.value = true
  try {
    const promises: Promise<unknown>[] = []
    for (const cci of sortedCCIs.value) {
      const newPriceStr = editingPrices.value[cci.collaboration_item_id]
      if (newPriceStr === undefined) continue

      const newPrice = parseFloat(newPriceStr)
      if (isNaN(newPrice) || newPrice < 0) continue

      const currentPrice = getItemPrice(cci)
      if (newPrice === currentPrice) continue

      // 如果新價格等於預設價格，傳 null 清除自訂價格
      const defaultPrice = cci.collaboration_item?.price || 0
      const priceToSave = newPrice === defaultPrice ? null : newPrice
      promises.push(
        updateCaseCollaborationItem(props.case.id, cci.collaboration_item_id, { price: priceToSave })
      )
    }

    if (promises.length > 0) {
      await Promise.all(promises)
      handleSuccess('價格已更新')
      emit('update')
    }

    isEditingPrices.value = false
    editingPrices.value = {}
  } catch (error: unknown) {
    handleError(error, '更新價格失敗')
  } finally {
    isSavingPrices.value = false
  }
}

// 恢復某項目為預設價格
const resetToDefault = (cci: CaseCollaborationItem) => {
  const defaultPrice = cci.collaboration_item?.price || 0
  editingPrices.value[cci.collaboration_item_id] = String(defaultPrice)
}

// 編輯中的總價
const editingTotalPrice = computed(() => {
  return sortedCCIs.value.reduce((sum, cci) => {
    const priceStr = editingPrices.value[cci.collaboration_item_id]
    const price = priceStr !== undefined ? parseFloat(priceStr) : getItemPrice(cci)
    return sum + (isNaN(price) ? 0 : price)
  }, 0)
})

// 取得 bundle 內含項目名稱
const getBundleItemNames = (item: CollaborationItem): string[] => {
  if (item.type !== 'bundle' || !item.bundle_items) return []
  return item.bundle_items
    .sort((a, b) => a.order - b.order)
    .map(ref => ref.item?.title || '未知項目')
}
</script>

<template>
  <AppSectionWithHeader
    title="合作項目"
    description="與此案件相關的合作項目"
  >
    <template #actions>
      <div class="flex items-center gap-1">
        <!-- 價格編輯模式的按鈕 -->
        <template v-if="isEditingPrices">
          <BaseButton
            icon="i-lucide-x"
            variant="ghost"
            size="sm"
            :disabled="isSavingPrices"
            @click="cancelPriceEdit"
          >
            取消
          </BaseButton>
          <BaseButton
            icon="i-lucide-check"
            variant="solid"
            size="sm"
            :loading="isSavingPrices"
            @click="savePrices"
          >
            儲存價格
          </BaseButton>
        </template>
        <!-- 一般模式的按鈕 -->
        <template v-else-if="editable && !isEditing">
          <BaseButton
            v-if="sortedCCIs.length > 0"
            icon="i-lucide-dollar-sign"
            variant="ghost"
            size="sm"
            @click="enterPriceEditMode"
          >
            編輯價格
          </BaseButton>
          <BaseButton
            icon="i-lucide-edit"
            variant="ghost"
            size="sm"
            @click="isEditing = true"
          >
            編輯
          </BaseButton>
        </template>
        <!-- 合作項目編輯模式 -->
        <BaseButton
          v-else-if="isEditing"
          icon="i-lucide-x"
          variant="ghost"
          size="sm"
          @click="isEditing = false"
        >
          取消
        </BaseButton>
      </div>
    </template>

    <div class="case-collaboration-items">
      <!-- 一般顯示 / 價格編輯模式 -->
      <div v-if="!isEditing" class="space-y-1">
        <div v-if="sortedCCIs.length === 0" class="text-center py-6 text-muted">
          <BaseIcon name="i-lucide-package" class="w-8 h-8 mx-auto mb-2 opacity-40" />
          <p class="text-sm mb-0.5">尚未選擇合作項目</p>
          <p class="text-xs text-dimmed">點擊「編輯」新增合作項目</p>
        </div>

        <!-- 項目列表 -->
        <template v-for="cci in sortedCCIs" :key="cci.id">
          <div
            v-if="cci.collaboration_item"
            class="flex items-center justify-between gap-4 py-2 px-3 rounded-lg hover:bg-subtle transition-colors"
          >
            <div class="flex items-center gap-2 min-w-0 flex-1">
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2">
                  <h4 class="font-medium text-highlighted truncate">
                    {{ cci.collaboration_item.title }}
                  </h4>
                  <span
                    v-if="cci.collaboration_item.type === 'bundle'"
                    class="text-xs px-2 py-0.5 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded flex-shrink-0"
                  >
                    組合
                  </span>
                </div>
                <p v-if="cci.collaboration_item.description" class="text-sm text-muted mt-0.5 truncate">
                  {{ cci.collaboration_item.description }}
                </p>
                <p v-if="getBundleItemNames(cci.collaboration_item).length > 0" class="text-xs text-muted mt-0.5">
                  包含：{{ getBundleItemNames(cci.collaboration_item).join('、') }}
                </p>
              </div>
            </div>

            <!-- 價格區域 -->
            <div class="flex items-center gap-1.5 ml-4 flex-shrink-0">
              <!-- 價格編輯模式：顯示輸入框 -->
              <template v-if="isEditingPrices">
                <div class="flex items-center gap-1">
                  <span class="text-sm text-muted">$</span>
                  <input
                    v-model="editingPrices[cci.collaboration_item_id]"
                    type="number"
                    min="0"
                    step="100"
                    class="w-24 text-right text-sm font-semibold border border-primary-300 dark:border-primary-600 rounded px-2 py-1 bg-transparent focus:outline-none focus:ring-1 focus:ring-primary-500"
                  />
                </div>
                <!-- 恢復預設按鈕（僅在值與預設不同時顯示） -->
                <button
                  v-if="editingPrices[cci.collaboration_item_id] !== String(cci.collaboration_item?.price || 0)"
                  class="text-xs text-muted hover:text-highlighted transition-colors"
                  title="恢復預設價格"
                  @click="resetToDefault(cci)"
                >
                  <BaseIcon name="i-lucide-undo-2" class="w-3.5 h-3.5" />
                </button>
              </template>

              <!-- 一般模式：顯示價格 -->
              <template v-else>
                <span class="text-sm font-semibold text-primary-600 dark:text-primary-400 whitespace-nowrap">
                  {{ formatAmount(getItemPrice(cci)) }}
                </span>
                <span
                  v-if="isCustomPrice(cci)"
                  class="text-xs px-1.5 py-0.5 bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-300 rounded"
                >
                  自訂
                </span>
              </template>
            </div>
          </div>
        </template>

        <!-- 總價顯示 -->
        <div v-if="sortedCCIs.length > 0" class="mt-4 pt-4 border-t border-default">
          <div class="flex items-center justify-between">
            <span class="text-base font-semibold text-highlighted">
              總價
            </span>
            <span class="text-xl font-bold text-primary-600 dark:text-primary-400">
              {{ formatAmount(isEditingPrices ? editingTotalPrice : totalPrice) }}
            </span>
          </div>
        </div>
      </div>

      <!-- 合作項目編輯模式 -->
      <CollaborationItemsEditor
        v-else
        :items="selectedItems"
        @save="handleUpdate"
        @cancel="isEditing = false"
      />
    </div>
  </AppSectionWithHeader>
</template>

<style scoped>
.case-collaboration-items {
  min-height: 100px;
}

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
