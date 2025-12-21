<script setup lang="ts">
/**
 * BaseModal - 彈窗元件的基礎封裝
 * 封裝 UModal，提供統一的彈窗介面
 */

defineOptions({ inheritAttrs: false })

interface ModalUi {
  content?: string
  overlay?: string
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
  /** 自訂 UI 覆寫 */
  ui?: ModalUi
}

const widthBySize: Record<NonNullable<Props['size']>, string> = {
  sm: 'sm:max-w-md',
  md: 'sm:max-w-xl',
  lg: 'sm:max-w-2xl',
  xl: 'sm:max-w-4xl',
  full: 'sm:max-w-[90vw]'
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  modelValue: false
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

const modalUi = computed(() => {
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
  <UModal
    v-model:open="isOpen"
    v-bind="$attrs"
    :title="title"
    :description="description"
    :ui="modalUi"
  >
    <template #body>
      <slot name="body">
        <slot />
      </slot>
    </template>

    <template #footer>
      <slot name="footer" />
    </template>
  </UModal>
</template>



