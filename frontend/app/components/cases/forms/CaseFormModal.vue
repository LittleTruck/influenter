<script setup lang="ts">
import type { Case, CreateCaseRequest, UpdateCaseRequest } from '~/types/cases'
import { useCaseFields } from '~/composables/useCaseFields'
import { useCases } from '~/composables/useCases'
import { useCaseForm } from '~/composables/useCaseForm'
import { useFormModal } from '~/composables/useFormModal'
import { BaseModal, BaseButton, BaseFormField } from '~/components/base'
import FieldInput from '~/components/cases/fields/FieldInput.vue'
import CollaborationItemsSelector from '~/components/cases/fields/CollaborationItemsSelector.vue'

interface Props {
  /** 是否顯示 */
  modelValue: boolean
  /** 案件資料（編輯模式） */
  case?: Case | null
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'submit': [data: CreateCaseRequest | UpdateCaseRequest]
}>()

const toast = useToast()
const { visibleFields } = useCaseFields()
const { createCase, updateCase } = useCases()
const { initializeFormData, validateForm } = useCaseForm()
const { submitForm, isSubmitting } = useFormModal<CreateCaseRequest>()
const { flatItems } = useCollaborationItems()

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const isEditMode = computed(() => !!props.case)

// 表單資料
const formData = reactive<Partial<CreateCaseRequest | UpdateCaseRequest>>({})

// 初始化表單資料
const initForm = async () => {
  // 清空表單
  Object.keys(formData).forEach(key => {
    delete formData[key as keyof typeof formData]
  })
  await initializeFormData(formData as Record<string, unknown>, props.case || undefined)
  
  // 確保 collaboration_items 初始化為陣列
  if (!formData.collaboration_items) {
    formData.collaboration_items = []
  }
  
  // 如果有自訂項目，需要將 ID 列表和自訂項目合併為完整項目列表
  if (props.case && (props.case as any).collaboration_items_custom) {
    const customItems = (props.case as any).collaboration_items_custom || []
    const itemIds = formData.collaboration_items as string[] || []
    
    // 構建完整項目列表
    const fullItems = itemIds.map(id => {
      // 檢查是否為自訂項目
      const customItem = customItems.find((item: any) => item.id === id)
      if (customItem) {
        return {
          ...customItem,
          isCustom: true
        }
      }
      // 從預設列表查找
      const item = flatItems.value.find(i => i.id === id)
      if (item) {
        return {
          id: item.id,
          title: item.title,
          description: item.description,
          price: item.price,
          isCustom: false
        }
      }
      return null
    }).filter(Boolean)
    
    formData.collaboration_items = fullItems as any
  }
}

// 監聽 props 變化，重新初始化表單
watch([() => props.case, () => props.modelValue], async ([newCase, isOpen]) => {
  if (isOpen) {
    await initForm()
  }
}, { immediate: true })

// 組件掛載時初始化
onMounted(async () => {
  if (props.modelValue) {
    await initForm()
  }
})

// 提交表單
const handleSubmit = async () => {
  const errors = validateForm(formData)
  if (errors.length > 0) {
    errors.forEach(error => {
      toast.add({
        title: '驗證錯誤',
        description: error,
        color: 'error'
      })
    })
    return
  }

  // 處理合作項目：將自訂項目分離出來
  const collaborationItems = formData.collaboration_items as any
  if (Array.isArray(collaborationItems) && collaborationItems.length > 0 && typeof collaborationItems[0] === 'object' && 'title' in collaborationItems[0]) {
    // 這是完整項目列表（包含自訂項目）
    const itemList = collaborationItems as Array<{ id?: string; title: string; description?: string; price: number; isCustom?: boolean }>
    const itemIds = itemList.map(item => item.id || `custom_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`)
    const customItems = itemList.filter(item => item.isCustom || !item.id || item.id.startsWith('custom_')).map(item => ({
      id: item.id || `custom_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      title: item.title,
      description: item.description,
      price: item.price
    }))
    
    formData.collaboration_items = itemIds
    ;(formData as any).collaboration_items_custom = customItems
  }

  const success = await submitForm(
    async () => {
      if (isEditMode.value && props.case) {
        await updateCase(props.case.id, formData as UpdateCaseRequest)
        return '案件已更新'
      } else {
        await createCase(formData as CreateCaseRequest)
        return '案件已建立'
      }
    },
    isEditMode.value ? '案件已更新' : '案件已建立'
  )

  if (success) {
    isOpen.value = false
    emit('submit', { ...formData })
  }
}

const handleCancel = () => {
  isOpen.value = false
}
</script>

<template>
  <BaseModal 
    v-model="isOpen" 
    :title="isEditMode ? '編輯案件' : '建立案件'"
    size="lg"
  >
    <template #body>
      <div class="space-y-4">
        <!-- 動態生成欄位 -->
        <FieldInput
          v-for="field in visibleFields"
          :key="field.id"
          :field="field"
          v-model="formData[field.name]"
          :editable="true"
          :show-label="true"
          :show-error="true"
        />

        <!-- 合作項目選擇器 -->
        <BaseFormField label="合作項目" name="collaboration_items">
          <CollaborationItemsSelector
            v-model="formData.collaboration_items"
            class="w-full"
          />
        </BaseFormField>
      </div>
    </template>

    <template #footer>
      <div class="flex justify-end gap-2">
        <BaseButton variant="ghost" @click="handleCancel">取消</BaseButton>
        <BaseButton
          :loading="isSubmitting"
          @click="handleSubmit"
        >
          {{ isEditMode ? '更新' : '建立' }}
        </BaseButton>
      </div>
    </template>
  </BaseModal>
</template>

