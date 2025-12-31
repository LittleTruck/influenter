<script setup lang="ts">
import type { CasePhase } from '~/types/cases'
import type { TimelineItem } from '@nuxt/ui'
import { BaseButton, BaseBadge, BaseIcon } from '~/components/base'
import { format } from 'date-fns'
import { zhTW } from 'date-fns/locale'

interface Props {
  /** 案件階段列表 */
  phases: CasePhase[]
  /** 是否可編輯 */
  editable?: boolean
  /** 載入狀態 */
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  editable: true,
  loading: false
})

const emit = defineEmits<{
  'edit-phase': [phase: CasePhase]
  'delete-phase': [phase: CasePhase]
  'add-phase': []
}>()

// 計算階段狀態
const getPhaseStatus = (phase: CasePhase): 'completed' | 'in_progress' | 'pending' => {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  
  const startDate = new Date(phase.start_date)
  startDate.setHours(0, 0, 0, 0)
  
  const endDate = new Date(phase.end_date)
  endDate.setHours(23, 59, 59, 999)

  if (today > endDate) {
    return 'completed'
  } else if (today >= startDate && today <= endDate) {
    return 'in_progress'
  } else {
    return 'pending'
  }
}

// 取得階段狀態顏色
const getPhaseColor = (status: 'completed' | 'in_progress' | 'pending'): 'success' | 'primary' | 'neutral' => {
  switch (status) {
    case 'completed':
      return 'success'
    case 'in_progress':
      return 'primary'
    default:
      return 'neutral'
  }
}

// 取得階段狀態圖示
const getPhaseIcon = (status: 'completed' | 'in_progress' | 'pending'): string => {
  switch (status) {
    case 'completed':
      return 'i-lucide-check-circle'
    case 'in_progress':
      return 'i-lucide-play-circle'
    default:
      return 'i-lucide-circle'
  }
}

// 取得階段狀態標籤
const getPhaseStatusLabel = (status: 'completed' | 'in_progress' | 'pending'): string => {
  switch (status) {
    case 'completed':
      return '已完成'
    case 'in_progress':
      return '進行中'
    default:
      return '待開始'
  }
}

// 格式化日期
const formatDate = (dateStr: string): string => {
  return format(new Date(dateStr), 'yyyy-MM-dd', { locale: zhTW })
}

// 處理編輯
const handleEdit = (phase: CasePhase) => {
  emit('edit-phase', phase)
}

// 處理刪除
const handleDelete = (phase: CasePhase) => {
  emit('delete-phase', phase)
}

// 處理新增
const handleAdd = () => {
  emit('add-phase')
}

// 排序後的階段列表
const sortedPhases = computed(() => {
  return [...props.phases].sort((a, b) => a.order - b.order)
})

// 轉換為 Timeline items
const timelineItems = computed<TimelineItem[]>(() => {
  return sortedPhases.value.map((phase) => {
    const status = getPhaseStatus(phase)
    return {
      date: `${formatDate(phase.start_date)} ~ ${formatDate(phase.end_date)} (${phase.duration_days} 天)`,
      title: phase.name,
      icon: getPhaseIcon(status),
      value: phase.id,
      // 將原始 phase 數據和狀態存儲在自定義屬性中
      _phase: phase,
      _status: status,
      _statusColor: getPhaseColor(status),
      _statusLabel: getPhaseStatusLabel(status)
    }
  })
})

// 計算當前活動的階段（進行中或第一個待開始的階段）
const activePhaseIndex = computed(() => {
  const inProgressIndex = sortedPhases.value.findIndex(phase => getPhaseStatus(phase) === 'in_progress')
  if (inProgressIndex !== -1) {
    return inProgressIndex
  }
  const pendingIndex = sortedPhases.value.findIndex(phase => getPhaseStatus(phase) === 'pending')
  if (pendingIndex !== -1) {
    return pendingIndex
  }
  // 如果都沒有，返回最後一個（已完成）
  return sortedPhases.value.length > 0 ? sortedPhases.value.length - 1 : undefined
})

// 根據階段狀態決定 Timeline 顏色
const timelineColor = computed<'primary' | 'success' | 'neutral'>(() => {
  if (sortedPhases.value.length === 0) return 'primary'
  const status = getPhaseStatus(sortedPhases.value[0])
  return getPhaseColor(status)
})
</script>

<template>
  <div class="case-phase-timeline">
    <!-- 載入狀態 -->
    <div v-if="loading" class="flex items-center justify-center py-8">
      <BaseIcon name="i-lucide-loader-2" class="w-6 h-6 animate-spin text-primary" />
    </div>

    <!-- Timeline -->
    <UTimeline
      v-else-if="sortedPhases.length > 0"
      :items="timelineItems"
      :default-value="activePhaseIndex"
      :color="timelineColor"
    >
      <template #title="{ item }">
        <div class="flex items-center gap-2">
          <span class="font-medium text-highlighted">{{ item.title }}</span>
          <BaseBadge
            :color="(item as any)._statusColor"
            size="xs"
            variant="subtle"
          >
            {{ (item as any)._statusLabel }}
          </BaseBadge>
        </div>
      </template>

      <template #date="{ item }">
        <div class="flex items-center justify-between gap-4 flex-shrink-0">
          <span class="text-xs text-dimmed whitespace-nowrap">{{ item.date }}</span>
          <div v-if="editable" class="flex items-center gap-2">
            <BaseButton
              icon="i-lucide-edit"
              variant="ghost"
              size="xs"
              @click.stop="handleEdit((item as any)._phase)"
            >
              編輯日期
            </BaseButton>
            <BaseButton
              icon="i-lucide-trash-2"
              variant="ghost"
              size="xs"
              color="error"
              @click.stop="handleDelete((item as any)._phase)"
            />
          </div>
        </div>
      </template>
    </UTimeline>

    <!-- 空狀態 -->
    <div v-else class="text-center py-8 text-muted">
      <BaseIcon name="i-lucide-list-x" class="w-12 h-12 mx-auto mb-2 opacity-50" />
      <p class="mb-1">尚未設定流程階段</p>
      <p class="text-xs">請點擊「套用流程」按鈕，從合作項目套用階段流程</p>
    </div>
  </div>
</template>

<style scoped>
.case-phase-timeline {
  min-height: 200px;
  padding: 1rem 0;
}
</style>
