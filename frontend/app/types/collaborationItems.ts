/**
 * 合作項目相關的型別定義
 */

/**
 * 流程範本（Workflow Template）
 */
export interface WorkflowTemplate {
  id: string
  name: string
  description?: string
  color: string // 固定顏色選項：primary, secondary, success, warning, error, info, purple, pink, indigo 等
  phases: CollaborationItemPhase[]
  order?: number // 排序順序
  created_at: string
  updated_at: string
}

/**
 * 合作項目
 */
export interface CollaborationItem {
  id: string
  title: string
  description?: string
  price: number
  parent_id?: string | null // 父項目 ID，null 表示頂層項目
  workflow_id?: string | null // 流程範本 ID，null 表示未設定
  order: number // 同一層級內的排序順序
  children?: CollaborationItem[] // 子項目（用於前端展示）
  workflow?: WorkflowTemplate // 流程範本（用於前端展示）
  created_at: string
  updated_at: string
}

/**
 * 建立合作項目請求
 */
export interface CreateCollaborationItemRequest {
  title: string
  description?: string
  price: number
  parent_id?: string | null
  workflow_id?: string | null
}

/**
 * 更新合作項目請求
 */
export interface UpdateCollaborationItemRequest {
  title?: string
  description?: string
  price?: number
  parent_id?: string | null
  workflow_id?: string | null
}

/**
 * 重新排序請求
 */
export interface ReorderItemsRequest {
  item_ids: string[] // 同一層級內的項目 ID 順序
  parent_id?: string | null // 要排序的層級
}

/**
 * 合作項目列表回應
 */
export interface CollaborationItemListResponse {
  data: CollaborationItem[]
}

/**
 * 合作項目階段
 */
export interface CollaborationItemPhase {
  id: string
  workflow_template_id?: string // 屬於哪個流程範本
  collaboration_item_id?: string // 舊版兼容（直接屬於合作項目）
  name: string
  duration_days: number
  order: number
  created_at: string
  updated_at: string
}

/**
 * 建立合作項目階段請求
 */
export interface CreateCollaborationItemPhaseRequest {
  collaboration_item_id: string
  name: string
  duration_days: number
  order?: number
}

/**
 * 更新合作項目階段請求
 */
export interface UpdateCollaborationItemPhaseRequest {
  name?: string
  duration_days?: number
  order?: number
}

/**
 * 建立流程範本請求
 */
export interface CreateWorkflowTemplateRequest {
  name: string
  description?: string
  color: string
}

/**
 * 更新流程範本請求
 */
export interface UpdateWorkflowTemplateRequest {
  name?: string
  description?: string
  color?: string
}




