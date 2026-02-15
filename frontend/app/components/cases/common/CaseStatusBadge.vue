<script setup lang="ts">
import type { CaseStatus } from '~/types/cases'
import { getStatusColor, getStatusLabel } from '~/utils/caseStatus'
import { BaseBadge, BaseIcon } from '~/components/base'

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
    cancelled: 'i-lucide-x-circle',
    other: 'i-lucide-file-question'
  }
  return icons[props.status] || 'i-lucide-circle'
})
</script>

<template>
  <BaseBadge
    :color="color"
    variant="subtle"
    :size="size"
    :class="'inline-flex items-center gap-1'"
  >
    <BaseIcon v-if="showIcon" :name="statusIcon" class="w-3 h-3" />
    {{ label }}
  </BaseBadge>
</template>

