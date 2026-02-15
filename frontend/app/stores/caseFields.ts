import { defineStore } from 'pinia'
import { fieldsStorage, generateTempId } from '~/utils/localStorage'
import type {
  CaseField,
  SystemField,
  CustomField,
  FieldListResponse,
  CreateFieldRequest,
  UpdateFieldRequest,
  ReorderFieldsRequest,
} from '~/types/fields'
import { SystemFieldName } from '~/types/fields'

// 前端預設系統屬性（當尚未串接後端時，確保表單有基本欄位可以操作）
const defaultSystemFields: SystemField[] = [
  {
    id: 'system-title',
    name: 'title',
    label: '案件標題',
    type: 'text',
    is_system: true,
    system_column_name: SystemFieldName.TITLE,
    is_required: true,
    is_visible: true,
    order: 1,
    placeholder: '例如：Nike 球鞋業配'
  },
  {
    id: 'system-brand_name',
    name: 'brand_name',
    label: '品牌名稱',
    type: 'text',
    is_system: true,
    system_column_name: SystemFieldName.BRAND_NAME,
    is_required: true,
    is_visible: true,
    order: 2,
    placeholder: '例如：Nike'
  },
  {
    id: 'system-status',
    name: 'status',
    label: '案件狀態',
    type: 'select',
    is_system: true,
    system_column_name: SystemFieldName.STATUS,
    is_required: true,
    is_visible: true,
    order: 3,
    options: [
      { label: '待確認', value: 'to_confirm' },
      { label: '進行中', value: 'in_progress' },
      { label: '已完成', value: 'completed' },
      { label: '已取消', value: 'cancelled' },
      { label: '非合作案件', value: 'other' }
    ]
  },
  {
    id: 'system-deadline_date',
    name: 'deadline_date',
    label: '截止日期',
    type: 'date',
    is_system: true,
    system_column_name: SystemFieldName.DEADLINE_DATE,
    is_required: false,
    is_visible: true,
    order: 4
  },
  {
    id: 'system-quoted_amount',
    name: 'quoted_amount',
    label: '預估報價',
    type: 'number',
    is_system: true,
    system_column_name: SystemFieldName.QUOTED_AMOUNT,
    is_required: false,
    is_visible: true,
    order: 5
  }
]

