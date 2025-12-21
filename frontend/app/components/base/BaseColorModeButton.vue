<script setup lang="ts">
/**
 * BaseColorModeButton - 顏色模式切換按鈕的基礎封裝
 * 封裝 UColorModeButton，提供統一的顏色模式切換介面
 */

defineOptions({ inheritAttrs: false })

interface Props {
  /** 自訂 UI 覆寫 */
  ui?: Record<string, any>
}

const props = defineProps<Props>()
</script>

<template>
  <UColorModeButton
    v-bind="$attrs"
    :ui="props.ui"
  />
</template>
