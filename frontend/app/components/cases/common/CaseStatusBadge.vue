<script setup lang="ts">
import type { CaseStatus } from '~/types/cases'
import { getStatusColor, getStatusLabel } from '~/utils/caseStatus'

interface Props {
  /** 案件狀態 */
  status: CaseStatus
  /** 是否顯示圖示 */
  showIcon?: boolean
  /** 尺寸 */
  size?: 'xs' | 'sm' | 'md'
}

const props = withDefaults(defineProps<Props>(), {
  showIcon: false,
  size: 'sm'
})

const color = computed(() => getStatusColor(props.status))
const label = computed(() => getStatusLabel(props.status))

const statusIcon = computed(() => {
  const icons = {
    to_confirm: 'i-lucide-clock',
    in_progress: 'i-lucide-play-circle',
    completed: 'i-lucide-check-circle',
    cancelled: 'i-lucide-x-circle'
  }
  return icons[props.status] || 'i-lucide-circle'
})
</script>

<template>
  <UBadge
    :color="color"
    variant="subtle"
    :size="size"
    :ui="{ base: 'inline-flex items-center gap-1' }"
  >
    <UIcon v-if="showIcon" :name="statusIcon" class="w-3 h-3" />
    {{ label }}
  </UBadge>
</template>

