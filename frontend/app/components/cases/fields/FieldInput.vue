<script setup lang="ts">
import type { CaseField } from '~/types/fields'
import { useCaseFields } from '~/composables/useCaseFields'
import { BaseInput, BaseTextarea, BaseSelect, BaseBadge, BaseIcon, BaseCheckbox } from '~/components/base'
import type { FieldValue } from '~/types/fields'

interface Props {
  /** 屬性定義 */
  field: CaseField
  /** 屬性值 */
  modelValue: FieldValue
  /** 是否可編輯 */
  editable?: boolean
  /** 是否顯示標籤 */
  showLabel?: boolean
  /** 是否顯示錯誤訊息 */
  showError?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  editable: true,
  showLabel: false,
  showError: false
})

const emit = defineEmits<{
  'update:modelValue': [value: FieldValue]
}>()

const { validateFieldValue } = useCaseFields()

const localValue = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const errorMessage = computed(() => {
  if (!props.showError) return null
  return validateFieldValue(props.field, localValue.value)
})

const hasError = computed(() => !!errorMessage.value)

// 取得選項標籤
const getOptionLabel = (value: string | number) => {
  const option = props.field.options?.find(opt => opt.value === value)
  return option?.label || String(value)
}

// Select 選項轉換（Nuxt UI 4 使用 :items，格式為字串陣列或物件陣列）
const selectItems = computed(() => {
  if (!props.field.options) return []
  // 轉換為 Nuxt UI 4 的格式：{ label, value } 或直接字串
  return props.field.options.map(opt => ({
    label: opt.label,
    value: opt.value
  }))
})
</script>

<template>
  <div class="field-input">
    <label v-if="showLabel" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
      {{ field.label }}
      <sup v-if="field.is_required" class="ml-1 text-xs text-red-500 align-top leading-none">*</sup>
    </label>

    <!-- 文字輸入 -->
    <BaseInput
      v-if="field.type === 'text' && editable"
      v-model="localValue"
      :placeholder="field.placeholder || `請輸入${field.label}`"
      :required="field.is_required"
      class="w-full"
    />
    <div v-else-if="field.type === 'text'" class="text-sm text-gray-900 dark:text-white py-2">
      {{ localValue || '-' }}
    </div>

    <!-- Email 輸入 -->
    <BaseInput
      v-else-if="field.type === 'email' && editable"
      v-model="localValue"
      type="email"
      :placeholder="field.placeholder || `請輸入${field.label}`"
      :required="field.is_required"
      class="w-full"
    />
    <div v-else-if="field.type === 'email'" class="text-sm text-gray-900 dark:text-white py-2">
      {{ localValue || '-' }}
    </div>

    <!-- 電話輸入 -->
    <BaseInput
      v-else-if="field.type === 'phone' && editable"
      v-model="localValue"
      type="tel"
      :placeholder="field.placeholder || `請輸入${field.label}`"
      :required="field.is_required"
      class="w-full"
    />
    <div v-else-if="field.type === 'phone'" class="text-sm text-gray-900 dark:text-white py-2">
      {{ localValue || '-' }}
    </div>

    <!-- URL 輸入 -->
    <BaseInput
      v-else-if="field.type === 'url' && editable"
      v-model="localValue"
      type="url"
      :placeholder="field.placeholder || `請輸入${field.label}`"
      :required="field.is_required"
      class="w-full"
    />
    <div v-else-if="field.type === 'url'" class="text-sm text-gray-900 dark:text-white py-2">
      {{ localValue || '-' }}
    </div>

    <!-- 數字輸入 -->
    <BaseInput
      v-else-if="field.type === 'number' && editable"
      v-model.number="localValue"
      type="number"
      :placeholder="field.placeholder || `請輸入${field.label}`"
      :required="field.is_required"
      class="w-full"
    />
    <div v-else-if="field.type === 'number'" class="text-sm text-gray-900 dark:text-white py-2">
      {{ localValue || '-' }}
    </div>

    <!-- 日期輸入 -->
    <BaseInput
      v-else-if="field.type === 'date' && editable"
      v-model="localValue"
      type="date"
      :required="field.is_required"
      class="w-full"
    />
    <div v-else-if="field.type === 'date'" class="text-sm text-gray-900 dark:text-white py-2">
      {{ localValue ? new Date(localValue).toLocaleDateString('zh-TW') : '-' }}
    </div>

    <!-- 多行文字 -->
    <BaseTextarea
      v-else-if="field.type === 'textarea' && editable"
      v-model="localValue"
      :placeholder="field.placeholder || `請輸入${field.label}`"
      :rows="3"
      :required="field.is_required"
      class="w-full"
    />
    <div v-else-if="field.type === 'textarea'" class="text-sm text-gray-900 dark:text-white py-2 whitespace-pre-wrap">
      {{ localValue || '-' }}
    </div>

    <!-- 單選下拉 -->
    <BaseSelect
      v-else-if="field.type === 'select' && editable"
      v-model="localValue"
      :items="selectItems"
      :placeholder="field.placeholder || `請選擇${field.label}`"
      class="w-full"
    />
    <div v-else-if="field.type === 'select'" class="text-sm text-gray-900 dark:text-white py-2">
      <BaseBadge v-if="localValue" color="primary" variant="subtle" size="sm">
        {{ getOptionLabel(localValue) }}
      </BaseBadge>
      <span v-else class="text-gray-400">-</span>
    </div>

    <!-- 多選 -->
    <BaseSelect
      v-else-if="field.type === 'multiselect' && editable"
      v-model="localValue"
      :items="selectItems"
      multiple
      :placeholder="field.placeholder || `請選擇${field.label}`"
      class="w-full"
    />
    <div v-else-if="field.type === 'multiselect'" class="text-sm py-2">
      <div v-if="Array.isArray(localValue) && localValue.length > 0" class="flex flex-wrap gap-1">
        <BaseBadge
          v-for="(item, index) in localValue"
          :key="index"
          color="primary"
          variant="subtle"
          size="sm"
        >
          {{ getOptionLabel(item) }}
        </BaseBadge>
      </div>
      <span v-else class="text-gray-400">-</span>
    </div>

    <!-- 複選框 -->
    <BaseCheckbox
      v-else-if="field.type === 'checkbox' && editable"
      v-model="localValue"
      :label="field.label"
      :required="field.is_required"
    />
    <div v-else-if="field.type === 'checkbox'" class="text-sm py-2">
      <BaseIcon
        :name="localValue ? 'i-lucide-check-circle' : 'i-lucide-circle'"
        :class="localValue ? 'text-green-500' : 'text-gray-400'"
        class="w-4 h-4"
      />
    </div>

    <!-- 錯誤訊息 -->
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0 -translate-y-2"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 -translate-y-2"
    >
      <div v-if="errorMessage" class="text-sm text-red-500 mt-1">
        {{ errorMessage }}
      </div>
    </Transition>
  </div>
</template>

<style scoped>
</style>

