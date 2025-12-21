<script setup lang="ts">
import type { CaseField } from '~/types/fields'
import { useCaseFields } from '~/composables/useCaseFields'
import { BaseBadge, BaseIcon } from '~/components/base'

interface Props {
  /** 屬性定義 */
  field: CaseField
  /** 屬性值 */
  value: any
  /** 是否顯示標籤 */
  showLabel?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  showLabel: false
})

const { formatFieldValue } = useCaseFields()

const displayValue = computed(() => {
  return formatFieldValue(props.field, props.value)
})

// 取得選項標籤
const getSelectLabel = (value: string | number | boolean): string => {
  const option = props.field.options?.find(opt => opt.value === value)
  return option?.label || String(value)
}
</script>

<template>
  <div class="field-renderer">
    <label v-if="showLabel" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
      {{ field.label }}
      <sup v-if="field.is_required" class="ml-1 text-xs text-red-500 align-top leading-none">*</sup>
    </label>

    <!-- 文字類型 -->
    <div v-if="field.type === 'text'" class="text-sm text-gray-900 dark:text-white">
      {{ displayValue }}
    </div>

    <!-- 數字類型 -->
    <div v-else-if="field.type === 'number'" class="text-sm text-gray-900 dark:text-white">
      {{ displayValue }}
    </div>

    <!-- 日期類型 -->
    <div v-else-if="field.type === 'date'" class="text-sm text-gray-900 dark:text-white">
      {{ displayValue }}
    </div>

    <!-- Email 類型 -->
    <div v-else-if="field.type === 'email'" class="text-sm">
      <a
        v-if="value"
        :href="`mailto:${value}`"
        class="text-primary-600 dark:text-primary-400 hover:underline"
      >
        {{ value }}
      </a>
      <span v-else class="text-gray-400">-</span>
    </div>

    <!-- URL 類型 -->
    <div v-else-if="field.type === 'url'" class="text-sm">
      <a
        v-if="value"
        :href="value"
        target="_blank"
        rel="noopener noreferrer"
        class="text-primary-600 dark:text-primary-400 hover:underline"
      >
        {{ value }}
      </a>
      <span v-else class="text-gray-400">-</span>
    </div>

    <!-- 電話類型 -->
    <div v-else-if="field.type === 'phone'" class="text-sm">
      <a
        v-if="value"
        :href="`tel:${value}`"
        class="text-primary-600 dark:text-primary-400 hover:underline"
      >
        {{ value }}
      </a>
      <span v-else class="text-gray-400">-</span>
    </div>

    <!-- 多行文字類型 -->
    <div v-else-if="field.type === 'textarea'" class="text-sm text-gray-900 dark:text-white whitespace-pre-wrap">
      {{ value || '-' }}
    </div>

    <!-- 單選類型 -->
    <div v-else-if="field.type === 'select'" class="text-sm">
      <BaseBadge v-if="value" color="primary" variant="subtle" size="sm">
        {{ getSelectLabel(value) }}
      </BaseBadge>
      <span v-else class="text-gray-400">-</span>
    </div>

    <!-- 多選類型 -->
    <div v-else-if="field.type === 'multiselect'" class="text-sm">
      <div v-if="Array.isArray(value) && value.length > 0" class="flex flex-wrap gap-1">
        <BaseBadge
          v-for="(item, index) in value"
          :key="index"
          color="primary"
          variant="subtle"
          size="sm"
        >
          {{ getSelectLabel(item) }}
        </BaseBadge>
      </div>
      <span v-else class="text-gray-400">-</span>
    </div>

    <!-- 複選框類型 -->
    <div v-else-if="field.type === 'checkbox'" class="text-sm">
      <BaseIcon
        :name="value ? 'i-lucide-check-circle' : 'i-lucide-circle'"
        :class="value ? 'text-green-500' : 'text-gray-400'"
        class="w-4 h-4"
      />
    </div>

    <!-- 預設顯示 -->
    <div v-else class="text-sm text-gray-900 dark:text-white">
      {{ displayValue }}
    </div>
  </div>
</template>

