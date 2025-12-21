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
}

const props = withDefaults(defineProps<Props>(), {
  orientation: 'vertical'
})
</script>

<template>
  <UNavigationMenu
    v-bind="$attrs"
    :items="props.items"
    :collapsed="props.collapsed"
    :orientation="props.orientation"
  >
    <slot />
  </UNavigationMenu>
</template>
