/**
 * 假資料工具
 * 用於在 API 尚未實作時提供初始顯示資料
 */

import type { CollaborationItem, CollaborationItemPhase, WorkflowTemplate } from '~/types/collaborationItems'

/**
 * 可用的流程顏色選項
 */
export const WORKFLOW_COLORS = [
  { value: 'primary', label: '主要色', class: 'bg-primary-500' },
  { value: 'secondary', label: '次要色', class: 'bg-gray-500' },
  { value: 'success', label: '成功', class: 'bg-green-500' },
  { value: 'warning', label: '警告', class: 'bg-yellow-500' },
  { value: 'error', label: '錯誤', class: 'bg-red-500' },
  { value: 'info', label: '資訊', class: 'bg-blue-500' },
  { value: 'purple', label: '紫色', class: 'bg-purple-500' },
  { value: 'pink', label: '粉色', class: 'bg-pink-500' },
  { value: 'indigo', label: '靛藍', class: 'bg-indigo-500' },
] as const

/**
 * 生成假資料的流程範本
 */
export const generateMockWorkflowTemplates = (count: number = 3): WorkflowTemplate[] => {
  const mockWorkflowNames = ['YouTube 標準流程', '貼文標準流程', '短影片流程']
  const mockPhasesData = [
    [{ name: '腳本撰寫', days: 3 }, { name: '影片拍攝', days: 5 }, { name: '後製剪輯', days: 7 }],
    [{ name: '文案撰寫', days: 2 }, { name: '圖片設計', days: 3 }],
    [{ name: '創意發想', days: 2 }, { name: '拍攝', days: 3 }, { name: '剪輯', days: 2 }]
  ]
  
  return Array.from({ length: Math.min(count, mockWorkflowNames.length) }, (_, index) => {
    const workflowId = `mock_workflow_${Date.now()}_${index}`
    return {
      id: workflowId,
      name: mockWorkflowNames[index],
      description: `${mockWorkflowNames[index]} 的預設流程`,
      color: WORKFLOW_COLORS[index % WORKFLOW_COLORS.length].value,
      phases: mockPhasesData[index].map((phase, phaseIndex) => ({
        id: `mock_phase_${workflowId}_${phaseIndex}`,
        workflow_template_id: workflowId,
        name: phase.name,
        duration_days: phase.days,
        order: phaseIndex,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      })),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    }
  })
}

/**
 * 生成假資料的階段範本（兼容舊版）
 */
export const generateMockPhases = (workflowTemplateId: string, count: number = 3): CollaborationItemPhase[] => {
  const mockPhaseNames = ['腳本撰寫', '影片拍攝', '後製剪輯', '審核修改', '上線發布']
  const mockDurations = [3, 5, 7, 2, 1]
  
  return Array.from({ length: Math.min(count, mockPhaseNames.length) }, (_, index) => ({
    id: `mock_phase_${Date.now()}_${index}`,
    workflow_template_id: workflowTemplateId,
    name: mockPhaseNames[index],
    duration_days: mockDurations[index] || 1,
    order: index,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  }))
}

/**
 * 生成假資料的合作項目
 */
export const generateMockCollaborationItems = (workflowTemplates: WorkflowTemplate[], count: number = 3): CollaborationItem[] => {
  const mockTitles = ['YouTube 合作', 'Instagram 貼文', 'TikTok 短影片', 'Facebook 貼文', 'Twitter 推文']
  
  return Array.from({ length: Math.min(count, mockTitles.length) }, (_, index) => ({
    id: `mock_item_${Date.now()}_${index}`,
    title: mockTitles[index],
    description: `這是 ${mockTitles[index]} 的範本`,
    price: (index + 1) * 1000,
    parent_id: null,
    workflow_id: workflowTemplates[index % workflowTemplates.length]?.id || null,
    workflow: workflowTemplates[index % workflowTemplates.length] || undefined,
    order: index,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  }))
}

