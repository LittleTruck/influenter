<script setup lang="ts">
import type { CollaborationItem } from '~/types/collaborationItems'
import { formatAmount } from '~/utils/formatters'

interface Props {
  item: CollaborationItem
  level?: number
  selectedIds: string[]
}

const props = withDefaults(defineProps<Props>(), {
  level: 0
})

const emit = defineEmits<{
  'toggle': [id: string]
}>()

const isSelected = computed(() => props.selectedIds.includes(props.item.id))
const hasChildren = computed(() => props.item.children && props.item.children.length > 0)

const handleToggle = () => {
  emit('toggle', props.item.id)
}
</script>

<template>
  <div class="collaboration-item-option">
    <div
      :class="[
        'flex items-center gap-2 p-2 rounded hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer',
        isSelected && 'bg-primary-50 dark:bg-primary-900/20'
      ]"
      :style="{ paddingLeft: `${level * 1.5 + 0.5}rem` }"
      @click="handleToggle"
    >
      <input
        type="checkbox"
        :checked="isSelected"
        class="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
        @click.stop="handleToggle"
      />
      <span class="flex-1 text-sm text-gray-900 dark:text-white">
        {{ item.title }}
      </span>
      <span class="text-xs text-gray-500 dark:text-gray-400">
        {{ formatAmount(item.price) }}
      </span>
    </div>
    <CollaborationItemOption
      v-for="child in item.children"
      :key="child.id"
      :item="child"
      :level="level + 1"
      :selected-ids="selectedIds"
      @toggle="$emit('toggle', $event)"
    />
  </div>
</template>

<style scoped>
.collaboration-item-option {
  transition: background-color 0.2s;
}
</style>




