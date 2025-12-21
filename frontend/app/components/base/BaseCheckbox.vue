<script setup lang="ts">
/**
 * BaseCheckbox - 複選框元件的基礎封裝
 * 封裝 UCheckbox，提供統一的複選框介面
 */

defineOptions({ inheritAttrs: false })

interface CheckboxUi {
  base?: string
  form?: string
  [key: string]: any
}

interface Props {
  /** 選中狀態 */
  modelValue?: boolean
  /** 標籤文字 */
  label?: string
  /** 是否禁用 */
  disabled?: boolean
  /** 是否必填 */
  required?: boolean
  /** 尺寸 */
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  /** 顏色 */
  color?: string
  /** 自定義 UI 配置 */
  ui?: Record<string, any>
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  disabled: false,
  required: false,
  size: 'md',
  color: 'primary'
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const checked = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})
</script>

<template>
  <UCheckbox
    v-bind="$attrs"
    v-model="checked"
    :label="label"
    :disabled="disabled"
    :required="required"
    :size="size"
    :color="color"
    :ui="ui"
    class="base-checkbox"
  >
    <slot />
  </UCheckbox>
</template>

<style scoped>
.base-checkbox {
  /* 可以在這裡添加全域複選框樣式 */
}
</style>



