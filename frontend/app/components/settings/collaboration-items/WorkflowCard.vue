<script setup lang="ts">
import type { WorkflowTemplate, CollaborationItemPhase, CreateCollaborationItemPhaseRequest, UpdateCollaborationItemPhaseRequest } from '~/types/collaborationItems'
import { WORKFLOW_COLORS } from '~/utils/mockData'
import { BaseButton, BaseIcon, BaseBadge, BaseCollapsible } from '~/components/base'
import PhaseFormModal from '~/components/collaboration-items/PhaseFormModal.vue'
import DraggableList from '~/components/base/DraggableList.vue'
import DraggableItemCard from './DraggableItemCard.vue'

interface Props {
  workflow: WorkflowTemplate
  expanded?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  expanded: false
})

const emit = defineEmits<{
  'toggle-expand': [workflowId: string]
  'edit-workflow': [workflow: WorkflowTemplate]
  'delete-workflow': [workflow: WorkflowTemplate]
  'add-phase': [workflowId: string]
  'edit-phase': [phase: CollaborationItemPhase]
  'delete-phase': [phase: CollaborationItemPhase]
  'reorder-phases': [workflowId: string, phases: CollaborationItemPhase[]]
}>()

const isExpanded = ref(props.expanded)
const localPhases = ref<CollaborationItemPhase[]>([...props.workflow.phases])

// 同步外部 phases 變化
watch(() => props.workflow.phases, (newPhases) => {
  localPhases.value = [...newPhases]
}, { deep: true, immediate: true })

// 監聽 isExpanded 變化，同步到父組件
watch(() => isExpanded.value, (newValue) => {
  emit('toggle-expand', props.workflow.id)
})

// 處理階段拖曳重排序
const handleReorderPhases = (phaseIds: string[]) => {
  // 根據新的順序重新排列 phases
  const reorderedPhases = phaseIds.map((id, index) => {
    const phase = localPhases.value.find(p => p.id === id)
    if (phase) {
      return { ...phase, order: index }
    }
    return null
  }).filter(Boolean) as CollaborationItemPhase[]
  
  localPhases.value = reorderedPhases
  emit('reorder-phases', props.workflow.id, localPhases.value)
}

// 取得顏色顯示類
const getColorClass = (color: string) => {
  const colorOption = WORKFLOW_COLORS.find(c => c.value === color)
  return colorOption?.class || 'bg-gray-500'
}

// 階段表單狀態
const showPhaseForm = ref(false)
const editingPhase = ref<CollaborationItemPhase | null>(null)

const handleAddPhase = () => {
  editingPhase.value = null
  showPhaseForm.value = true
}

const handleEditPhase = (phase: CollaborationItemPhase) => {
  editingPhase.value = phase
  showPhaseForm.value = true
}

const handleDeletePhase = (phase: CollaborationItemPhase) => {
  emit('delete-phase', phase)
}

const handlePhaseSubmit = (data: CreateCollaborationItemPhaseRequest | UpdateCollaborationItemPhaseRequest) => {
  if (editingPhase.value) {
    // 更新階段
    const index = localPhases.value.findIndex(p => p.id === editingPhase.value?.id)
    if (index !== -1) {
      localPhases.value[index] = {
        ...localPhases.value[index],
        name: data.name || localPhases.value[index].name,
        duration_days: data.duration_days || localPhases.value[index].duration_days,
        updated_at: new Date().toISOString()
      }
      emit('reorder-phases', props.workflow.id, localPhases.value)
    }
  } else {
    // 新增階段
    const newPhase: CollaborationItemPhase = {
      id: `phase_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      workflow_template_id: props.workflow.id,
      name: data.name!,
      duration_days: data.duration_days!,
      order: localPhases.value.length,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    }
    localPhases.value.push(newPhase)
    emit('reorder-phases', props.workflow.id, localPhases.value)
  }
  showPhaseForm.value = false
  editingPhase.value = null
}
</script>

<template>
  <div class="workflow-card">
    <BaseCollapsible
      v-model:open="isExpanded"
      class="collapsible-item"
      :ui="{ content: 'pb-0 mb-0' }"
    >
      <!-- 卡片主體 -->
      <div class="flex items-center gap-3 p-3 border border-gray-200 dark:border-gray-700 rounded-lg hover:bg-white dark:hover:bg-gray-700/50 transition-colors cursor-pointer bg-white dark:bg-gray-900/50">
        <!-- 展開/收起按鈕 -->
        <BaseIcon
          v-if="localPhases.length > 0"
          :name="isExpanded ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right'"
          class="w-5 h-5 flex-shrink-0"
        />
        <div v-else class="w-5 flex-shrink-0" />

        <!-- 拖曳手柄 -->
        <BaseIcon
          name="i-lucide-grip-vertical"
          class="w-5 h-5 text-gray-400 drag-handle cursor-grab flex-shrink-0"
          @click.stop
        />

        <!-- 顏色標記 -->
        <div :class="['w-4 h-4 rounded-full flex-shrink-0', getColorClass(workflow.color)]" />

        <!-- 流程名稱 -->
        <div class="flex-1 min-w-0">
          <h4 class="font-medium text-gray-900 dark:text-white truncate">
            {{ workflow.name }}
          </h4>
        </div>

        <!-- 操作按鈕 -->
        <div class="flex items-center gap-1 flex-shrink-0 ml-2" @click.stop>
          <BaseButton
            icon="i-lucide-plus"
            variant="ghost"
            size="xs"
            @click="handleAddPhase"
          >
            階段
          </BaseButton>
          <BaseButton
            icon="i-lucide-edit"
            variant="ghost"
            size="xs"
            @click="emit('edit-workflow', workflow)"
          />
          <BaseButton
            icon="i-lucide-trash-2"
            variant="ghost"
            size="xs"
            color="error"
            @click="emit('delete-workflow', workflow)"
          />
        </div>
      </div>

      <!-- 階段列表（展開時顯示） -->
      <template #content>
        <div class="pl-12 py-2 bg-gray-50 dark:bg-gray-800/30">
          <DraggableList
            v-model:items="localPhases"
            group-name="phases"
            @reorder="handleReorderPhases"
          >
            <template #item="{ element: phase, index }">
              <DraggableItemCard
                :item="phase"
                :is-first="index === 0"
                @edit="handleEditPhase(phase)"
                @delete="handleDeletePhase(phase)"
              >
                <template #content="{ item: phase }">
                  <div class="flex items-center gap-2">
                    <h4 class="font-medium text-gray-900 dark:text-white truncate">{{ phase.name }}</h4>
                    <BaseBadge size="sm" variant="soft" color="neutral">
                      {{ phase.duration_days }} 天
                    </BaseBadge>
                  </div>
                </template>
              </DraggableItemCard>
            </template>
          </DraggableList>
        </div>
      </template>
    </BaseCollapsible>

    <!-- 階段表單 Modal -->
    <PhaseFormModal
      v-model="showPhaseForm"
      :phase="editingPhase"
      :collaboration-item-id="workflow.id"
      @submit="handlePhaseSubmit"
    />
  </div>
</template>

<style scoped>
/* 拖曳樣式已統一在 DraggableList 組件中 */
</style>

