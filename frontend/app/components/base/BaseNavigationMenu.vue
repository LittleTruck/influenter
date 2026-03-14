<script setup lang="ts">
/**
 * BaseNavigationMenu - 導航選單的基礎封裝
 * 封裝 UNavigationMenu，提供統一的導航選單介面
 */

import type { NavigationMenuItem } from '@nuxt/ui'

defineOptions({ inheritAttrs: false })

interface Props {
  /** 導航項目 */
  items: NavigationMenuItem[][]
  /** 是否摺疊 */
  collapsed?: boolean
  /** 方向 */
  orientation?: 'vertical' | 'horizontal'
  /** 摺疊時是否顯示 tooltip */
  tooltip?: boolean
  /** 摺疊時是否顯示 popover（用於有子項目的導航） */
  popover?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  orientation: 'vertical',
  tooltip: false,
  popover: false,
})
</script>

<template>
  <UNavigationMenu
    v-bind="$attrs"
    :items="props.items"
    :collapsed="props.collapsed"
    :orientation="props.orientation"
    :tooltip="props.tooltip"
    :popover="props.popover"
  >
    <slot />
  </UNavigationMenu>
</template>

