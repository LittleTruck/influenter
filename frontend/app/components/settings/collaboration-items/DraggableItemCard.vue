<script setup lang="ts" generic="T extends { id: string }">
import { BaseIcon, BaseButton } from '~/components/base'

interface Props {
  /** 項目資料 */
  item: T
  /** 是否為第一個項目 */
  isFirst?: boolean
  /** 拖曳手柄圖標 */
  dragHandleIcon?: string
  /** 是否顯示展開按鈕 */
  showExpand?: boolean
  /** 是否展開 */
  expanded?: boolean
  /** 自定義內容插槽 */
  content?: any
  /** 操作按鈕插槽 */
  actions?: any
}

const props = withDefaults(defineProps<Props>(), {
  isFirst: false,
  dragHandleIcon: 'i-lucide-grip-vertical',
  showExpand: false,
  expanded: false
})

const emit = defineEmits<{
  'toggle-expand': []
  'edit': []
  'delete': []
}>()
</script>

<template>
  <div class="flex items-center gap-3 p-3 border border-gray-200 dark:border-gray-700 rounded-lg hover:bg-white dark:hover:bg-gray-700/50 transition-colors bg-white dark:bg-gray-900/50">
    <!-- 展開/收起按鈕 -->
    <BaseIcon
      v-if="showExpand"
      :name="expanded ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right'"
      class="w-5 h-5 flex-shrink-0 cursor-pointer"
      @click="emit('toggle-expand')"
    />
    <div v-else class="w-5 flex-shrink-0" />

    <!-- 拖曳手柄 -->
    <BaseIcon
      :name="dragHandleIcon"
      class="w-5 h-5 text-gray-400 drag-handle cursor-grab flex-shrink-0"
      @click.stop
    />

    <!-- 內容插槽 -->
    <div class="flex-1 min-w-0">
      <slot name="content" :item="item" />
    </div>

    <!-- 操作按鈕插槽 -->
    <div class="flex items-center gap-1 flex-shrink-0 ml-2" @click.stop>
      <slot name="actions" :item="item">
        <BaseButton
          icon="i-lucide-edit"
          variant="ghost"
          size="xs"
          @click="emit('edit')"
        />
        <BaseButton
          icon="i-lucide-trash-2"
          variant="ghost"
          size="xs"
          color="error"
          @click="emit('delete')"
        />
      </slot>
    </div>
  </div>
</template>

