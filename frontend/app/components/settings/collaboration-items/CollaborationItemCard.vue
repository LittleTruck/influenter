<script setup lang="ts">
import type { CollaborationItem } from '~/types/collaborationItems'
import { formatAmount } from '~/utils/formatters'
import draggable from 'vuedraggable'
import type { DragChangeEvent } from '~/types/dragEvents'
import { BaseButton, BaseIcon, BaseBadge } from '~/components/base'

interface Props {
  /** 項目資料 */
  item: CollaborationItem
  /** 縮排層級 */
  level?: number
  /** 是否展開 */
  expanded?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  level: 0,
  expanded: false
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
const localChildren = computed({
  get: () => props.item.children || [],
  set: async (newChildren) => {
    // 拖曳後更新順序
    const itemIds = newChildren.map(item => item.id)
    emit('reorder', itemIds, props.item.id)
  }
})

// 拖曳選項 - 禁止跨層級拖曳
const dragOptions = computed(() => ({
  animation: 200,
  group: {
    name: `items-level-${props.level}`,
    pull: false,
    put: false
  },
  disabled: false,
  ghostClass: 'sortable-ghost',
  chosenClass: 'sortable-chosen',
  dragClass: 'sortable-drag',
  handle: '.drag-handle'
}))

// 處理拖曳變更
const handleChange = (evt: DragChangeEvent<CollaborationItem>) => {
  if (evt.moved || evt.added) {
    const itemIds = localChildren.value.map(item => item.id)
    emit('reorder', itemIds, props.item.id)
  }
}

// 切換展開/收起
const toggleExpand = () => {
  if (props.item.children && props.item.children.length > 0) {
    isExpanded.value = !isExpanded.value
    emit('toggle-expand', props.item)
  }
}

// 監聽 expanded prop 變化
watch(() => props.expanded, (newValue) => {
  isExpanded.value = newValue
})
</script>

<template>
  <div class="collaboration-item-card">
    <!-- 卡片主體 -->
    <div
      :class="[
        'flex items-center justify-between p-3 border border-gray-200 dark:border-gray-700 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors',
        `level-${level}`
      ]"
      :style="{ marginLeft: `${level * 1.5}rem` }"
    >
      <div class="flex items-center flex-1 min-w-0 gap-2">
        <!-- 展開/收起按鈕 -->
        <BaseButton
          v-if="item.children && item.children.length > 0"
          :icon="isExpanded ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right'"
          variant="ghost"
          size="xs"
          class="flex-shrink-0 transition-transform"
          @click.stop="toggleExpand"
        />
        <div v-else class="w-5 flex-shrink-0" />

        <!-- 拖曳手柄 -->
        <div class="drag-handle cursor-grab active:cursor-grabbing flex-shrink-0">
          <BaseIcon
            name="i-lucide-grip-vertical"
            class="w-4 h-4 text-gray-400"
          />
        </div>

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
          <p v-if="item.description" class="text-sm text-gray-500 dark:text-gray-400 mt-1 line-clamp-1">
            {{ item.description }}
          </p>
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
    <Transition
      enter-active-class="transition-all duration-200 ease-out"
      enter-from-class="opacity-0 max-h-0"
      enter-to-class="opacity-100 max-h-[2000px]"
      leave-active-class="transition-all duration-200 ease-in"
      leave-from-class="opacity-100 max-h-[2000px]"
      leave-to-class="opacity-0 max-h-0"
    >
      <div v-if="isExpanded && item.children && item.children.length > 0" class="children-container overflow-hidden">
        <div class="space-y-2 pt-2">
          <draggable
            v-model="localChildren"
            v-bind="dragOptions"
            item-key="id"
            @change="handleChange"
          >
            <template #item="{ element }">
              <CollaborationItemCard
                :item="element"
                :level="level + 1"
                @add-item="emit('add-item', $event)"
                @edit-item="emit('edit-item', $event)"
                @delete-item="emit('delete-item', $event)"
                @reorder="emit('reorder', $event[0], $event[1])"
                @toggle-expand="emit('toggle-expand', $event)"
              />
            </template>
          </draggable>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.children-container {
  margin-left: 0.5rem;
}

/* 拖曳效果 */
:deep(.sortable-ghost) {
  opacity: 0.5;
  background: rgba(var(--color-primary-500) / 0.1);
}

:deep(.sortable-chosen) {
  transform: scale(1.02);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  z-index: 10;
}

:deep(.sortable-drag) {
  cursor: grabbing !important;
  z-index: 1000;
}
</style>




