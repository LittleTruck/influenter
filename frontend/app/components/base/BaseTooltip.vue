<script setup lang="ts">
/**
 * BaseTooltip - Tooltip 元件的基礎封裝
 * 封裝 UTooltip，提供統一的提示框介面
 */

defineOptions({ inheritAttrs: false })

interface Props {
  /** 提示文字 */
  text?: string
  /** 位置 */
  placement?: 'top' | 'bottom' | 'left' | 'right'
}

const props = defineProps<Props>()
</script>

<template>
  <UTooltip v-bind="$attrs" :text="props.text" :placement="props.placement">
    <slot />
  </UTooltip>
</template>



