<script setup lang="ts">
import type { WorkflowTemplate } from '~/types/collaborationItems'
import { WORKFLOW_COLORS } from '~/utils/mockData'
import { BaseModal, BaseButton, BaseInput, BaseTextarea, BaseFormField } from '~/components/base'

interface Props {
  modelValue: boolean
  workflow?: WorkflowTemplate | null
}

const props = withDefaults(defineProps<Props>(), {
  workflow: null
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'submit': [data: { name: string; description?: string; color: string }]
}>()

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const isEditMode = computed(() => !!props.workflow)

const formData = reactive({
  name: '',
  description: '',
  color: 'primary'
})

// 初始化表單
watch(() => props.modelValue, (isOpen) => {
  if (isOpen) {
    if (props.workflow) {
      formData.name = props.workflow.name
      formData.description = props.workflow.description || ''
      formData.color = props.workflow.color
    } else {
      formData.name = ''
      formData.description = ''
      formData.color = 'primary'
    }
  }
}, { immediate: true })


const handleSubmit = () => {
  if (!formData.name.trim()) {
    return
  }
  emit('submit', {
    name: formData.name.trim(),
    description: formData.description.trim() || undefined,
    color: formData.color
  })
  isOpen.value = false
}
</script>

<template>
  <BaseModal
    v-model="isOpen"
    :title="isEditMode ? '編輯流程' : '新增流程'"
    size="md"
  >
    <template #body>
      <form class="space-y-4" @submit.prevent="handleSubmit">
        <BaseFormField label="流程名稱" name="name" required>
          <BaseInput
            v-model="formData.name"
            placeholder="請輸入流程名稱"
            class="w-full"
          />
        </BaseFormField>

        <BaseFormField label="描述" name="description">
          <BaseTextarea
            v-model="formData.description"
            placeholder="請輸入流程描述（選填）"
            :rows="3"
            class="w-full"
          />
        </BaseFormField>

        <BaseFormField label="顏色" name="color" required>
          <div class="flex items-center gap-2 flex-wrap">
            <div
              v-for="color in WORKFLOW_COLORS"
              :key="color.value"
              :class="[
                'w-8 h-8 rounded-full cursor-pointer border-2 transition-all',
                color.class,
                formData.color === color.value
                  ? 'border-gray-900 dark:border-white scale-110'
                  : 'border-transparent hover:scale-105'
              ]"
              :title="color.label"
              @click="formData.color = color.value"
            />
          </div>
        </BaseFormField>

        <div class="flex justify-end gap-2 pt-2">
          <BaseButton
            color="neutral"
            variant="ghost"
            @click="isOpen = false"
          >
            取消
          </BaseButton>
          <BaseButton
            type="submit"
          >
            {{ isEditMode ? '更新' : '建立' }}
          </BaseButton>
        </div>
      </form>
    </template>
  </BaseModal>
</template>

