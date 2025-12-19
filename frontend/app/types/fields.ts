// 案件屬性相關的型別定義

/**
 * 屬性類型枚舉
 */
export type FieldType = 
  | 'text'
  | 'number'
  | 'date'
  | 'select'
  | 'multiselect'
  | 'checkbox'
  | 'url'
  | 'email'
  | 'phone'
  | 'textarea'

/**
 * 案件屬性定義（系統 + 自定義）
 */
export interface CaseField {
  id: string
  name: string
  label: string
  type: FieldType
  is_system: boolean
  is_required: boolean
  is_visible: boolean
  order: number
  default_value?: FieldValue
  options?: FieldOption[] // 用於 select 和 multiselect
  placeholder?: string
  description?: string
  created_at?: string
  updated_at?: string
}

/**
 * 系統屬性名稱（固定屬性）
 */
export enum SystemFieldName {
  ID = 'id',
  TITLE = 'title',
  BRAND_NAME = 'brand_name',
  STATUS = 'status',
  QUOTED_AMOUNT = 'quoted_amount',
  FINAL_AMOUNT = 'final_amount',
  CURRENCY = 'currency',
  DEADLINE_DATE = 'deadline_date',
  CREATED_AT = 'created_at',
  UPDATED_AT = 'updated_at',
  CONTACT_NAME = 'contact_name',
  CONTACT_EMAIL = 'contact_email',
  CONTACT_PHONE = 'contact_phone',
}

/**
 * 系統屬性定義
 */
export interface SystemField extends CaseField {
  is_system: true
  system_column_name: SystemFieldName
}

/**
 * 自定義屬性定義
 */
export interface CustomField extends CaseField {
  is_system: false
}

/**
 * 屬性選項（用於 select 和 multiselect）
 */
export interface FieldOption {
  label: string
  value: string | number
  color?: string
}

/**
 * 屬性值類型
 * 根據不同的 FieldType 對應不同的值類型
 */
export type FieldValue = 
  | string 
  | number 
  | boolean 
  | string[] 
  | Date 
  | null 
  | undefined

/**
 * 屬性值對象（用於 API 回應）
 */
export interface FieldValueObject {
  field_id: string
  field_name: string
  value: FieldValue
  field_type: FieldType
}

/**
 * 建立屬性請求
 */
export interface CreateFieldRequest {
  name: string
  label: string
  type: FieldType
  is_required?: boolean
  is_visible?: boolean
  default_value?: FieldValue
  options?: FieldOption[]
  placeholder?: string
  description?: string
}

/**
 * 更新屬性請求
 */
export interface UpdateFieldRequest {
  label?: string
  is_required?: boolean
  is_visible?: boolean
  default_value?: FieldValue
  options?: FieldOption[]
  placeholder?: string
  description?: string
  order?: number
}

/**
 * 重新排序屬性請求
 */
export interface ReorderFieldsRequest {
  field_ids: string[] // 按順序排列的屬性 ID 陣列
}

/**
 * 屬性列表回應
 */
export interface FieldListResponse {
  system_fields: SystemField[]
  custom_fields: CustomField[]
}

