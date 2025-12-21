<script setup lang="ts">
/**
 * BaseSelect - 選擇器元件的基礎封裝
 * 封裝 USelectMenu，提供統一的選擇器介面
 */

defineOptions({ inheritAttrs: false })

interface Option {
  label: string
  value: string | number
  icon?: string
  disabled?: boolean
  [key: string]: any // 允許額外的屬性（如 config）
}

interface SelectUi {
  container?: string
  base?: string
  [key: string]: any
}

interface Props {
  /** 選中的值 */
  modelValue?: string | number | string[] | number[]
  /** 選項列表 */
  options?: Option[]
  /** items 列表（Nuxt UI 格式，如果有則優先使用） */
  items?: Array<{
    label?: string
    value?: string | number
    [key: string]: any
  }>
  /** value-key */
  valueKey?: string
  /** 佔位符 */
  placeholder?: string
  /** 是否多選 */
  multiple?: boolean
  /** 是否可搜尋 */
  searchable?: boolean
  /** 是否禁用 */
  disabled?: boolean
  /** 尺寸 */
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  /** 自定義 UI 配置 */
  ui?: SelectUi
}

const props = withDefaults(defineProps<Props>(), {
  multiple: false,
  searchable: false,
  size: 'md',
  valueKey: 'value'
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number | string[] | number[]]
}>()

const selected = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val as string | number | string[] | number[]),
})

// 如果提供了 items，直接使用；否則從 options 轉換
const items = computed(() => {
  if (props.items) return props.items
  if (!props.options) return []
  return props.options.map(opt => ({
    label: opt.label,
    value: opt.value,
    ...(opt.icon && { icon: opt.icon }),
    ...(opt.disabled !== undefined && { disabled: opt.disabled }),
    // 保留其他額外屬性
    ...Object.fromEntries(Object.entries(opt).filter(([key]) => !['label', 'value', 'icon', 'disabled'].includes(key)))
  }))
})
</script>

<template>
  <USelectMenu
    v-bind="$attrs"
    v-model="selected"
    :items="items"
    :value-key="valueKey"
    :placeholder="placeholder"
    :multiple="multiple"
    :search-input="searchable !== false"
    :disabled="disabled"
    :size="size"
    :ui="ui"
    :class="[
      'base-select',
      disabled && 'base-select-disabled'
    ]"
    @update:model-value="(val) => emit('update:modelValue', val)"
  >
    <slot />
  </USelectMenu>
</template>

<style scoped>
.base-select {
  /* 可以在這裡添加全域選擇器樣式 */
}

.base-select-disabled {
  opacity: 0.5;
  cursor: not-allowed !important;
}

.base-select-disabled :deep(button) {
  cursor: not-allowed !important;
  background-color: rgb(243 244 246) !important; /* bg-gray-100 */
  border-color: rgb(209 213 219) !important; /* border-gray-300 */
  opacity: 1;
}

.dark .base-select-disabled :deep(button) {
  background-color: rgb(31 41 55 / 0.6) !important; /* dark:bg-gray-800/60 */
  border-color: rgb(55 65 81) !important; /* dark:border-gray-700 */
}

.base-select-disabled :deep(button:hover),
.base-select-disabled :deep(button:focus) {
  background-color: rgb(243 244 246) !important; /* 防止 hover 效果 */
  border-color: rgb(209 213 219) !important;
}

.dark .base-select-disabled :deep(button:hover),
.dark .base-select-disabled :deep(button:focus) {
  background-color: rgb(31 41 55 / 0.6) !important;
  border-color: rgb(55 65 81) !important;
}
</style>



