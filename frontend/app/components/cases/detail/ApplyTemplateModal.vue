<script setup lang="ts">
import type { ApplyTemplateRequest } from '~/types/cases'
import { BaseModal, BaseButton, BaseFormField, BaseSelect } from '~/components/base'
import { useWorkflowTemplates } from '~/composables/useWorkflowTemplates'

interface Props {
  /** 是否顯示 */
  modelValue: boolean
  /** 案件開始日期 */
  caseStartDate: string
  /** 案件 ID */
  caseId: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'submit': [data: ApplyTemplateRequest]
}>()

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const { workflowTemplates, fetchWorkflows } = useWorkflowTemplates()
const toast = useToast()

// 選中的流程
const selectedWorkflowId = ref<string>('')

// 流程選項
const workflowOptions = computed(() => {
  return workflowTemplates.value.map(w => ({
    label: w.name,
    value: w.id
  }))
})

// 處理套用
const handleApply = () => {
  if (!selectedWorkflowId.value) {
    toast.add({
      title: '請選擇流程',
      color: 'warning'
    })
    return
  }

  emit('submit', {
    workflow_id: selectedWorkflowId.value,
    start_date: props.caseStartDate
  } as ApplyTemplateRequest)

  isOpen.value = false
}

// 處理取消
const handleCancel = () => {
  selectedWorkflowId.value = ''
  isOpen.value = false
}

// 監聽 modal 打開，載入流程列表
watch(isOpen, (open) => {
  if (open) {
    fetchWorkflows()
    selectedWorkflowId.value = ''
  }
})
</script>

<template>
  <BaseModal
    v-model="isOpen"
    title="套用流程"
    description="選擇一個流程，將自動套用其階段到此案件"
    size="md"
  >
    <template #body>
      <BaseFormField label="選擇流程">
        <BaseSelect
          v-model="selectedWorkflowId"
          :options="workflowOptions"
          placeholder="請選擇流程"
          class="w-full"
        />
      </BaseFormField>
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
          :disabled="!selectedWorkflowId"
          @click="handleApply"
        >
          套用
        </BaseButton>
      </div>
    </template>
  </BaseModal>
</template>

