<script setup lang="ts">
import type { CollaborationItemPhase } from '~/types/collaborationItems'
import { BaseButton, BaseIcon, BaseBadge } from '~/components/base'
import draggable from 'vuedraggable'

interface Props {
  /** 階段列表 */
  phases: CollaborationItemPhase[]
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
  'add-phase': []
  'edit-phase': [phase: CollaborationItemPhase]
  'delete-phase': [phase: CollaborationItemPhase]
  'reorder': [phases: CollaborationItemPhase[]]
}>()

// 拖曳排序後的階段列表
const sortedPhases = ref<CollaborationItemPhase[]>([...props.phases])

// 監聽 props.phases 變化
watch(() => props.phases, (newPhases) => {
  sortedPhases.value = [...newPhases]
}, { deep: true })

// 處理拖曳結束
const handleDragEnd = () => {
  // 更新 order
  const updatedPhases = sortedPhases.value.map((phase, index) => ({
    ...phase,
    order: index
  }))
  emit('reorder', updatedPhases)
}

// 處理編輯
const handleEdit = (phase: CollaborationItemPhase) => {
  emit('edit-phase', phase)
}

// 處理刪除
const handleDelete = (phase: CollaborationItemPhase) => {
  emit('delete-phase', phase)
}

// 處理新增
const handleAdd = () => {
  emit('add-phase')
}
</script>

<template>
  <div class="phase-list">
    <!-- 標題和新增按鈕 -->
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg font-semibold text-highlighted">
        階段流程
      </h3>
      <BaseButton
        v-if="editable"
        icon="i-lucide-plus"
        size="sm"
        @click="handleAdd"
      >
        新增階段
      </BaseButton>
    </div>

    <!-- 階段列表 -->
    <div v-if="loading" class="flex items-center justify-center py-8">
      <BaseIcon name="i-lucide-loader-2" class="w-6 h-6 animate-spin text-primary" />
    </div>

    <div v-else-if="sortedPhases.length === 0" class="text-center py-8 text-muted">
      <BaseIcon name="i-lucide-list-x" class="w-12 h-12 mx-auto mb-2 opacity-50" />
      <p class="mb-1">尚未設定階段流程</p>
      <p class="text-xs mb-4">階段流程定義了此合作項目的預設流程，案件可以套用此流程</p>
      <BaseButton
        v-if="editable"
        icon="i-lucide-plus"
        variant="outline"
        size="sm"
        class="mt-4"
        @click="handleAdd"
      >
        新增第一個階段流程
      </BaseButton>
    </div>

    <!-- 階段列表 -->
    <div v-else class="space-y-3">
      <div
        v-for="(phase, index) in sortedPhases"
        :key="phase.id"
        class="flex items-center gap-4 p-3 rounded-lg border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800/50"
      >
        <!-- 順序編號 -->
        <div class="w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center text-primary font-semibold flex-shrink-0">
          {{ index + 1 }}
        </div>

        <!-- 階段資訊 -->
        <div class="flex-1 min-w-0">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <h4 class="font-medium text-highlighted">{{ phase.name }}</h4>
              <BaseBadge size="xs" variant="subtle" color="neutral">
                {{ phase.duration_days }} 天
              </BaseBadge>
            </div>
            <div v-if="editable" class="flex items-center gap-2">
              <BaseButton
                icon="i-lucide-edit"
                variant="ghost"
                size="xs"
                @click="handleEdit(phase)"
              />
              <BaseButton
                icon="i-lucide-trash-2"
                variant="ghost"
                size="xs"
                color="error"
                @click="handleDelete(phase)"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.phase-list {
  min-height: 200px;
}

.phase-timeline {
  @apply w-full;
}
</style>

