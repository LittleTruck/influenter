<script setup lang="ts">
/**
 * BaseButton - 按鈕元件的基礎封裝
 * 封裝 UButton，提供統一的按鈕介面
 */

defineOptions({ inheritAttrs: false })

interface Props {
  /** 按鈕類型 */
  variant?: 'solid' | 'outline' | 'soft' | 'ghost' | 'link'
  /** 按鈕尺寸 */
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  /** 按鈕顏色 */
  color?: string
  /** 圖示 */
  icon?: string
  /** 後置圖示 */
  trailingIcon?: string
  /** 載入中 */
  loading?: boolean
  /** 禁用 */
  disabled?: boolean
  /** 區塊級按鈕 */
  block?: boolean
  /** 方形按鈕 */
  square?: boolean
  /** 文字 */
  label?: string
  /** 頭像 */
  avatar?: { src: string; alt: string }
  /** 自定義 UI 配置 */
  ui?: {
    base?: string
    size?: string
    variant?: string
    [key: string]: any
  }
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'solid',
  size: 'md',
  color: 'primary',
  loading: false,
  disabled: false,
  block: false,
  square: false,
})

const buttonUi = computed(() => {
  const baseUi = props.ui || {}
  return {
    ...baseUi,
    base: [
      'cursor-pointer disabled:cursor-not-allowed',
      baseUi.base
    ].filter(Boolean).join(' ')
  }
})
</script>

<template>
  <UButton
    :variant="variant"
    :size="size"
    :color="color"
    :icon="icon"
    :trailing-icon="trailingIcon"
    :loading="loading"
    :disabled="disabled"
    :block="block"
    :square="square"
    :label="label"
    :avatar="avatar"
    :ui="buttonUi"
    v-bind="$attrs"
  >
    <slot />
  </UButton>
</template>



