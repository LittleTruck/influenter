<script setup lang="ts">
interface Props {
  /** 錯誤訊息 */
  message: string
  /** 標題 */
  title?: string
}

const props = withDefaults(defineProps<Props>(), {
  title: '載入失敗'
})
</script>

<template>
  <div class="flex flex-col items-center justify-center py-12 px-4 text-center">
    <div class="mb-4 p-4 rounded-full bg-red-100 dark:bg-red-900/20">
      <UIcon name="i-lucide-alert-circle" class="w-12 h-12 text-red-600 dark:text-red-400" />
    </div>
    <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
      {{ title }}
    </h3>
    <p class="text-sm text-gray-500 dark:text-gray-400 mb-6 max-w-sm">
      {{ message }}
    </p>
  </div>
</template>



