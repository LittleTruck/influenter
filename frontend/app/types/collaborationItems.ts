/**
 * 合作項目相關的型別定義
 */

/**
 * 合作項目類型
 */
export type CollaborationItemType = 'individual' | 'bundle'

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
  notes?: string // 注意事項（純文字）
  price: number
  type: CollaborationItemType // 'individual' 或 'bundle'
  bundle_items?: BundleItemRef[] // 僅 bundle 類型有值
  workflow_id?: string | null // 僅 individual 類型可設定流程範本
  order: number // 排序順序
  workflow?: WorkflowTemplate // 流程範本（用於前端展示）
  created_at: string
  updated_at: string
}

/**
 * 組合項目內含單項的關聯
 */
export interface BundleItemRef {
  id: string
  bundle_id: string
  item_id: string
  order: number
  item?: CollaborationItem // 解析後的單項資訊
}

/**
 * 建立合作項目請求
 */
export interface CreateCollaborationItemRequest {
  title: string
  description?: string
  notes?: string
  price: number
  type: CollaborationItemType
  bundle_item_ids?: string[] // 僅 bundle 類型：包含的 individual 項目 ID
  workflow_id?: string | null // 僅 individual 類型
}

/**
 * 更新合作項目請求
 */
export interface UpdateCollaborationItemRequest {
  title?: string
  description?: string
  notes?: string
  price?: number
  bundle_item_ids?: string[] // 僅 bundle 類型
  workflow_id?: string | null // 僅 individual 類型
}

/**
 * 重新排序請求
 */
export interface ReorderItemsRequest {
  item_ids: string[] // 項目 ID 順序
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
