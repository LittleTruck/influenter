<script setup lang="ts">
import type { Task } from '~/types/cases'
import draggable from 'vuedraggable'
import { useCases } from '~/composables/useCases'
import { useErrorHandler } from '~/composables/useErrorHandler'
import TaskFormModal from '~/components/cases/forms/TaskFormModal.vue'
import LoadingState from '~/components/common/LoadingState.vue'
import EmptyState from '~/components/common/EmptyState.vue'
import { format } from 'date-fns'

interface Props {
  /** 案件 ID */
  caseId: string
  /** 任務列表 */
  tasks: Task[]
  /** 是否載入中 */
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false
})

const emit = defineEmits<{
  'task-update': []
}>()

const { completeTask, deleteTask, reorderTasks } = useCases()
const { handleError, handleSuccess } = useErrorHandler()

const showTaskForm = ref(false)
const editingTask = ref<Task | null>(null)

// 本地任務列表（用於拖曳）
const localTasks = computed({
  get: () => props.tasks,
  set: async (newTasks) => {
    // 拖曳後更新順序
    const taskIds = newTasks.map(t => t.id)
    try {
      await reorderTasks(props.caseId, taskIds)
      emit('task-update')
    } catch (error: any) {
      handleError(error, '重新排序失敗', { showToast: true, log: false })
    }
  }
})

// 拖曳選項
const dragOptions = {
  animation: 200,
  easing: 'cubic-bezier(0.25, 0.8, 0.25, 1)',
  handle: '.drag-handle',
  ghostClass: 'sortable-ghost',
  chosenClass: 'sortable-chosen',
  dragClass: 'sortable-drag'
}

// 任務統計
const taskStats = computed(() => {
  const total = props.tasks.length
  const completed = props.tasks.filter(t => t.status === 'completed').length
  const percentage = total > 0 ? Math.round((completed / total) * 100) : 0
  return { total, completed, percentage }
})

// 圓形進度計算
const circumference = 2 * Math.PI * 20
const progressOffset = computed(() => {
  return circumference - (taskStats.value.percentage / 100) * circumference
})

// 進度顏色
const progressColorClass = computed(() => {
  const pct = taskStats.value.percentage
  if (pct === 100) return 'text-green-500'
  if (pct >= 75) return 'text-blue-500'
  if (pct >= 50) return 'text-yellow-500'
  return 'text-red-500'
})

// 處理完成任務
const handleCompleteTask = async (taskId: string) => {
  try {
    await completeTask(taskId)
    emit('task-update')
    handleSuccess('任務已完成')
  } catch (error: any) {
    handleError(error, '完成任務失敗')
  }
}

// 處理刪除任務
const handleDeleteTask = async (taskId: string) => {
  try {
    await deleteTask(taskId)
    emit('task-update')
    handleSuccess('任務已刪除')
  } catch (error: any) {
    handleError(error, '刪除任務失敗')
  }
}

// 處理編輯任務
const handleEditTask = (task: Task) => {
  editingTask.value = task
  showTaskForm.value = true
}

// 處理新增任務
const handleAddTask = () => {
  editingTask.value = null
  showTaskForm.value = true
}

// 處理表單提交
const handleFormSubmit = () => {
  emit('task-update')
}

// 格式化日期
const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  return format(new Date(dateStr), 'yyyy/MM/dd')
}
</script>

