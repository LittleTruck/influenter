<script setup lang="ts">
import type { CaseDetail } from '~/types/cases'
import type { CollaborationItem } from '~/types/collaborationItems'
import { useCollaborationItems } from '~/composables/useCollaborationItems'
import { useCases } from '~/composables/useCases'
import { useErrorHandler } from '~/composables/useErrorHandler'
import { formatAmount } from '~/utils/formatters'
import { BaseButton, BaseIcon } from '~/components/base'
import CollaborationItemsEditor from '~/components/cases/fields/CollaborationItemsEditor.vue'
import AppSectionWithHeader from '~/components/ui/AppSectionWithHeader.vue'
import BaseCollapsible from '~/components/base/BaseCollapsible.vue'

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
const { flatItems, buildTree } = useCollaborationItems()

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

// 構建樹形結構
const treeItems = computed(() => {
  return buildTree(selectedItems.value)
})

// 計算總價（遞歸計算所有項目）
const calculateTotalPrice = (items: Array<CollaborationItem & { isCustom?: boolean; children?: any[] }>): number => {
  return items.reduce((sum, item) => {
    const itemPrice = item.price || 0
    const childrenPrice = item.children ? calculateTotalPrice(item.children) : 0
    return sum + itemPrice + childrenPrice
  }, 0)
}

const totalPrice = computed(() => {
  return calculateTotalPrice(treeItems.value)
})

// 展開/收起的狀態
const expandedItems = ref<string[]>([])

const toggleItem = (itemId: string) => {
  const index = expandedItems.value.indexOf(itemId)
  if (index > -1) {
    expandedItems.value.splice(index, 1)
  } else {
    expandedItems.value.push(itemId)
  }
}

// 遞歸渲染項目
const renderItem = (item: CollaborationItem & { isCustom?: boolean; children?: any[] }, level = 0) => {
  const hasChildren = item.children && item.children.length > 0
  const isExpanded = expandedItems.value.includes(item.id)
  
  return {
    item,
    level,
    hasChildren,
    isExpanded
  }
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
        <div v-if="treeItems.length === 0" class="text-sm text-gray-500 dark:text-gray-400 p-4 text-center">
          尚未選擇合作項目
        </div>
        
        <!-- 遞歸渲染項目樹 -->
        <template v-for="item in treeItems" :key="item.id">
          <div class="collaboration-item">
            <div
              class="flex items-center justify-between gap-4 py-2 px-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
            >
              <div class="flex items-center gap-2 min-w-0 flex-1">
                <!-- 展開/收起按鈕 -->
                <BaseIcon
                  v-if="item.children && item.children.length > 0"
                  :name="expandedItems.includes(item.id) ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right'"
                  class="w-4 h-4 text-gray-400 cursor-pointer flex-shrink-0"
                  @click="toggleItem(item.id)"
                />
                <div v-else class="w-4 flex-shrink-0" />
                
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <h4 class="font-medium text-gray-900 dark:text-white truncate">
                      {{ item.title }}
                    </h4>
                    <span
                      v-if="item.isCustom"
                      class="text-xs px-2 py-0.5 bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300 rounded flex-shrink-0"
                    >
                      自訂
                    </span>
                  </div>
                  <p v-if="item.description" class="text-sm text-gray-500 dark:text-gray-400 mt-0.5 truncate">
                    {{ item.description }}
                  </p>
                </div>
              </div>
              <span class="text-sm font-semibold text-primary-600 dark:text-primary-400 ml-4 flex-shrink-0 whitespace-nowrap">
                {{ formatAmount(item.price || 0) }}
              </span>
            </div>
            
            <!-- 子項目 -->
            <BaseCollapsible
              v-if="item.children && item.children.length > 0"
              :open="expandedItems.includes(item.id)"
              @update:open="() => toggleItem(item.id)"
              :ui="{ content: 'pb-0 pl-6' }"
            >
              <template #content>
                <div class="space-y-1">
                  <template v-for="child in item.children" :key="child.id">
                    <div
                      class="flex items-center justify-between gap-4 py-2 px-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
                    >
                      <div class="flex items-center gap-2 min-w-0 flex-1">
                        <div class="w-4 flex-shrink-0" />
                        <div class="min-w-0 flex-1">
                          <div class="flex items-center gap-2">
                            <h4 class="font-medium text-gray-900 dark:text-white truncate">
                              {{ child.title }}
                            </h4>
                            <span
                              v-if="child.isCustom"
                              class="text-xs px-2 py-0.5 bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300 rounded flex-shrink-0"
                            >
                              自訂
                            </span>
                          </div>
                          <p v-if="child.description" class="text-sm text-gray-500 dark:text-gray-400 mt-0.5 truncate">
                            {{ child.description }}
                          </p>
                        </div>
                      </div>
                      <span class="text-sm font-semibold text-primary-600 dark:text-primary-400 ml-4 flex-shrink-0 whitespace-nowrap">
                        {{ formatAmount(child.price || 0) }}
                      </span>
                    </div>
                  </template>
                </div>
              </template>
            </BaseCollapsible>
          </div>
        </template>
        
        <!-- 總價顯示 -->
        <div v-if="treeItems.length > 0" class="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
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
