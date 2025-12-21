<script setup lang="ts">
/**
 * BaseInput - 輸入框元件的基礎封裝
 * 封裝 UInput，提供統一的輸入框介面
 */

defineOptions({ inheritAttrs: false })

interface Props {
  /** 輸入值 */
  modelValue?: string | number
  /** 輸入類型 */
  type?: 'text' | 'email' | 'password' | 'number' | 'tel' | 'url' | 'color'
  /** 佔位符 */
  placeholder?: string
  /** 是否禁用 */
  disabled?: boolean
  /** 是否必填 */
  required?: boolean
  /** 圖示 */
  icon?: string
  /** 尺寸 */
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
}

const props = withDefaults(defineProps<Props>(), {
  type: 'text',
  size: 'md',
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
}>()

const value = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val as string | number),
})
</script>

<template>
  <UInput
    v-bind="$attrs"
    v-model="value"
    :type="type"
    :placeholder="placeholder"
    :disabled="disabled"
    :required="required"
    :icon="icon"
    :size="size"
    :class="[
      'base-input',
      disabled && 'base-input-disabled'
    ]"
  />
</template>

<style scoped>
.base-input {
  /* 可以在這裡添加全域輸入框樣式 */
}

.base-input-disabled {
  opacity: 0.5;
  cursor: not-allowed !important;
  background-color: rgb(243 244 246) !important; /* bg-gray-100 */
  border-color: rgb(209 213 219) !important; /* border-gray-300 */
}

.dark .base-input-disabled {
  background-color: rgb(31 41 55 / 0.6) !important; /* dark:bg-gray-800/60 */
  border-color: rgb(55 65 81) !important; /* dark:border-gray-700 */
}

.base-input-disabled :deep(input) {
  cursor: not-allowed !important;
  background-color: transparent !important;
}

.base-input-disabled :deep(input::placeholder) {
  opacity: 0.4;
}
</style>



