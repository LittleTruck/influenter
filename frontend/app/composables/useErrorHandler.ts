/**
 * 錯誤處理 composable
 * 統一處理錯誤提示和日誌記錄
 */
import { logger } from '~/utils/logger'

interface ApiError {
  message?: string
  data?: {
    message?: string
  }
  statusCode?: number
}

export const useErrorHandler = () => {
  const toast = useToast()

  /**
   * 處理錯誤並顯示提示
   * @param error - 錯誤對象
   * @param defaultMessage - 默認錯誤訊息
   * @param options - 選項配置
   * @returns 錯誤訊息
   */
  const handleError = (
    error: unknown,
    defaultMessage = '操作失敗',
    options?: {
      log?: boolean
      showToast?: boolean
      component?: string
      action?: string
    }
  ): string => {
    const { log = true, showToast = true, component, action } = options || {}
    
    // 提取錯誤訊息
    let message = defaultMessage
    if (error && typeof error === 'object') {
      const apiError = error as ApiError
      message = apiError.message || apiError.data?.message || defaultMessage
    } else if (typeof error === 'string') {
      message = error
    }

    // 記錄錯誤
    if (log) {
      logger.error(message, error instanceof Error ? error : new Error(message), {
        component,
        action
      })
    }

    // 顯示提示
    if (showToast) {
      toast.add({
        title: '錯誤',
        description: message,
        color: 'error'
      })
    }

    return message
  }

  /**
   * 處理成功並顯示提示
   * @param message - 成功訊息
   */
  const handleSuccess = (message: string): void => {
    toast.add({
      title: '成功',
      description: message,
      color: 'success'
    })
  }

  return {
    handleError,
    handleSuccess
  }
}

