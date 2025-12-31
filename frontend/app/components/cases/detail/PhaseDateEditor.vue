<script setup lang="ts">
import type { CasePhase, UpdateCasePhaseRequest } from '~/types/cases'
import { BaseModal, BaseButton, BaseFormField, BaseInput } from '~/components/base'
import { format, addDays, parseISO } from 'date-fns'

interface Props {
  /** 是否顯示 */
  modelValue: boolean
  /** 編輯的階段 */
  phase: CasePhase | null
}

const props = withDefaults(defineProps<Props>(), {
  phase: null
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'submit': [data: UpdateCasePhaseRequest]
}>()

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const toast = useToast()

// 表單數據
const formData = reactive({
  start_date: '',
  duration_days: 1,
  end_date: ''
})

// 計算結束日期
const calculateEndDate = (startDate: string, days: number): string => {
  if (!startDate) return ''
  const start = parseISO(startDate)
  const end = addDays(start, days - 1) // 減1因為開始日算第一天
  return format(end, 'yyyy-MM-dd')
}

// 初始化表單數據
watch(() => props.phase, (phase) => {
  if (phase) {
    formData.start_date = phase.start_date
    formData.duration_days = phase.duration_days
    formData.end_date = calculateEndDate(phase.start_date, phase.duration_days)
  } else {
    formData.start_date = ''
    formData.duration_days = 1
    formData.end_date = ''
  }
}, { immediate: true })

// 監聽開始日期和天數變化，自動計算結束日期
watch([() => formData.start_date, () => formData.duration_days], ([startDate, days]) => {
  if (startDate && days > 0) {
    formData.end_date = calculateEndDate(startDate, days)
  } else {
    formData.end_date = ''
  }
})

// 處理提交
const handleSubmit = () => {
  if (!formData.start_date || formData.duration_days < 1) {
    toast.add({
      title: '請填寫完整資訊',
      color: 'error'
    })
    return
  }

  const endDate = calculateEndDate(formData.start_date, formData.duration_days)

  emit('submit', {
    start_date: formData.start_date,
    end_date: endDate,
    duration_days: formData.duration_days
  } as UpdateCasePhaseRequest)

  isOpen.value = false
}

// 處理取消
const handleCancel = () => {
  isOpen.value = false
}

// 監聽 modal 關閉
watch(isOpen, (open) => {
  if (!open && props.phase) {
    // 重置表單
    formData.start_date = props.phase.start_date
    formData.duration_days = props.phase.duration_days
    formData.end_date = calculateEndDate(props.phase.start_date, props.phase.duration_days)
  }
})
</script>

<template>
  <BaseModal
    v-model="isOpen"
    title="編輯階段日期"
    description="調整階段的開始日期和執行天數"
  >
    <template #body>
      <div class="space-y-4">
        <!-- 開始日期 -->
        <BaseFormField
          label="開始日期"
          required
        >
          <BaseInput
            v-model="formData.start_date"
            type="date"
          />
        </BaseFormField>

        <!-- 天數 -->
        <BaseFormField
          label="執行天數"
          required
          description="此階段的執行天數"
        >
          <BaseInput
            v-model.number="formData.duration_days"
            type="number"
            min="1"
            max="365"
          />
        </BaseFormField>

        <!-- 結束日期（自動計算，唯讀） -->
        <BaseFormField
          label="結束日期"
          description="自動根據開始日期和天數計算"
        >
          <BaseInput
            :model-value="formData.end_date"
            type="date"
            disabled
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
          儲存
        </BaseButton>
      </div>
    </template>
  </BaseModal>
</template>
