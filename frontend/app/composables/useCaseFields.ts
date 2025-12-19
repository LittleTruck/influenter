import type { CaseField, FieldType, FieldValue } from '~/types/fields'

/**
 * 案件屬性管理 composable
 * 提供便捷的屬性操作方法
 */
export const useCaseFields = () => {
  // 確保在正確的上下文中使用 store
  let fieldsStore: ReturnType<typeof useCaseFieldsStore>
  try {
    fieldsStore = useCaseFieldsStore()
  } catch (error) {
    console.error('Failed to initialize case fields store:', error)
    throw error
  }

  /**
   * 根據屬性類型返回對應的渲染組件名稱
   */
  const getFieldRenderer = (fieldType: FieldType): string => {
    // 這個函數可以返回組件名稱，實際渲染由組件處理
    return fieldType
  }

  /**
   * 根據屬性類型返回對應的輸入組件名稱
   */
  const getFieldInput = (fieldType: FieldType): string => {
    return fieldType
  }

  /**
   * 驗證屬性值
   */
  const validateFieldValue = (field: CaseField, value: FieldValue): string | null => {
    // 必填驗證
    if (field.is_required && (value === null || value === undefined || value === '')) {
      return `${field.label} 為必填欄位`
    }

    // 類型驗證
    switch (field.type) {
      case 'email':
        if (value && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
          return '請輸入有效的 Email 地址'
        }
        break
      case 'url':
        if (value && !/^https?:\/\/.+/.test(value)) {
          return '請輸入有效的 URL（需包含 http:// 或 https://）'
        }
        break
      case 'phone':
        if (value && !/^[\d\s\-+()]+$/.test(value)) {
          return '請輸入有效的電話號碼'
        }
        break
      case 'number':
        if (value && isNaN(Number(value))) {
          return '請輸入有效的數字'
        }
        break
    }

    return null
  }

  /**
   * 格式化屬性值顯示
   */
  const formatFieldValue = (field: CaseField, value: FieldValue): string => {
    if (value === null || value === undefined || value === '') {
      return '-'
    }

    switch (field.type) {
      case 'date':
        return new Date(value).toLocaleDateString('zh-TW')
      case 'number':
        return new Intl.NumberFormat('zh-TW').format(Number(value))
      case 'checkbox':
        return value ? '是' : '否'
      case 'multiselect':
        return Array.isArray(value) ? value.join(', ') : String(value)
      default:
        return String(value)
    }
  }

  return {
    // Store 方法
    fetchFields: fieldsStore.fetchFields,
    createField: fieldsStore.createField,
    updateField: fieldsStore.updateField,
    deleteField: fieldsStore.deleteField,
    reorderFields: fieldsStore.reorderFields,
    toggleFieldVisibility: fieldsStore.toggleFieldVisibility,

    // 工具函數
    getFieldRenderer,
    getFieldInput,
    validateFieldValue,
    formatFieldValue,

    // Computed 值
    systemFields: computed(() => fieldsStore.systemFields),
    customFields: computed(() => fieldsStore.customFields),
    allFields: computed(() => fieldsStore.allFields),
    visibleFields: computed(() => fieldsStore.visibleFields),
    requiredFields: computed(() => fieldsStore.requiredFields),
    loading: computed(() => fieldsStore.loading),
    error: computed(() => fieldsStore.error)
  }
}

