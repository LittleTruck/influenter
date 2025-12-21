<script setup lang="ts">
/**
 * BaseAvatar - 頭像元件的基礎封裝
 * 封裝 UAvatar，提供統一的頭像介面
 */

defineOptions({ inheritAttrs: false })

interface Props {
  /** 頭像圖片 URL */
  src?: string
  /** 替代文字 */
  alt?: string
  /** 尺寸 */
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | '2xl' | '3xl'
  /** 圖示（當沒有圖片時） */
  icon?: string
  /** 文字（當沒有圖片時） */
  text?: string
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
})
</script>

<template>
  <UAvatar
    v-bind="$attrs"
    :src="src"
    :alt="alt"
    :size="size"
    :icon="icon"
    :text="text"
    class="base-avatar"
  />
</template>

<style scoped>
.base-avatar {
  /* 可以在這裡添加全域頭像樣式 */
}
</style>



