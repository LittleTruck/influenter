<script setup lang="ts">
/**
 * BaseDropdownMenu - 下拉選單元件的基礎封裝
 * 封裝 UDropdownMenu，提供統一的下拉選單介面
 * 所有屬性與事件經由 $attrs 直透
 */

defineOptions({ inheritAttrs: false })
</script>

<template>
  <UDropdownMenu v-bind="$attrs">
    <slot />
  </UDropdownMenu>
</template>

<style scoped>
</style>



