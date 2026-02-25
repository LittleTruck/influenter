/**
 * 流程範本相關 composable
 * 提供便捷的流程範本操作方法
 */

import type { WorkflowTemplate, CollaborationItemPhase } from '~/types/collaborationItems'

export const useWorkflowTemplates = () => {
  const workflowTemplates = ref<WorkflowTemplate[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const config = useRuntimeConfig()
  const authStore = useAuthStore()

  const headers = computed(() => ({
    Authorization: `Bearer ${authStore.token}`
  }))

  /**
   * 載入流程範本列表
   */
  const fetchWorkflows = async () => {
    loading.value = true
    error.value = null

    try {
      const data = await $fetch<{ data: WorkflowTemplate[] }>(
        `${config.public.apiBase}/api/v1/workflow-templates`,
        {
          headers: headers.value
        }
      )
      workflowTemplates.value = data.data ?? []
    } catch (e: any) {
      const is404 = e?.statusCode === 404 || e?.status === 404 || e?.response?.status === 404
      if (!is404) {
        error.value = '載入流程範本失敗'
        console.error('載入流程範本失敗:', e)
      }
    } finally {
      loading.value = false
    }
  }

  /**
   * 根據 ID 查找流程範本
   */
  const findWorkflowById = (id: string): WorkflowTemplate | null => {
    return workflowTemplates.value.find(w => w.id === id) || null
  }

  /**
   * 建立流程範本
   */
  const createWorkflow = async (data: { name: string; description?: string; color: string }): Promise<WorkflowTemplate> => {
    const newWorkflow = await $fetch<WorkflowTemplate>(
      `${config.public.apiBase}/api/v1/workflow-templates`,
      {
        method: 'POST',
        body: data,
        headers: headers.value
      }
    )
    workflowTemplates.value.push(newWorkflow)
    return newWorkflow
  }

  /**
   * 更新流程範本
   */
  const updateWorkflow = async (id: string, data: { name?: string; description?: string; color?: string }): Promise<WorkflowTemplate> => {
    const updatedWorkflow = await $fetch<WorkflowTemplate>(
      `${config.public.apiBase}/api/v1/workflow-templates/${id}`,
      {
        method: 'PATCH',
        body: data,
        headers: headers.value
      }
    )
    const index = workflowTemplates.value.findIndex(w => w.id === id)
    if (index !== -1) {
      workflowTemplates.value[index] = updatedWorkflow
    }
    return updatedWorkflow
  }

  /**
   * 刪除流程範本
   */
  const deleteWorkflow = async (id: string): Promise<void> => {
    await $fetch(
      `${config.public.apiBase}/api/v1/workflow-templates/${id}`,
      {
        method: 'DELETE',
        headers: headers.value
      }
    )
    workflowTemplates.value = workflowTemplates.value.filter(w => w.id !== id)
  }

  /**
   * 重新排序流程範本
   */
  const reorderWorkflows = async (workflowIds: string[]): Promise<void> => {
    const sortedWorkflows = workflowIds.map(id => workflowTemplates.value.find(w => w.id === id)).filter(Boolean) as WorkflowTemplate[]
    sortedWorkflows.forEach((workflow, index) => {
      workflow.order = index
    })
    workflowTemplates.value = sortedWorkflows
  }

  /**
   * 建立流程階段
   */
  const createPhase = async (workflowId: string, data: { name: string; duration_days: number; order?: number }): Promise<CollaborationItemPhase> => {
    const newPhase = await $fetch<CollaborationItemPhase>(
      `${config.public.apiBase}/api/v1/workflow-templates/${workflowId}/phases`,
      {
        method: 'POST',
        body: data,
        headers: headers.value
      }
    )
    // 更新本地範本的階段列表
    const workflow = workflowTemplates.value.find(w => w.id === workflowId)
    if (workflow) {
      workflow.phases.push(newPhase)
    }
    return newPhase
  }

  /**
   * 更新流程階段
   */
  const updatePhase = async (workflowId: string, phaseId: string, data: { name?: string; duration_days?: number; order?: number }): Promise<CollaborationItemPhase> => {
    const updatedPhase = await $fetch<CollaborationItemPhase>(
      `${config.public.apiBase}/api/v1/workflow-templates/${workflowId}/phases/${phaseId}`,
      {
        method: 'PATCH',
        body: data,
        headers: headers.value
      }
    )
    const workflow = workflowTemplates.value.find(w => w.id === workflowId)
    if (workflow) {
      const index = workflow.phases.findIndex(p => p.id === phaseId)
      if (index !== -1) {
        workflow.phases[index] = updatedPhase
      }
    }
    return updatedPhase
  }

  /**
   * 刪除流程階段
   */
  const deletePhase = async (workflowId: string, phaseId: string): Promise<void> => {
    await $fetch(
      `${config.public.apiBase}/api/v1/workflow-templates/${workflowId}/phases/${phaseId}`,
      {
        method: 'DELETE',
        headers: headers.value
      }
    )
    const workflow = workflowTemplates.value.find(w => w.id === workflowId)
    if (workflow) {
      workflow.phases = workflow.phases.filter(p => p.id !== phaseId)
    }
  }

  return {
    workflowTemplates: computed(() => workflowTemplates.value),
    loading: computed(() => loading.value),
    error: computed(() => error.value),
    fetchWorkflows,
    findWorkflowById,
    createWorkflow,
    updateWorkflow,
    deleteWorkflow,
    reorderWorkflows,
    createPhase,
    updatePhase,
    deletePhase
  }
}
