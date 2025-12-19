<script setup lang="ts">
import type { Case, CaseStatus } from '~/types/cases'
import type { DragChangeEvent } from '~/types/dragEvents'
import draggable from 'vuedraggable'
import CaseBoardCard from './CaseBoardCard.vue'
import EmptyState from '~/components/common/EmptyState.vue'

interface Props {
  /** 狀態 */
  status: CaseStatus
  /** 狀態標籤 */
  label: string
  /** 狀態顏色 */
  color?: string
  /** 案件列表 */
  cases: Case[]
  /** 是否為拖曳目標 */
  isOver?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  color: '#3b82f6',
  isOver: false
})

const emit = defineEmits<{
  'update:cases': [cases: Case[]]
  'card-click': [caseId: string]
  'card-move': [caseId: string, newStatus: CaseStatus]
}>()

// 本地案件列表（用於 v-model）
const localCases = computed({
  get: () => props.cases,
  set: (value) => emit('update:cases', value)
})

// 拖曳選項
const dragOptions = {
  animation: 200,
  easing: 'cubic-bezier(0.25, 0.8, 0.25, 1)',
  group: 'case-cards',
  disabled: false,
  ghostClass: 'sortable-ghost',
  chosenClass: 'sortable-chosen',
  dragClass: 'sortable-drag',
  swapThreshold: 0.65,
  invertSwap: true,
  delay: 0,
  delayOnTouchOnly: false
}

// 處理卡片移動
const handleChange = (evt: DragChangeEvent<Case>) => {
  if (evt.added) {
    // 卡片被添加到這個欄位
    const caseId = evt.added.element.id
    emit('card-move', caseId, props.status)
  }
}

const handleCardClick = (caseId: string) => {
  emit('card-click', caseId)
}
</script>

<template>
  <div
    :data-column-status="status"
    :class="[
      'case-board-column flex flex-col min-w-[25rem] w-[25rem] flex-shrink-0 h-full px-2 transition-all duration-300',
      isOver && 'column-drag-over'
    ]"
  >
    <div
      class="flex flex-col h-full bg-gray-50 dark:bg-gray-800/30 rounded-xl p-4 transition-all duration-300"
    >
      <!-- Header -->
      <div class="flex items-center gap-2 mb-3 pb-3 border-b border-gray-200 dark:border-gray-700">
        <div
          class="h-2 w-2 rounded-full column-indicator flex-shrink-0"
          :style="{ backgroundColor: color }"
        />
        <span class="font-semibold text-sm text-gray-700 dark:text-gray-300">
          {{ label }}
        </span>
        <UBadge color="neutral" variant="subtle" size="xs" class="ml-auto">
          {{ cases.length }}
        </UBadge>
      </div>

      <!-- Cards 區域 -->
      <div class="flex-1 overflow-y-auto overflow-x-visible">
        <draggable
          v-model="localCases"
          v-bind="dragOptions"
          :group="{ name: 'case-cards', pull: true, put: true }"
          item-key="id"
          class="case-cards-container"
          @change="handleChange"
        >
          <template #item="{ element }">
            <div class="case-card-item mb-2">
              <CaseBoardCard :case-data="element" @card-click="handleCardClick" />
            </div>
          </template>
        </draggable>

        <!-- 空狀態 -->
        <EmptyState
          v-if="localCases.length === 0"
          icon="i-lucide-inbox"
          title="暫無案件"
          :show-icon-background="false"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.case-cards-container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-height: 100px;
}

.case-card-item {
  cursor: grab;
  transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
}

.case-card-item:active {
  cursor: grabbing;
}


/* 拖曳目標高亮效果 */
.column-drag-over > div {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.1), rgba(147, 51, 234, 0.1));
  border: 2px dashed rgba(59, 130, 246, 0.5);
  animation: pulse-border 1.5s ease-in-out infinite;
}

@keyframes pulse-border {
  0%,
  100% {
    border-color: rgba(59, 130, 246, 0.5);
  }
  50% {
    border-color: rgba(59, 130, 246, 0.8);
  }
}

/* 指示器動畫 */
.column-indicator {
  transition: transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.column-drag-over .column-indicator {
  transform: scale(1.4);
  animation: pulse-indicator 1s ease-in-out infinite;
}

@keyframes pulse-indicator {
  0%,
  100% {
    opacity: 1;
    transform: scale(1.4);
  }
  50% {
    opacity: 0.6;
    transform: scale(1.6);
  }
}

/* 拖曳時的視覺效果 */
:deep(.sortable-ghost) {
  opacity: 0.5;
  transform: rotate(2deg);
}

:deep(.sortable-chosen) {
  transform: scale(1.02);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.15);
  z-index: 1000;
}

:deep(.sortable-drag) {
  cursor: grabbing !important;
  z-index: 1000;
}
</style>

