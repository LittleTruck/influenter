<script setup lang="ts">
import type { ViewType } from '~/types/cases'
import { BaseButton } from '~/components/base'

interface Props {
  modelValue: ViewType
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: ViewType]
}>()

const handleViewChange = (view: ViewType) => {
  emit('update:modelValue', view)
}
</script>

<template>
  <div class="flex items-center gap-1 bg-gray-100 dark:bg-gray-800 p-1 rounded-lg">
    <BaseButton
      :color="props.modelValue === 'board' ? 'primary' : 'neutral'"
      :variant="props.modelValue === 'board' ? 'solid' : 'ghost'"
      size="xs"
      icon="i-lucide-layout-grid"
      @click="handleViewChange('board')"
    >
      看板
    </BaseButton>
    <BaseButton
      :color="props.modelValue === 'list' ? 'primary' : 'neutral'"
      :variant="props.modelValue === 'list' ? 'solid' : 'ghost'"
      size="xs"
      icon="i-lucide-list"
      @click="handleViewChange('list')"
    >
      列表
    </BaseButton>
  </div>
</template>



