/**
 * 錯誤處理工具函數
 * 統一提取錯誤訊息和記錄錯誤
 */
import { logger } from './logger'

interface ApiError {
  message?: string
  data?: {
    message?: string
  }
  statusCode?: number
}

/**
 * 從錯誤對象中提取錯誤訊息
 */
export const extractErrorMessage = (error: unknown, defaultMessage: string): string => {
  if (error instanceof Error) {
    return error.message || defaultMessage
  }
  if (error && typeof error === 'object') {
    const apiError = error as ApiError
    return apiError.message || apiError.data?.message || defaultMessage
  }
  if (typeof error === 'string') {
    return error
  }
  return defaultMessage
}

/**
 * 記錄錯誤並返回錯誤訊息
 */
export const logError = (
  error: unknown,
  defaultMessage: string,
  context?: { component?: string; action?: string }
): string => {
  const errorMessage = extractErrorMessage(error, defaultMessage)
  const errorObj = error instanceof Error ? error : new Error(errorMessage)
  
  logger.error(errorMessage, errorObj, context)
  
  return errorMessage
}






