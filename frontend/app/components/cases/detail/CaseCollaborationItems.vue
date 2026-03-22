<script setup lang="ts">
import type { CaseDetail } from '~/types/cases'
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

const { updateCase } = useCases()
const { handleError, handleSuccess } = useErrorHandler()

// 編輯模式
const isEditing = ref(false)

// 處理更新
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

// 從 case_collaboration_items 多對多關聯取得選中的項目
const selectedItems = computed(() => {
  const cciList = props.case.case_collaboration_items
  if (!cciList || cciList.length === 0) {
    return [] as CollaborationItem[]
  }

  return cciList
    .sort((a, b) => a.order - b.order)
    .map(cci => cci.collaboration_item)
    .filter((item): item is CollaborationItem => !!item)
})

// 計算總價（扁平列表簡單加總）
const totalPrice = computed(() => {
  return selectedItems.value.reduce((sum, item) => sum + (item.price || 0), 0)
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
      <BaseButton
        v-if="editable"
        :icon="isEditing ? 'i-lucide-x' : 'i-lucide-edit'"
        variant="ghost"
        size="sm"
        @click="isEditing = !isEditing"
      >
        {{ isEditing ? '取消' : '編輯' }}
      </BaseButton>
    </template>

    <div class="case-collaboration-items">
      <div v-if="!isEditing" class="space-y-1">
        <div v-if="selectedItems.length === 0" class="text-center py-6 text-muted">
          <BaseIcon name="i-lucide-package" class="w-8 h-8 mx-auto mb-2 opacity-40" />
          <p class="text-sm mb-0.5">尚未選擇合作項目</p>
          <p class="text-xs text-dimmed">點擊「編輯」新增合作項目</p>
        </div>

        <!-- 扁平渲染項目列表 -->
        <template v-for="item in selectedItems" :key="item.id">
          <div
            class="flex items-center justify-between gap-4 py-2 px-3 rounded-lg hover:bg-subtle transition-colors"
          >
            <div class="flex items-center gap-2 min-w-0 flex-1">
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2">
                  <h4 class="font-medium text-highlighted truncate">
                    {{ item.title }}
                  </h4>
                  <span
                    v-if="item.type === 'bundle'"
                    class="text-xs px-2 py-0.5 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded flex-shrink-0"
                  >
                    組合
                  </span>
                </div>
                <p v-if="item.description" class="text-sm text-muted mt-0.5 truncate">
                  {{ item.description }}
                </p>
                <!-- Bundle 內含項目 -->
                <p v-if="getBundleItemNames(item).length > 0" class="text-xs text-muted mt-0.5">
                  包含：{{ getBundleItemNames(item).join('、') }}
                </p>
              </div>
            </div>
            <span class="text-sm font-semibold text-primary-600 dark:text-primary-400 ml-4 flex-shrink-0 whitespace-nowrap">
              {{ formatAmount(item.price || 0) }}
            </span>
          </div>
        </template>

        <!-- 總價顯示 -->
        <div v-if="selectedItems.length > 0" class="mt-4 pt-4 border-t border-default">
          <div class="flex items-center justify-between">
            <span class="text-base font-semibold text-highlighted">
              總價
            </span>
            <span class="text-xl font-bold text-primary-600 dark:text-primary-400">
              {{ formatAmount(totalPrice) }}
            </span>
          </div>
        </div>
      </div>

      <!-- 編輯模式 -->
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
</style>
