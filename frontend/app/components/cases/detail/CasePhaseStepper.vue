<script setup lang="ts">
import type { CasePhase } from '~/types/cases'
import { BaseIcon, BaseButton } from '~/components/base'
import { format } from 'date-fns'
import { zhTW } from 'date-fns/locale'

interface Props {
  /** 案件階段列表 */
  phases: CasePhase[]
  /** 是否可編輯 */
  editable?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  editable: false
})

const emit = defineEmits<{
  'edit-phase': [phase: CasePhase]
  'delete-phase': [phase: CasePhase]
}>()

// 排序後的階段列表
const sortedPhases = computed(() => {
  return [...props.phases].sort((a, b) => a.order - b.order)
})

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

// 格式化日期
const formatDate = (dateStr: string): string => {
  return format(new Date(dateStr), 'MM/dd', { locale: zhTW })
}

// 展開的 popover 項
const expandedPhaseId = ref<string | null>(null)

const togglePopover = (phaseId: string) => {
  expandedPhaseId.value = expandedPhaseId.value === phaseId ? null : phaseId
}

const closePopover = () => {
  expandedPhaseId.value = null
}
</script>

<template>
  <div class="case-phase-stepper">
    <!-- 空狀態 -->
    <div v-if="sortedPhases.length === 0" class="text-center py-6 text-muted">
      <BaseIcon name="i-lucide-git-branch" class="w-8 h-8 mx-auto mb-2 opacity-40" />
      <p class="text-sm">尚未設定流程階段</p>
    </div>

    <!-- 水平流程條 -->
    <div v-else class="flex items-start w-full overflow-x-auto pb-2">
      <template v-for="(phase, index) in sortedPhases" :key="phase.id">
        <!-- 階段圓圈 + 標籤 -->
        <div class="flex flex-col items-center flex-shrink-0 relative" style="min-width: 100px;">
          <!-- 圓圈 -->
          <button
            :class="[
              'w-10 h-10 rounded-full flex items-center justify-center text-sm font-semibold transition-all relative z-10',
              getPhaseStatus(phase) === 'completed'
                ? 'bg-green-500 text-white shadow-sm'
                : getPhaseStatus(phase) === 'in_progress'
                  ? 'bg-primary text-white shadow-md'
                  : 'bg-gray-100 dark:bg-gray-800 text-gray-400 dark:text-gray-500 border-2 border-gray-300 dark:border-gray-600'
            ]"
            @click="(getPhaseStatus(phase) !== 'pending' && editable) ? togglePopover(phase.id) : undefined"
          >
            <!-- 已完成：勾號 -->
            <BaseIcon
              v-if="getPhaseStatus(phase) === 'completed'"
              name="i-lucide-check"
              class="w-5 h-5"
            />
            <!-- 進行中：pulse 動畫 -->
            <template v-else-if="getPhaseStatus(phase) === 'in_progress'">
              <span class="absolute inset-0 rounded-full bg-primary animate-ping opacity-20" />
              <BaseIcon name="i-lucide-play" class="w-4 h-4 relative z-10" />
            </template>
            <!-- 待開始：數字 -->
            <span v-else>{{ index + 1 }}</span>
          </button>

          <!-- 階段名稱 -->
          <span
            :class="[
              'mt-2 text-xs font-medium text-center leading-tight max-w-[90px] truncate',
              getPhaseStatus(phase) === 'completed'
                ? 'text-green-600 dark:text-green-400'
                : getPhaseStatus(phase) === 'in_progress'
                  ? 'text-primary font-semibold'
                  : 'text-dimmed'
            ]"
          >
            {{ phase.name }}
          </span>

          <!-- 日期範圍 -->
          <span class="mt-0.5 text-[10px] text-dimmed whitespace-nowrap">
            {{ formatDate(phase.start_date) }} ~ {{ formatDate(phase.end_date) }}
          </span>

          <!-- Popover 詳情 -->
          <Transition
            enter-active-class="transition duration-150 ease-out"
            enter-from-class="opacity-0 translate-y-1"
            enter-to-class="opacity-100 translate-y-0"
            leave-active-class="transition duration-100 ease-in"
            leave-from-class="opacity-100 translate-y-0"
            leave-to-class="opacity-0 translate-y-1"
          >
            <div
              v-if="expandedPhaseId === phase.id && editable"
              class="absolute top-full mt-2 z-20 bg-elevated border border-default rounded-lg shadow-lg p-3 min-w-[180px]"
            >
              <div class="text-sm space-y-2">
                <div>
                  <span class="text-dimmed text-xs">期間</span>
                  <p class="font-medium text-highlighted">
                    {{ format(new Date(phase.start_date), 'yyyy/MM/dd') }} ~ {{ format(new Date(phase.end_date), 'yyyy/MM/dd') }}
                  </p>
                </div>
                <div>
                  <span class="text-dimmed text-xs">天數</span>
                  <p class="font-medium text-highlighted">{{ phase.duration_days }} 天</p>
                </div>
                <div class="flex gap-2 pt-1 border-t border-default">
                  <BaseButton
                    icon="i-lucide-edit"
                    variant="ghost"
                    size="xs"
                    @click="emit('edit-phase', phase); closePopover()"
                  >
                    編輯
                  </BaseButton>
                  <BaseButton
                    icon="i-lucide-trash-2"
                    variant="ghost"
                    size="xs"
                    color="error"
                    @click="emit('delete-phase', phase); closePopover()"
                  >
                    刪除
                  </BaseButton>
                </div>
              </div>
            </div>
          </Transition>
        </div>

        <!-- 連接線段 -->
        <div
          v-if="index < sortedPhases.length - 1"
          class="flex items-center flex-1 min-w-[24px] mt-5"
        >
          <div
            :class="[
              'h-0.5 w-full rounded-full transition-colors',
              getPhaseStatus(phase) === 'completed'
                ? 'bg-green-500'
                : 'bg-gray-200 dark:bg-gray-700'
            ]"
          />
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.case-phase-stepper {
  padding: 0.5rem 0;
}

@keyframes ping {
  75%, 100% {
    transform: scale(1.5);
    opacity: 0;
  }
}
</style>
