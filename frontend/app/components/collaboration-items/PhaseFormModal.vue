<script setup lang="ts">
import type { CollaborationItemPhase, CreateCollaborationItemPhaseRequest, UpdateCollaborationItemPhaseRequest } from '~/types/collaborationItems'
import { BaseModal, BaseButton, BaseFormField, BaseInput } from '~/components/base'
import { z } from 'zod'

interface Props {
  /** 是否顯示 */
  modelValue: boolean
  /** 編輯的階段（如果有） */
  phase?: CollaborationItemPhase | null
  /** 合作項目 ID */
  collaborationItemId: string
}

const props = withDefaults(defineProps<Props>(), {
  phase: null
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'submit': [data: CreateCollaborationItemPhaseRequest | UpdateCollaborationItemPhaseRequest]
}>()

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const toast = useToast()

// 表單數據
const formData = reactive({
  name: '',
  duration_days: 1
})

// 表單驗證
const schema = z.object({
  name: z.string().min(1, '階段名稱不能為空'),
  duration_days: z.number().int().min(1, '天數必須大於 0')
})

const errors = ref<Record<string, string>>({})

// 初始化表單數據
watch(() => props.phase, (phase) => {
  if (phase) {
    formData.name = phase.name
    formData.duration_days = phase.duration_days
  } else {
    formData.name = ''
    formData.duration_days = 1
  }
  errors.value = {}
}, { immediate: true })

// 重置表單
const resetForm = () => {
  formData.name = ''
  formData.duration_days = 1
  errors.value = {}
}

// 驗證表單
const validate = (): boolean => {
  try {
    schema.parse(formData)
    errors.value = {}
    return true
  } catch (error) {
    if (error instanceof z.ZodError) {
      errors.value = {}
      error.errors.forEach((err) => {
        if (err.path[0]) {
          errors.value[err.path[0] as string] = err.message
        }
      })
    }
    return false
  }
}

// 處理提交
const handleSubmit = () => {
  if (!validate()) {
    toast.add({
      title: '請檢查表單錯誤',
      color: 'error'
    })
    return
  }

  if (props.phase) {
    // 更新模式
    emit('submit', {
      name: formData.name,
      duration_days: formData.duration_days
    } as UpdateCollaborationItemPhaseRequest)
  } else {
    // 新增模式
    emit('submit', {
      collaboration_item_id: props.collaborationItemId,
      name: formData.name,
      duration_days: formData.duration_days
    } as CreateCollaborationItemPhaseRequest)
  }

  resetForm()
  isOpen.value = false
}

// 處理取消
const handleCancel = () => {
  resetForm()
  isOpen.value = false
}

// 監聽 modal 關閉
watch(isOpen, (open) => {
  if (!open) {
    resetForm()
  }
})
</script>

<template>
  <BaseModal
    v-model="isOpen"
    :title="phase ? '編輯階段' : '新增階段'"
    :description="phase ? '修改階段資訊' : '為合作項目新增一個階段'"
  >
    <template #body>
      <div class="space-y-4">
        <!-- 階段名稱 -->
        <BaseFormField
          label="階段名稱"
          :error="errors.name"
          required
        >
          <BaseInput
            v-model="formData.name"
            placeholder="例如：腳本、影片A、文本"
            :disabled="!!phase"
            class="w-full"
          />
        </BaseFormField>

        <!-- 天數 -->
        <BaseFormField
          label="預設天數"
          :error="errors.duration_days"
          required
          description="此階段的預設執行天數"
        >
          <BaseInput
            v-model.number="formData.duration_days"
            type="number"
            min="1"
            max="365"
            placeholder="1"
            class="w-full"
          />
        </BaseFormField>
      </div>
    </template>

    <template #footer>
      <div class="flex justify-end gap-2">
        <BaseButton
          color="neutral"
          variant="outline"
          @click="handleCancel"
        >
          取消
        </BaseButton>
        <BaseButton
          color="primary"
          @click="handleSubmit"
        >
          {{ phase ? '更新' : '建立' }}
        </BaseButton>
      </div>
    </template>
  </BaseModal>
</template>

