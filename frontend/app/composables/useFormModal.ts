/**
 * 表單 Modal 共用邏輯
 * 處理表單驗證、提交、錯誤處理等通用邏輯
 */
export const useFormModal = <T extends Record<string, unknown> = Record<string, unknown>>() => {
  const toast = useToast()
  const isSubmitting = ref(false)

  /**
   * 執行表單提交
   * @param submitFn - 提交函數
   * @param successMessage - 成功訊息
   * @param errorMessage - 錯誤訊息
   * @returns 是否成功
   */
  const submitForm = async (
    submitFn: () => Promise<string>,
    successMessage: string,
    errorMessage = '操作失敗'
  ): Promise<boolean> => {
    if (isSubmitting.value) return false

    isSubmitting.value = true
    try {
      await submitFn()
      toast.add({
        title: '成功',
        description: successMessage,
        color: 'success'
      })
      return true
    } catch (error: unknown) {
      const errorMessageFinal = error instanceof Error ? error.message : errorMessage
      toast.add({
        title: '錯誤',
        description: errorMessageFinal,
        color: 'error'
      })
      return false
    } finally {
      isSubmitting.value = false
    }
  }

  /**
   * 驗證必填欄位
   * @param data - 表單數據
   * @param requiredFields - 必填欄位列表
   * @returns 錯誤訊息或 null
   */
  const validateRequired = (data: T, requiredFields: (keyof T)[]): string | null => {
    for (const field of requiredFields) {
      const value = data[field]
      if (value === null || value === undefined || value === '') {
        return `${String(field)} 為必填欄位`
      }
    }
    return null
  }

  return {
    isSubmitting: computed(() => isSubmitting.value),
    submitForm,
    validateRequired
  }
}

