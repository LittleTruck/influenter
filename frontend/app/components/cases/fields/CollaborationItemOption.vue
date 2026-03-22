<script setup lang="ts">
import type { CollaborationItem } from '~/types/collaborationItems'
import { formatAmount } from '~/utils/formatters'

interface Props {
  item: CollaborationItem
  selectedIds: string[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'toggle': [id: string]
}>()

const isSelected = computed(() => props.selectedIds.includes(props.item.id))

// 取得 bundle 內含項目名稱
const bundleItemNames = computed(() => {
  if (props.item.type !== 'bundle' || !props.item.bundle_items) return []
  return props.item.bundle_items
    .sort((a, b) => a.order - b.order)
    .map(ref => ref.item?.title || '未知項目')
})

const handleToggle = () => {
  emit('toggle', props.item.id)
}
</script>

<template>
  <div class="collaboration-item-option">
    <div
      :class="[
        'flex items-center gap-2 p-2 rounded hover:bg-subtle cursor-pointer',
        isSelected && 'bg-primary-50 dark:bg-primary-900/20'
      ]"
      @click="handleToggle"
    >
      <input
        type="checkbox"
        :checked="isSelected"
        class="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
        @click.stop="handleToggle"
      />
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2">
          <span class="text-sm text-highlighted truncate">
            {{ item.title }}
          </span>
          <span
            v-if="item.type === 'bundle'"
            class="text-xs px-1.5 py-0.5 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded flex-shrink-0"
          >
            組合
          </span>
        </div>
        <!-- Bundle 內含項目名稱 -->
        <p v-if="bundleItemNames.length > 0" class="text-xs text-muted mt-0.5 truncate">
          包含：{{ bundleItemNames.join('、') }}
        </p>
      </div>
      <span class="text-xs text-muted flex-shrink-0">
        {{ formatAmount(item.price) }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.collaboration-item-option {
  transition: background-color 0.2s;
}
</style>
