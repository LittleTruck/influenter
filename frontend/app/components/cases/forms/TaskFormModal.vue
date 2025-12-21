<script setup lang="ts">
import type { Task, CreateTaskRequest, UpdateTaskRequest } from '~/types/cases'
import { useCases } from '~/composables/useCases'
import { useFormModal } from '~/composables/useFormModal'
import { BaseModal, BaseButton, BaseInput, BaseTextarea, BaseFormGroup } from '~/components/base'

interface Props {
  /** 是否顯示 */
  modelValue: boolean
  /** 案件 ID */
  caseId: string
  /** 任務資料（編輯模式） */
  task?: Task | null
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'submit': []
}>()

const toast = useToast()
const { createTask, updateTask } = useCases()
const { submitForm, validateRequired, isSubmitting } = useFormModal()

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const isEditMode = computed(() => !!props.task)

// 表單資料
const formData = reactive<CreateTaskRequest>({
  title: '',
  description: '',
  due_date: '',
  due_time: '',
  reminder_days: 1
})

// 初始化表單資料
onMounted(() => {
  if (props.task) {
    Object.assign(formData, {
      title: props.task.title,
      description: props.task.description || '',
      due_date: props.task.due_date || '',
      due_time: props.task.due_time || '',
      reminder_days: props.task.reminder_days || 1
    })
  }
})

// 提交表單
const handleSubmit = async () => {
  const error = validateRequired(formData, ['title'])
  if (error) {
    toast.add({
      title: '驗證錯誤',
      description: error,
      color: 'error'
    })
    return
  }

  const success = await submitForm(
    async () => {
      if (isEditMode.value && props.task) {
        await updateTask(props.task.id, formData as UpdateTaskRequest)
        return '任務已更新'
      } else {
        await createTask(props.caseId, formData)
        return '任務已建立'
      }
    },
    isEditMode.value ? '任務已更新' : '任務已建立'
  )

  if (success) {
    isOpen.value = false
    emit('submit')
  }
}

const handleCancel = () => {
  isOpen.value = false
}
</script>

<template>
  <BaseModal 
    v-model="isOpen" 
    :title="isEditMode ? '編輯任務' : '建立任務'"
    size="md"
  >
    <template #body>
      <div class="space-y-4">
        <BaseFormGroup label="任務標題" required>
          <BaseInput
            v-model="formData.title"
            placeholder="請輸入任務標題"
            required
          />
        </BaseFormGroup>

        <BaseFormGroup label="描述">
          <BaseTextarea
            v-model="formData.description"
            placeholder="請輸入任務描述"
            :rows="3"
          />
        </BaseFormGroup>

        <div class="grid grid-cols-2 gap-4">
          <BaseFormGroup label="截止日期">
            <BaseInput
              v-model="formData.due_date"
              type="date"
            />
          </BaseFormGroup>

          <BaseFormGroup label="截止時間">
            <BaseInput
              v-model="formData.due_time"
              type="time"
            />
          </BaseFormGroup>
        </div>

        <BaseFormGroup label="提醒天數">
          <BaseInput
            v-model.number="formData.reminder_days"
            type="number"
          />
        </BaseFormGroup>
      </div>
    </template>

    <template #footer>
      <div class="flex items-center justify-end gap-2">
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

