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
const { findItemById } = useCollaborationItems()

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
  
  // 從 case_collaboration_items 多對多關聯初始化
  const fd = formData as any
  if (!fd.collaboration_items) {
    fd.collaboration_items = []
  }

  if (props.case?.case_collaboration_items?.length) {
    fd.collaboration_items = props.case.case_collaboration_items
      .sort((a, b) => a.order - b.order)
      .map(cci => cci.collaboration_item_id)
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

  // 合作項目 ID 列表（直接傳送給後端）
  const fd = formData as any
  if (Array.isArray(fd.collaboration_items)) {
    fd.collaboration_items = fd.collaboration_items.filter((id: any) => typeof id === 'string' && id)
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
            v-model="(formData as any).collaboration_items"
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

