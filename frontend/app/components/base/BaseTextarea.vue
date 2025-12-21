<script setup lang="ts">
/**
 * BaseTextarea - 多行輸入框元件的基礎封裝
 * 封裝 UTextarea，提供統一的多行輸入框介面
 */

defineOptions({ inheritAttrs: false })

interface Props {
  /** 輸入值 */
  modelValue?: string
  /** 佔位符 */
  placeholder?: string
  /** 是否禁用 */
  disabled?: boolean
  /** 行數 */
  rows?: number
  /** 是否必填 */
  required?: boolean
  /** 尺寸 */
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
}

const props = withDefaults(defineProps<Props>(), {
  rows: 3,
  size: 'md',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const value = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val as string),
})
</script>

<template>
  <UTextarea
    v-bind="$attrs"
    v-model="value"
    :placeholder="placeholder"
    :disabled="disabled"
    :rows="rows"
    :required="required"
    :size="size"
    :class="[
      'base-textarea',
      disabled && 'base-textarea-disabled'
    ]"
  />
</template>

<style scoped>
.base-textarea {
  /* 可以在這裡添加全域多行輸入框樣式 */
}

.base-textarea-disabled {
  opacity: 0.5;
  cursor: not-allowed !important;
  background-color: rgb(243 244 246) !important; /* bg-gray-100 */
  border-color: rgb(209 213 219) !important; /* border-gray-300 */
}

.dark .base-textarea-disabled {
  background-color: rgb(31 41 55 / 0.6) !important; /* dark:bg-gray-800/60 */
  border-color: rgb(55 65 81) !important; /* dark:border-gray-700 */
}

.base-textarea-disabled :deep(textarea) {
  cursor: not-allowed !important;
  background-color: transparent !important;
}

.base-textarea-disabled :deep(textarea::placeholder) {
  opacity: 0.4;
}
</style>



