<script setup lang="ts">
/**
 * BaseDashboardSidebar - Dashboard 側邊欄的基礎封裝
 * 封裝 UDashboardSidebar，提供統一的側邊欄介面
 */

defineOptions({ inheritAttrs: false })

interface Props {
  /** 是否可摺疊 */
  collapsible?: boolean
  /** 是否可調整大小 */
  resizable?: boolean
  /** UI 樣式覆蓋 */
  ui?: Record<string, string>
}

const props = withDefaults(defineProps<Props>(), {
  collapsible: true,
  resizable: true,
  ui: undefined
})
</script>

<template>
  <UDashboardSidebar
    v-bind="$attrs"
    :collapsible="props.collapsible"
    :resizable="props.resizable"
    :ui="props.ui"
  >
    <template #header="slotProps">
      <slot name="header" v-bind="slotProps" />
    </template>

    <template #default="slotProps">
      <slot v-bind="slotProps" />
    </template>

    <template #footer="slotProps">
      <slot name="footer" v-bind="slotProps" />
    </template>
  </UDashboardSidebar>
</template>

