/**
 * 格式化工具函數
 * 統一管理所有格式化邏輯，避免在 composables 中混雜工具函數
 */

/**
 * 格式化金額
 */
export const formatAmount = (amount?: number, currency = 'TWD'): string => {
  if (!amount) return '-'
  return new Intl.NumberFormat('zh-TW', {
    style: 'currency',
    currency: currency === 'TWD' ? 'TWD' : currency,
    minimumFractionDigits: 0
  }).format(amount)
}

/**
 * 格式化日期（相對時間）
 */
export const formatRelativeDate = (dateStr?: string): string => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  const now = new Date()
  const diffDays = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24))

  if (diffDays === 0) {
    return '今天'
  } else if (diffDays === 1) {
    return '昨天'
  } else if (diffDays < 7) {
    return `${diffDays} 天前`
  } else {
    return date.toLocaleDateString('zh-TW', { month: 'short', day: 'numeric' })
  }
}

/**
 * 格式化完整日期
 */
export const formatFullDate = (dateStr?: string): string => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-TW', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

/**
 * 檢查截止日期是否緊急（< 3 天）
 */
export const isDeadlineUrgent = (deadlineDate?: string): boolean => {
  if (!deadlineDate) return false
  const deadline = new Date(deadlineDate)
  const now = new Date()
  const diffDays = Math.floor((deadline.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  return diffDays >= 0 && diffDays < 3
}






