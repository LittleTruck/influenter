/**
 * 案件狀態相關工具函數
 * 統一管理狀態相關的常量和映射
 */
import type { CaseStatus } from '~/types/cases'

/**
 * 狀態顏色映射
 */
export const STATUS_COLORS: Record<CaseStatus, 'warning' | 'primary' | 'success' | 'error' | 'neutral'> = {
  to_confirm: 'warning',
  in_progress: 'primary',
  completed: 'success',
  cancelled: 'error'
} as const

/**
 * 狀態標籤映射
 */
export const STATUS_LABELS: Record<CaseStatus, string> = {
  to_confirm: '待確認',
  in_progress: '進行中',
  completed: '已完成',
  cancelled: '已取消'
} as const

/**
 * 狀態顏色十六進位值映射（用於漸層背景等）
 */
export const STATUS_COLOR_HEX: Record<CaseStatus, string> = {
  to_confirm: '#eab308',
  in_progress: '#3b82f6',
  completed: '#22c55e',
  cancelled: '#ef4444'
} as const

/**
 * 取得狀態顏色
 */
export const getStatusColor = (status: CaseStatus): 'warning' | 'primary' | 'success' | 'error' | 'neutral' => {
  return STATUS_COLORS[status] || 'neutral'
}

/**
 * 取得狀態標籤文字
 */
export const getStatusLabel = (status: CaseStatus): string => {
  return STATUS_LABELS[status] || status
}

/**
 * 取得狀態顏色十六進位值
 */
export const getStatusColorHex = (status: CaseStatus): string => {
  return STATUS_COLOR_HEX[status] || '#6b7280'
}