<template>
  <div class="case-tasks-list">
    <!-- 進度視覺化 -->
    <div class="task-progress flex items-center gap-3 mb-6 p-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
      <!-- 圓形進度 -->
      <div class="relative w-12 h-12 flex-shrink-0">
        <svg class="transform -rotate-90 w-12 h-12">
          <circle
            cx="24"
            cy="24"
            r="20"
            stroke="currentColor"
            stroke-width="4"
            fill="none"
            class="text-gray-200 dark:text-gray-700"
          />
          <circle
            cx="24"
            cy="24"
            r="20"
            stroke="currentColor"
            stroke-width="4"
            fill="none"
            :stroke-dasharray="circumference"
            :stroke-dashoffset="progressOffset"
            :class="['transition-all duration-500', progressColorClass]"
          />
        </svg>
        <div class="absolute inset-0 flex items-center justify-center">
          <span class="text-xs font-semibold">{{ taskStats.percentage }}%</span>
        </div>
      </div>

      <div class="flex-1">
        <div class="text-sm font-medium text-gray-900 dark:text-white">
          {{ taskStats.completed }}/{{ taskStats.total }} 任務完成
        </div>
        <div class="text-xs text-gray-500 dark:text-gray-400">進度追蹤</div>
      </div>

      <UButton
        icon="i-lucide-plus"
        size="sm"
        @click="handleAddTask"
      >
        新增任務
      </UButton>
    </div>

    <!-- 任務列表 -->
    <div v-if="loading" class="flex items-center justify-center py-8">
      <UIcon name="i-lucide-loader-2" class="w-6 h-6 animate-spin text-primary-500" />
    </div>

    <EmptyState
      v-else-if="tasks.length === 0"
      icon="i-lucide-check-square"
      title="還沒有任務"
      :show-icon-background="false"
    />

    <draggable
      v-else
      v-model="localTasks"
      v-bind="dragOptions"
      item-key="id"
      class="space-y-2"
    >
      <template #item="{ element: task }">
        <div
          :class="[
            'task-item flex items-center gap-3 p-3 rounded-lg border transition-all',
            task.status === 'completed'
              ? 'bg-gray-50 dark:bg-gray-800/30 border-gray-200 dark:border-gray-700'
              : 'bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700 hover:shadow-md'
          ]"
        >
          <!-- 拖曳把手 -->
          <UIcon
            name="i-lucide-grip-vertical"
            class="drag-handle w-5 h-5 text-gray-400 cursor-grab active:cursor-grabbing flex-shrink-0"
          />

          <!-- 完成狀態 -->
          <UCheckbox
            :model-value="task.status === 'completed'"
            @update:model-value="
              task.status === 'completed' ? null : handleCompleteTask(task.id)
            "
            class="flex-shrink-0"
          />

          <!-- 任務內容 -->
          <div class="flex-1 min-w-0">
            <div
              :class="[
                'text-sm font-medium',
                task.status === 'completed'
                  ? 'text-gray-500 dark:text-gray-400 line-through'
                  : 'text-gray-900 dark:text-white'
              ]"
            >
              {{ task.title }}
            </div>
            <div v-if="task.description" class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              {{ task.description }}
            </div>
            <div v-if="task.due_date" class="text-xs text-gray-500 dark:text-gray-400 mt-1 flex items-center gap-1">
              <UIcon name="i-lucide-calendar" class="w-3 h-3" />
              {{ formatDate(task.due_date) }}
            </div>
          </div>

          <!-- 操作按鈕 -->
          <div class="flex items-center gap-1 flex-shrink-0">
            <UButton
              icon="i-lucide-edit"
              variant="ghost"
              size="xs"
              @click="handleEditTask(task)"
            />
            <UButton
              icon="i-lucide-trash-2"
              variant="ghost"
              size="xs"
              color="error"
              @click="handleDeleteTask(task.id)"
            />
          </div>
        </div>
      </template>
    </draggable>

    <!-- 任務表單 Modal -->
    <TaskFormModal
      v-model="showTaskForm"
      :case-id="caseId"
      :task="editingTask"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<style scoped>
.task-item {
  will-change: transform;
}

:deep(.sortable-ghost) {
  opacity: 0.5;
  transform: rotate(2deg);
}

:deep(.sortable-chosen) {
  transform: scale(1.02);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 1000;
}

:deep(.sortable-drag) {
  cursor: grabbing !important;
  z-index: 1000;
}
</style>

