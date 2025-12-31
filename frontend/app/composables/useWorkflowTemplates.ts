/**
 * 流程範本相關 composable
 * 提供便捷的流程範本操作方法
 */

import { generateMockWorkflowTemplates } from '~/utils/mockData'
import type { WorkflowTemplate } from '~/types/collaborationItems'

export const useWorkflowTemplates = () => {
  // 使用假資料作為初始狀態
  const workflowTemplates = ref<WorkflowTemplate[]>(generateMockWorkflowTemplates(3))
  const loading = ref(false)
  const error = ref<string | null>(null)

  /**
   * 載入流程範本列表
   */
  const fetchWorkflows = async () => {
    loading.value = true
    error.value = null

    try {
      // TODO: 呼叫 API 載入流程範本列表
      // const config = useRuntimeConfig()
      // const authStore = useAuthStore()
      // const data = await $fetch<{ data: WorkflowTemplate[] }>(
      //   `${config.public.apiBase}/api/v1/workflow-templates`,
      //   {
      //     headers: {
      //       Authorization: `Bearer ${authStore.token}`
      //     }
      //   }
      // )
      // workflowTemplates.value = data.data

      // 暫時：使用假資料
      await new Promise(resolve => setTimeout(resolve, 300))
      workflowTemplates.value = generateMockWorkflowTemplates(3)
    } catch (e: any) {
      const is404 = e?.statusCode === 404 || e?.status === 404 || e?.response?.status === 404
      if (!is404) {
        error.value = '載入流程範本失敗'
        console.error('載入流程範本失敗:', e)
      }
      // 即使失敗也保留假資料
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
    // TODO: 呼叫 API 建立流程範本
    const newWorkflow: WorkflowTemplate = {
      id: `workflow_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      name: data.name,
      description: data.description,
      color: data.color || 'primary',
      phases: [],
      order: workflowTemplates.value.length,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    }
    workflowTemplates.value.push(newWorkflow)
    return newWorkflow
  }

  /**
   * 更新流程範本
   */
  const updateWorkflow = async (id: string, data: { name?: string; description?: string; color?: string }): Promise<WorkflowTemplate> => {
    // TODO: 呼叫 API 更新流程範本
    const index = workflowTemplates.value.findIndex(w => w.id === id)
    if (index !== -1) {
      workflowTemplates.value[index] = {
        ...workflowTemplates.value[index],
        ...data,
        updated_at: new Date().toISOString()
      }
      return workflowTemplates.value[index]
    }
    throw new Error('流程範本不存在')
  }

  /**
   * 刪除流程範本
   */
  const deleteWorkflow = async (id: string): Promise<void> => {
    // TODO: 呼叫 API 刪除流程範本
    workflowTemplates.value = workflowTemplates.value.filter(w => w.id !== id)
  }

  /**
   * 重新排序流程範本
   */
  const reorderWorkflows = async (workflowIds: string[]): Promise<void> => {
    // TODO: 呼叫 API 重新排序
    const sortedWorkflows = workflowIds.map(id => workflowTemplates.value.find(w => w.id === id)).filter(Boolean) as WorkflowTemplate[]
    sortedWorkflows.forEach((workflow, index) => {
      workflow.order = index
    })
    workflowTemplates.value = sortedWorkflows
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
    reorderWorkflows
  }
}

