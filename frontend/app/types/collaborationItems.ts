/**
 * 合作項目相關的型別定義
 */

/**
 * 合作項目
 */
export interface CollaborationItem {
  id: string
  title: string
  description?: string
  price: number
  parent_id?: string | null // 父項目 ID，null 表示頂層項目
  order: number // 同一層級內的排序順序
  children?: CollaborationItem[] // 子項目（用於前端展示）
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
}

/**
 * 更新合作項目請求
 */
export interface UpdateCollaborationItemRequest {
  title?: string
  description?: string
  price?: number
  parent_id?: string | null
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

