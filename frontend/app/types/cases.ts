// 案件相關的型別定義
import type { FieldValue } from '~/types/fields'
import type { CollaborationItem } from '~/types/collaborationItems'

/**
 * 案件狀態枚舉
 */
export type CaseStatus = 'to_confirm' | 'in_progress' | 'completed' | 'cancelled' | 'other'

/**
 * 視圖類型
 */
export type ViewType = 'board' | 'list'

/**
 * 任務狀態
 */
export type TaskStatus = 'pending' | 'in_progress' | 'completed' | 'cancelled'

/**
 * 案件基本資訊
 */
export interface Case {
  id: string
  title: string
  brand_name: string
  collaboration_type?: string
  status: CaseStatus
  quoted_amount?: number
  final_amount?: number
  currency?: string
  deadline_date?: string
  contact_name?: string
  contact_email?: string
  contact_phone?: string
  email_count?: number
  task_count?: number
  completed_task_count?: number
  collaboration_items?: string[] // 選中的合作項目 ID 列表
  created_at: string
  updated_at: string
}

/**
 * 案件完整詳情
 */
export interface CaseDetail extends Case {
  description?: string
  payment_status?: string
  contract_date?: string
  delivery_date?: string
  publish_date?: string
  notes?: string
  tags?: string[]
  source?: string
  emails?: CaseEmail[]
  tasks?: Task[]
  updates?: CaseUpdate[]
  collaboration_items_detail?: CollaborationItem[] // 完整的合作項目資訊
  collaboration_items_total?: number // 總價
  collaboration_items_custom?: Array<{ id: string; title: string; description?: string; price: number }> // 自訂合作項目
  phases?: CasePhase[] // 案件階段列表
  start_date?: string // 案件開始日期
}

/**
 * 案件郵件（簡化版，用於案件詳情）
 */
export interface CaseEmail {
  id: string
  subject?: string
  from_email: string
  from_name?: string
  received_at: string
  email_type?: string
}

/**
 * 任務資訊
 */
export interface Task {
  id: string
  case_id: string
  title: string
  description?: string
  due_date?: string
  due_time?: string
  status: TaskStatus
  completed_at?: string
  order: number // 用於拖曳排序
  source?: string
  reminder_days?: number
  created_at: string
  updated_at?: string
}

/**
 * 案件狀態變更歷史
 */
export interface CaseUpdate {
  id: string
  case_id: string
  update_type: string
  old_value?: string
  new_value?: string
  created_at: string
}

/**
 * 案件查詢參數
 */
export interface CaseQueryParams {
  page?: number
  per_page?: number
  status?: CaseStatus
  brand?: string
  sort?: string
  search?: string
}

/**
 * 案件列表回應
 */
export interface CaseListResponse {
  data: Case[]
  pagination: {
    page: number
    per_page: number
    total: number
    total_pages: number
  }
}

/**
 * 建立案件請求
 */
export interface CreateCaseRequest {
  title: string
  brand_name: string
  collaboration_type?: string
  description?: string
  quoted_amount?: number
  deadline_date?: string
  contact_name?: string
  contact_email?: string
  contact_phone?: string
  notes?: string
  tags?: string[]
  collaboration_items?: string[] // 選中的合作項目 ID 列表
  custom_fields?: Record<string, any>
}

/**
 * 更新案件請求
 */
export interface UpdateCaseRequest {
  title?: string
  brand_name?: string
  collaboration_type?: string
  description?: string
  status?: CaseStatus
  quoted_amount?: number
  final_amount?: number
  deadline_date?: string
  contact_name?: string
  contact_email?: string
  contact_phone?: string
  notes?: string
  tags?: string[]
  collaboration_items?: string[] // 選中的合作項目 ID 列表
  custom_fields?: Record<string, any>
}

/**
 * 建立任務請求
 */
export interface CreateTaskRequest {
  title: string
  description?: string
  due_date?: string
  due_time?: string
  reminder_days?: number
}

/**
 * 更新任務請求
 */
export interface UpdateTaskRequest {
  title?: string
  description?: string
  due_date?: string
  due_time?: string
  status?: TaskStatus
  order?: number
}

/**
 * 任務排序請求
 */
export interface ReorderTasksRequest {
  task_ids: string[] // 按順序排列的任務 ID 陣列
}

/**
 * 階段基本資訊
 */
export interface Phase {
  id: string
  name: string
  duration_days: number
  order: number
  start_date?: string
  end_date?: string
}

/**
 * 案件階段
 */
export interface CasePhase {
  id: string
  case_id: string
  name: string
  start_date: string
  end_date: string
  duration_days: number
  order: number
  collaboration_item_phase_id?: string // 來源階段（如果有）
  created_at: string
  updated_at: string
}

/**
 * 建立案件階段請求
 */
export interface CreateCasePhaseRequest {
  name: string
  start_date: string
  duration_days: number
  order?: number
  collaboration_item_phase_id?: string
}

/**
 * 更新案件階段請求
 */
export interface UpdateCasePhaseRequest {
  name?: string
  start_date?: string
  end_date?: string
  duration_days?: number
  order?: number
}

/**
 * 套用流程請求
 */
export interface ApplyTemplateRequest {
  workflow_id: string
  start_date: string
}

