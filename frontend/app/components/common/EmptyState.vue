<script setup lang="ts">
interface Props {
  /** 圖示名稱 */
  icon?: string
  /** 標題 */
  title: string
  /** 描述 */
  description?: string
  /** 按鈕文字 */
  actionLabel?: string
  /** 是否顯示圖示背景 */
  showIconBackground?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  icon: 'i-lucide-inbox',
  showIconBackground: true
})

const emit = defineEmits<{
  'action': []
}>()

const handleAction = () => {
  emit('action')
}
</script>

<template>
  <div class="flex flex-col items-center justify-center py-12 px-4 text-center">
    <div
      v-if="showIconBackground"
      class="mb-4 p-6 rounded-full bg-gradient-to-br from-primary-100 to-primary-200 dark:from-primary-900/20 dark:to-primary-800/20"
    >
      <UIcon :name="icon" class="w-16 h-16 text-primary-500" />
    </div>
    <UIcon v-else :name="icon" class="w-12 h-12 mb-3 text-gray-400 dark:text-gray-500" />
    
    <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
      {{ title }}
    </h3>
    <p v-if="description" class="text-sm text-gray-500 dark:text-gray-400 mb-6 max-w-sm">
      {{ description }}
    </p>
    <UButton
      v-if="actionLabel"
      icon="i-lucide-plus"
      @click="handleAction"
    >
      {{ actionLabel }}
    </UButton>
  </div>
</template>



