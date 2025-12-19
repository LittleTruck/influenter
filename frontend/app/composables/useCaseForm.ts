/**
 * 案件表單共用邏輯
 * 處理案件表單的初始化、驗證等邏輯
 */
import { logger } from '~/utils/logger'
import type { Case } from '~/types/cases'
import { useCaseFields } from '~/composables/useCaseFields'

export const useCaseForm = () => {
  const { visibleFields, validateFieldValue } = useCaseFields()

  /**
   * 初始化表單資料
   */
  const initializeFormData = async (
    formData: Record<string, unknown>,
    caseData?: Case | null
  ): Promise<void> => {
    if (caseData) {
      // 編輯模式：載入案件資料
      Object.assign(formData, caseData)
    } else {
      // 新建模式：使用預設值
      visibleFields.value.forEach(field => {
        if (field.default_value !== undefined) {
          formData[field.name] = field.default_value
        }
      })
    }
  }

  /**
   * 驗證表單
   */
  const validateForm = (formData: Record<string, unknown>): string[] => {
    const errors: string[] = []

    visibleFields.value.forEach(field => {
      if (field.is_required) {
        const error = validateFieldValue(field, formData[field.name])
        if (error) {
          errors.push(error)
        }
      }
    })

    return errors
  }

  return {
    initializeFormData,
    validateForm
  }
}

