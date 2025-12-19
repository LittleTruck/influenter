<script setup lang="ts">
interface Props {
  /** 載入訊息 */
  message?: string
  /** 尺寸 */
  size?: 'sm' | 'md' | 'lg'
}

const props = withDefaults(defineProps<Props>(), {
  message: '',
  size: 'md'
})

const iconSize = computed(() => {
  const sizes = {
    sm: 'w-6 h-6',
    md: 'w-8 h-8',
    lg: 'w-12 h-12'
  }
  return sizes[props.size]
})
</script>

<template>
  <div class="flex flex-col items-center justify-center py-12">
    <UIcon name="i-lucide-loader-2" :class="[iconSize, 'animate-spin text-primary-500']" />
    <p v-if="message" class="text-sm text-gray-500 dark:text-gray-400 mt-2">
      {{ message }}
    </p>
  </div>
</template>