export const useCaseFieldsStore = defineStore('caseFields', () => {
  // State
  const systemFields: Ref<SystemField[]> = ref([])
  const customFields: Ref<CustomField[]> = ref([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Actions
  const fetchFields = async () => {
    loading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const data = await $fetch<FieldListResponse>(`${config.public.apiBase}/api/v1/cases/fields`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      systemFields.value = data.system_fields || []
      customFields.value = data.custom_fields || []
      
      // 同步到 localStorage
      fieldsStorage.setFields({
        system_fields: systemFields.value,
        custom_fields: customFields.value
      })
    } catch (e: any) {
      error.value = e.message || '取得屬性列表失敗'
      console.error('Failed to fetch fields:', e)

      // Fallback 到 localStorage
      const localFields = fieldsStorage.getFields()
      if (localFields.system_fields.length > 0 || localFields.custom_fields.length > 0) {
        systemFields.value = localFields.system_fields
        customFields.value = localFields.custom_fields
      } else {
        // 如果後端尚未就緒且 localStorage 也沒有，使用前端預設的系統屬性
        systemFields.value = [...defaultSystemFields]
        fieldsStorage.setFields({
          system_fields: defaultSystemFields,
          custom_fields: []
        })
      }
    } finally {
      loading.value = false
    }
  }

  const createField = async (fieldData: CreateFieldRequest) => {
    loading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const newField = await $fetch<CustomField>(`${config.public.apiBase}/api/v1/cases/fields`, {
        method: 'POST',
        body: fieldData,
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      customFields.value.push(newField)
      
      // 同步到 localStorage
      fieldsStorage.addField(newField)

      return newField
    } catch (e: any) {
      error.value = e.message || '建立屬性失敗（已儲存到本地）'
      console.error('Failed to create field:', e)
      // Fallback 到 localStorage：建立臨時屬性
      const tempId = generateTempId()
      const maxOrder = Math.max(...customFields.value.map(f => f.order), 0)
      const newField: CustomField = {
        id: tempId,
        ...fieldData,
        is_system: false,
        order: maxOrder + 1,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      } as CustomField

      customFields.value.push(newField)
      fieldsStorage.addField(newField)

      return newField
    } finally {
      loading.value = false
    }
  }

  const updateField = async (id: string, fieldData: UpdateFieldRequest) => {
    loading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const updatedField = await $fetch<CaseField>(
        `${config.public.apiBase}/api/v1/cases/fields/${id}`,
        {
          method: 'PATCH',
          body: fieldData,
          headers: {
            Authorization: `Bearer ${authStore.token}`
          }
        }
      )

      // 更新對應的列表
      if (updatedField.is_system) {
        const index = systemFields.value.findIndex(f => f.id === id)
        if (index !== -1) {
          systemFields.value[index] = updatedField as SystemField
        }
      } else {
        const index = customFields.value.findIndex(f => f.id === id)
        if (index !== -1) {
          customFields.value[index] = updatedField as CustomField
        }
      }

      // 同步到 localStorage
      fieldsStorage.updateField(id, updatedField)

      return updatedField
    } catch (e: any) {
      error.value = e.message || '更新屬性失敗（已儲存到本地）'
      console.error('Failed to update field:', e)
      // Fallback 到 localStorage
      const allFields = [...systemFields.value, ...customFields.value]
      const field = allFields.find(f => f.id === id)
      if (field) {
        const updated = { ...field, ...fieldData, updated_at: new Date().toISOString() }
        if (field.is_system) {
          const index = systemFields.value.findIndex(f => f.id === id)
          if (index !== -1) {
            systemFields.value[index] = updated as SystemField
          }
        } else {
          const index = customFields.value.findIndex(f => f.id === id)
          if (index !== -1) {
            customFields.value[index] = updated as CustomField
          }
        }
        fieldsStorage.updateField(id, updated)
        return updated as CaseField
      }
      throw e
    } finally {
      loading.value = false
    }
  }

  const deleteField = async (id: string) => {
    loading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      await $fetch(`${config.public.apiBase}/api/v1/cases/fields/${id}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      // 從列表中移除（只可能是自定義屬性）
      customFields.value = customFields.value.filter(f => f.id !== id)
      
      // 從 localStorage 刪除
      fieldsStorage.deleteField(id)
    } catch (e: any) {
      error.value = e.message || '刪除屬性失敗（已從本地移除）'
      console.error('Failed to delete field:', e)
      // Fallback 到 localStorage
      customFields.value = customFields.value.filter(f => f.id !== id)
      fieldsStorage.deleteField(id)
    } finally {
      loading.value = false
    }
  }

  const reorderFields = async (fieldIds: string[]) => {
    loading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const data: ReorderFieldsRequest = { field_ids: fieldIds }

      await $fetch(`${config.public.apiBase}/api/v1/cases/fields/reorder`, {
        method: 'PATCH',
        body: data,
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      // 重新載入屬性列表以獲取更新後的順序
      await fetchFields()
    } catch (e: any) {
      error.value = e.message || '重新排序屬性失敗'
      console.error('Failed to reorder fields:', e)
      throw e
    } finally {
      loading.value = false
    }
  }

  const toggleFieldVisibility = async (id: string) => {
    const field = [...systemFields.value, ...customFields.value].find(f => f.id === id)
    if (!field) return

    return updateField(id, { is_visible: !field.is_visible })
  }

  // Getters
  const allFields = computed(() => {
    const baseFields =
      systemFields.value.length === 0 && customFields.value.length === 0
        ? defaultSystemFields
        : [...systemFields.value, ...customFields.value]

    return [...baseFields].sort((a, b) => a.order - b.order)
  })

  const visibleFields = computed(() => {
    return allFields.value.filter(f => f.is_visible)
  })

  const requiredFields = computed(() => {
    return allFields.value.filter(f => f.is_required)
  })

  // 重置狀態
  const reset = () => {
    systemFields.value = []
    customFields.value = []
    loading.value = false
    error.value = null
  }

  return {
    // State
    systemFields,
    customFields,
    loading,
    error,

    // Getters
    allFields,
    visibleFields,
    requiredFields,

    // Actions
    fetchFields,
    createField,
    updateField,
    deleteField,
    reorderFields,
    toggleFieldVisibility,
    reset
  }
})



