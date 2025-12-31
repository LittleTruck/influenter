<script setup lang="ts">
import type { WorkflowTemplate, CollaborationItemPhase } from '~/types/collaborationItems'
import WorkflowCard from './WorkflowCard.vue'
import DraggableList from '~/components/base/DraggableList.vue'

interface Props {
  /** 流程列表 */
  workflows: WorkflowTemplate[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'add-workflow': []
  'edit-workflow': [workflow: WorkflowTemplate]
  'delete-workflow': [workflow: WorkflowTemplate]
  'add-phase': [workflowId: string]
  'edit-phase': [phase: CollaborationItemPhase]
  'delete-phase': [phase: CollaborationItemPhase]
  'reorder-workflows': [workflowIds: string[]]
  'reorder-phases': [workflowId: string, phases: CollaborationItemPhase[]]
}>()

// 展開狀態追蹤
const expandedWorkflows = ref<Record<string, boolean>>({})

// 本地流程列表（用於拖曳）
const localWorkflows = ref([...props.workflows])

// 同步外部 workflows 變化
watch(() => props.workflows, (newWorkflows) => {
  localWorkflows.value = [...newWorkflows]
}, { deep: true, immediate: true })

// 處理拖曳重排序
const handleReorder = (workflowIds: string[]) => {
  emit('reorder-workflows', workflowIds)
}

// 處理展開/收起
const handleToggleExpand = (workflowId: string) => {
  expandedWorkflows.value[workflowId] = !expandedWorkflows.value[workflowId]
}
</script>

<template>
  <div class="workflow-tree">
    <DraggableList
      v-model:items="localWorkflows"
      group-name="workflows"
      @reorder="handleReorder"
    >
      <template #item="{ element }">
        <WorkflowCard
          :workflow="element"
          :expanded="expandedWorkflows[element.id] || false"
          @toggle-expand="handleToggleExpand"
          @edit-workflow="emit('edit-workflow', $event)"
          @delete-workflow="emit('delete-workflow', $event)"
          @add-phase="emit('add-phase', $event)"
          @edit-phase="emit('edit-phase', $event)"
          @delete-phase="emit('delete-phase', $event)"
          @reorder-phases="emit('reorder-phases', $event[0], $event[1])"
        />
      </template>
    </DraggableList>
  </div>
</template>

<style scoped>
.workflow-tree {
  min-height: 100px;
}
</style>


