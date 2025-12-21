<script setup lang="ts">
/**
 * BaseSlideover - Slideover 元件的基礎封裝
 * 封裝 USlideover，提供統一的 Slideover 介面
 */

defineOptions({ inheritAttrs: false })

interface SlideoverUi {
  content?: string
  overlay?: string
  header?: string
  body?: string
  footer?: string
  [key: string]: any
}

interface Props {
  /** 標題 */
  title?: string
  /** 描述 */
  description?: string
  /** 是否顯示 */
  modelValue?: boolean
  /** 預設寬度尺寸 */
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full'
  /** 從哪個方向滑入 */
  side?: 'left' | 'right' | 'top' | 'bottom'
  /** 是否顯示 overlay */
  overlay?: boolean
  /** 是否可關閉 */
  dismissible?: boolean
  /** 自訂 UI 覆寫 */
  ui?: SlideoverUi
}

const widthBySize: Record<NonNullable<Props['size']>, string> = {
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
  xl: 'max-w-xl',
  full: 'max-w-full'
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  side: 'right',
  modelValue: false,
  overlay: true,
  dismissible: true
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

const slideoverUi = computed(() => {
  const baseUi = props.ui ? { ...props.ui } : {}
  // 如果已經有自定義 content 樣式，就不添加 size 類
  if (!baseUi.content || !baseUi.content.includes('max-w')) {
    const sizeClass = widthBySize[props.size] ?? widthBySize.md
    baseUi.content = [sizeClass, baseUi.content ?? ''].filter(Boolean).join(' ')
  }
  return baseUi
})
</script>

<template>
  <USlideover
    v-model:open="isOpen"
    v-bind="$attrs"
    :title="title"
    :description="description"
    :side="side"
    :overlay="overlay"
    :dismissible="dismissible"
    :ui="slideoverUi"
  >
    <template #header>
      <slot name="header" />
    </template>

    <template #body>
      <slot name="body">
        <slot />
      </slot>
    </template>

    <template #footer>
      <slot name="footer" />
    </template>
  </USlideover>
</template>
