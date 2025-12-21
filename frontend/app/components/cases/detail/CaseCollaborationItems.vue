<script setup lang="ts">
import type { CaseDetail } from '~/types/cases'
import type { CollaborationItem } from '~/types/collaborationItems'
import { useCollaborationItems } from '~/composables/useCollaborationItems'
import { useCases } from '~/composables/useCases'
import { useErrorHandler } from '~/composables/useErrorHandler'
import { formatAmount } from '~/utils/formatters'
import { BaseButton } from '~/components/base'
import CollaborationItemsEditor from '~/components/cases/fields/CollaborationItemsEditor.vue'

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
const { flatItems } = useCollaborationItems()

// 編輯模式
const isEditing = ref(false)

// 處理更新
const handleUpdate = async (items: Array<{ id?: string; title: string; description?: string; price: number }>) => {
  try {
    // 將項目轉換為 ID 列表（自訂項目使用臨時 ID）
    const itemIds = items.map(item => item.id || `custom_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`)
    
    await updateCase(props.case.id, {
      collaboration_items: itemIds,
      // 同時保存自訂項目的完整資訊（用於顯示）
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

// 取得選中的項目（包括自訂項目）
const selectedItems = computed(() => {
  if (!props.case.collaboration_items || props.case.collaboration_items.length === 0) {
    return []
  }
  
  const items: Array<CollaborationItem & { isCustom?: boolean }> = []
  
  props.case.collaboration_items.forEach(id => {
    // 檢查是否為自訂項目
    const customItems = (props.case as any).collaboration_items_custom || []
    const customItem = customItems.find((item: any) => item.id === id)
    
    if (customItem) {
      items.push({
        ...customItem,
        isCustom: true,
        parent_id: null,
        order: 0,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      } as CollaborationItem & { isCustom?: boolean })
    } else {
      // 從預設列表查找
      const item = flatItems.value.find(i => i.id === id)
      if (item) {
        items.push(item)
      }
    }
  })
  
  return items
})

// 計算總價
const totalPrice = computed(() => {
  return selectedItems.value.reduce((sum, item) => sum + (item.price || 0), 0)
})
</script>

<template>
  <div class="case-collaboration-items">
    <div v-if="!isEditing" class="space-y-3">
      <div v-if="selectedItems.length === 0" class="text-sm text-gray-500 dark:text-gray-400 p-4 text-center">
        尚未選擇合作項目
      </div>
      <div v-else class="space-y-2">
        <div
          v-for="item in selectedItems"
          :key="item.id"
          class="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800/50"
        >
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <h4 class="font-medium text-gray-900 dark:text-white">
                {{ item.title }}
              </h4>
              <span
                v-if="item.isCustom"
                class="text-xs px-2 py-0.5 bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300 rounded"
              >
                自訂
              </span>
            </div>
            <p v-if="item.description" class="text-sm text-gray-500 dark:text-gray-400 mt-1">
              {{ item.description }}
            </p>
          </div>
          <span class="text-sm font-semibold text-primary-600 dark:text-primary-400 ml-4 flex-shrink-0">
            {{ formatAmount(item.price) }}
          </span>
        </div>
        
        <!-- 總價顯示 -->
        <div class="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
          <div class="flex items-center justify-between">
            <span class="text-base font-semibold text-gray-900 dark:text-white">
              總價
            </span>
            <span class="text-xl font-bold text-primary-600 dark:text-primary-400">
              {{ formatAmount(totalPrice) }}
            </span>
          </div>
        </div>
      </div>

      <!-- 編輯按鈕 -->
      <div v-if="editable" class="flex justify-end pt-2">
        <BaseButton
          icon="i-lucide-edit"
          variant="outline"
          size="sm"
          @click="isEditing = true"
        >
          編輯
        </BaseButton>
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
</template>

<style scoped>
.case-collaboration-items {
  min-height: 100px;
}
</style>
