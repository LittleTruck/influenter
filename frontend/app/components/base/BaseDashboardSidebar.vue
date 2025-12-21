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
}

const props = withDefaults(defineProps<Props>(), {
  collapsible: true,
  resizable: true
})
</script>

<template>
  <UDashboardSidebar
    v-bind="$attrs"
    :collapsible="props.collapsible"
    :resizable="props.resizable"
  >
    <slot />
  </UDashboardSidebar>
</template>
