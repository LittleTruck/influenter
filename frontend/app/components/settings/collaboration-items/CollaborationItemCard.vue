<script setup lang="ts">
import type { CollaborationItem, WorkflowTemplate } from '~/types/collaborationItems'
import { useWorkflowTemplates } from '~/composables/useWorkflowTemplates'
import { formatAmount } from '~/utils/formatters'
import DraggableList from '~/components/base/DraggableList.vue'
import { BaseButton, BaseIcon, BaseBadge, BaseCollapsible } from '~/components/base'
import DraggableItemCard from './DraggableItemCard.vue'

interface Props {
  /** 項目資料 */
  item: CollaborationItem
  /** 縮排層級 */
  level?: number
  /** 是否展開 */
  expanded?: boolean
  /** 父項目（用於繼承流程） */
  parentItem?: CollaborationItem | null
}

const props = withDefaults(defineProps<Props>(), {
  level: 0,
  expanded: false,
  parentItem: null
})

const { findWorkflowById } = useWorkflowTemplates()

// 計算實際使用的流程（優先使用自己的，否則繼承父項目的）
const effectiveWorkflow = computed<WorkflowTemplate | null>(() => {
  // 如果項目有自己的流程，使用自己的
  if (props.item.workflow_id && props.item.workflow) {
    return props.item.workflow
  }
  
  // 如果項目有自己的 workflow_id，嘗試查找
  if (props.item.workflow_id) {
    return findWorkflowById(props.item.workflow_id)
  }
  
  // 否則繼承父項目的流程
  if (props.parentItem) {
    if (props.parentItem.workflow) {
      return props.parentItem.workflow
    }
    if (props.parentItem.workflow_id) {
      return findWorkflowById(props.parentItem.workflow_id)
    }
  }
  
  return null
})

const emit = defineEmits<{
  'add-item': [parentId: string]
  'edit-item': [item: CollaborationItem]
  'delete-item': [item: CollaborationItem]
  'reorder': [itemIds: string[], parentId: string | null]
  'toggle-expand': [item: CollaborationItem]
}>()

const isExpanded = ref(props.expanded)

// 本地子項目列表（用於拖曳）
const localChildren = ref<CollaborationItem[]>([...(props.item.children || [])])

// 同步外部 children 變化
watch(() => props.item.children, (newChildren) => {
  localChildren.value = [...(newChildren || [])]
}, { deep: true, immediate: true })

// 處理子項目拖曳重排序
const handleReorderChildren = (itemIds: string[]) => {
  emit('reorder', itemIds, props.item.id)
}

// 監聽 expanded prop 變化
watch(() => props.expanded, (newValue) => {
  isExpanded.value = newValue
})

// 監聽 isExpanded 變化，同步到父組件
watch(() => isExpanded.value, (newValue) => {
  if (props.item.children && props.item.children.length > 0) {
    emit('toggle-expand', props.item)
  }
})
</script>

