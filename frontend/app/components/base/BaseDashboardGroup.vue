<script setup lang="ts">
/**
 * BaseDashboardGroup - Dashboard 群組的基礎封裝
 * 封裝 UDashboardGroup，提供統一的 Dashboard 群組介面
 */

defineOptions({ inheritAttrs: false })

interface Props {
  /** 單位（例如 'rem'） */
  unit?: string
}

const props = withDefaults(defineProps<Props>(), {
  unit: 'rem'
})
</script>

<template>
  <UDashboardGroup v-bind="$attrs" :unit="props.unit">
    <slot />
  </UDashboardGroup>
</template>
