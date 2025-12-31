<script setup lang="ts">
interface Props {
  /** 骨架項目數量 */
  count?: number
}

const props = withDefaults(defineProps<Props>(), {
  count: 5
})
</script>

<template>
  <div class="space-y-2">
    <div
      v-for="i in count"
      :key="i"
      class="flex items-center justify-between p-3 border border-gray-200 dark:border-gray-700 rounded-lg"
    >
      <!-- 左側內容骨架 -->
      <div class="flex items-center gap-3 flex-1">
        <div class="w-6 h-6 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
        <div class="space-y-2 flex-1">
          <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" :style="{ width: `${50 + (i * 5)}%` }" />
          <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" style="width: 30%" />
        </div>
      </div>

      <!-- 右側按鈕骨架 -->
      <div class="flex items-center gap-1 flex-shrink-0">
        <div class="w-6 h-6 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
        <div class="w-6 h-6 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
        <div class="w-6 h-6 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
      </div>
    </div>
  </div>
</template>