<template>
  <div class="collaboration-item-card">
    <BaseCollapsible
      v-if="item.children && item.children.length > 0"
      v-model:open="isExpanded"
      class="collapsible-item"
      :ui="{ content: 'pb-0 mb-0' }"
    >
      <!-- 卡片主體 -->
      <div class="flex items-center gap-3 p-3 border border-gray-200 dark:border-gray-700 rounded-lg hover:bg-white dark:hover:bg-gray-700/50 transition-colors cursor-pointer bg-white dark:bg-gray-900/50">
        <div class="flex items-center flex-1 min-w-0 gap-3">
          <!-- 展開/收起按鈕 -->
          <BaseIcon
            :name="isExpanded ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right'"
            class="w-5 h-5 flex-shrink-0"
          />

          <!-- 拖曳手柄 -->
          <BaseIcon
            name="i-lucide-grip-vertical"
            class="w-5 h-5 text-gray-400 drag-handle cursor-grab flex-shrink-0"
            @click.stop
          />

          <!-- 項目資訊 -->
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <h4 class="font-medium text-gray-900 dark:text-white truncate">
                {{ item.title }}
              </h4>
              <BaseBadge color="primary" variant="subtle" size="xs">
                {{ formatAmount(item.price) }}
              </BaseBadge>
            </div>
          </div>
        </div>

        <!-- 操作按鈕 -->
        <div class="flex items-center gap-1 flex-shrink-0 ml-2" @click.stop>
          <BaseButton
            icon="i-lucide-plus"
            variant="ghost"
            size="xs"
            @click="emit('add-item', item.id)"
          >
            子項目
          </BaseButton>
          <BaseButton
            icon="i-lucide-edit"
            variant="ghost"
            size="xs"
            @click="emit('edit-item', item)"
          />
          <BaseButton
            icon="i-lucide-trash-2"
            variant="ghost"
            size="xs"
            color="error"
            @click="emit('delete-item', item)"
          />
        </div>
      </div>

      <!-- 子項目列表（展開時顯示） -->
      <template #content>
        <div class="pl-12 py-2 bg-gray-50 dark:bg-gray-800/30">
          <DraggableList
            v-model:items="localChildren"
            group-name="collaboration-items"
            @reorder="handleReorderChildren"
          >
            <template #item="{ element }">
              <DraggableItemCard
                :item="element"
                :show-expand="!!(element.children && element.children.length > 0)"
                :expanded="false"
                @edit="emit('edit-item', element)"
                @delete="emit('delete-item', element)"
              >
                <template #content="{ item: element }">
                  <div class="flex items-center gap-2">
                    <h4 class="font-medium text-gray-900 dark:text-white truncate">
                      {{ element.title }}
                    </h4>
                    <BaseBadge color="primary" variant="subtle" size="xs">
                      {{ formatAmount(element.price) }}
                    </BaseBadge>
                  </div>
                </template>
                <template #actions="{ item: element }">
                  <BaseButton
                    icon="i-lucide-plus"
                    variant="ghost"
                    size="xs"
                    @click="emit('add-item', element.id)"
                  >
                    子項目
                  </BaseButton>
                  <BaseButton
                    icon="i-lucide-edit"
                    variant="ghost"
                    size="xs"
                    @click="emit('edit-item', element)"
                  />
                  <BaseButton
                    icon="i-lucide-trash-2"
                    variant="ghost"
                    size="xs"
                    color="error"
                    @click="emit('delete-item', element)"
                  />
                </template>
              </DraggableItemCard>
            </template>
          </DraggableList>
        </div>
      </template>
    </BaseCollapsible>
    <!-- 沒有子項目的項目 -->
    <div
      v-else
      class="flex items-center gap-3 p-3 border border-gray-200 dark:border-gray-700 rounded-lg hover:bg-white dark:hover:bg-gray-700/50 transition-colors bg-white dark:bg-gray-900/50"
    >
      <div class="w-5 flex-shrink-0" />
      <BaseIcon
        name="i-lucide-grip-vertical"
        class="w-5 h-5 text-gray-400 drag-handle cursor-grab flex-shrink-0"
      />
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2">
          <h4 class="font-medium text-gray-900 dark:text-white truncate">
            {{ item.title }}
          </h4>
          <BaseBadge color="primary" variant="subtle" size="xs">
            {{ formatAmount(item.price) }}
          </BaseBadge>
        </div>
      </div>
      <div class="flex items-center gap-1 flex-shrink-0 ml-2" @click.stop>
        <BaseButton
          icon="i-lucide-plus"
          variant="ghost"
          size="xs"
          @click="emit('add-item', item.id)"
        >
          子項目
        </BaseButton>
        <BaseButton
          icon="i-lucide-edit"
          variant="ghost"
          size="xs"
          @click="emit('edit-item', item)"
        />
        <BaseButton
          icon="i-lucide-trash-2"
          variant="ghost"
          size="xs"
          color="error"
          @click="emit('delete-item', item)"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 拖曳樣式已統一在 DraggableList 組件中 */
</style>
