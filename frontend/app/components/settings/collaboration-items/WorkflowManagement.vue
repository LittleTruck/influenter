<script setup lang="ts">
import type { WorkflowTemplate, CollaborationItemPhase } from '~/types/collaborationItems'
import { useWorkflowTemplates } from '~/composables/useWorkflowTemplates'
import { useErrorHandler } from '~/composables/useErrorHandler'
import { BaseButton, BaseAlert } from '~/components/base'
import EmptyState from '~/components/common/EmptyState.vue'
import WorkflowFormModal from './WorkflowFormModal.vue'
import WorkflowTree from './WorkflowTree.vue'

const { handleError, handleSuccess } = useErrorHandler()
const { workflowTemplates: templateList, fetchWorkflows, createWorkflow, updateWorkflow, deleteWorkflow, reorderWorkflows, updatePhase } = useWorkflowTemplates()

// 使用本地狀態來管理流程（可以修改）
const workflows = ref<WorkflowTemplate[]>([])
const apiError = ref<string | null>(null)

// 同步 composable 的資料到本地狀態
watch(templateList, (newList) => {
  workflows.value = [...newList]
}, { immediate: true })

// 表單狀態
const showWorkflowForm = ref(false)
const editingWorkflow = ref<WorkflowTemplate | null>(null)

// 載入數據
onMounted(async () => {
  await fetchWorkflows()
})

// 處理新增流程
const handleAddWorkflow = () => {
  editingWorkflow.value = null
  showWorkflowForm.value = true
}

// 暴露方法給父組件
defineExpose({
  handleAddWorkflow
})

// 處理編輯流程
const handleEditWorkflow = (workflow: WorkflowTemplate) => {
  editingWorkflow.value = workflow
  showWorkflowForm.value = true
}

// 處理刪除流程
const handleDeleteWorkflow = async (workflow: WorkflowTemplate) => {
  try {
    await deleteWorkflow(workflow.id)
    workflows.value = workflows.value.filter(w => w.id !== workflow.id)
    handleSuccess('流程已刪除')
  } catch (error: unknown) {
    handleError(error, '刪除失敗')
  }
}

// 處理重新排序流程
const handleReorderWorkflows = async (workflowIds: string[]) => {
  try {
    await reorderWorkflows(workflowIds)
    // 重新排序本地列表
    const sortedWorkflows = workflowIds.map(id => workflows.value.find(w => w.id === id)).filter(Boolean) as WorkflowTemplate[]
    workflows.value = sortedWorkflows
    handleSuccess('排序已更新')
  } catch (error: unknown) {
    handleError(error, '排序失敗')
  }
}

// 處理重新排序階段
const handleReorderPhases = async (workflowId: string, phases: CollaborationItemPhase[]) => {
  try {
    const workflow = workflows.value.find(w => w.id === workflowId)
    if (workflow) {
      workflow.phases = phases
      // 逐一更新每個階段的 order
      await Promise.all(
        phases.map((phase, index) =>
          updatePhase(workflowId, phase.id, { order: index })
        )
      )
      handleSuccess('階段排序已更新')
    }
  } catch (error: unknown) {
    handleError(error, '排序失敗')
  }
}

// 處理流程表單提交
const handleWorkflowSubmit = async (data: { name: string; description?: string; color: string }) => {
  try {
    if (editingWorkflow.value) {
      // 更新流程
      await updateWorkflow(editingWorkflow.value.id, data)
      const index = workflows.value.findIndex(w => w.id === editingWorkflow.value?.id)
      if (index !== -1) {
        workflows.value[index] = {
          ...workflows.value[index],
          ...data,
          updated_at: new Date().toISOString()
        }
      }
      handleSuccess('流程已更新')
    } else {
      // 新增流程
      const newWorkflow = await createWorkflow(data)
      workflows.value.push(newWorkflow)
      handleSuccess('流程已新增')
    }
    showWorkflowForm.value = false
    editingWorkflow.value = null
  } catch (error: unknown) {
    handleError(error, '操作失敗')
  }
}
</script>

<template>
  <div>
    <!-- API 錯誤提示 -->
    <BaseAlert
      v-if="apiError"
      color="warning"
      variant="soft"
      :title="apiError"
      :close="true"
      class="mb-4"
      @close="apiError = null"
    />


    <!-- 空狀態 -->
    <EmptyState
      v-if="workflows.length === 0"
      icon="i-lucide-list-checks"
      title="還沒有流程"
      action-label="建立第一個流程"
      :show-icon-background="false"
      @action="handleAddWorkflow"
    />

    <!-- 流程列表 -->
    <WorkflowTree
      v-else
      :workflows="workflows"
      @add-workflow="handleAddWorkflow"
      @edit-workflow="handleEditWorkflow"
      @delete-workflow="handleDeleteWorkflow"
      @reorder-workflows="handleReorderWorkflows"
      @reorder-phases="handleReorderPhases"
    />

    <!-- 流程表單 Modal -->
    <WorkflowFormModal
      v-model="showWorkflowForm"
      :workflow="editingWorkflow"
      @submit="handleWorkflowSubmit"
    />
  </div>
</template>

