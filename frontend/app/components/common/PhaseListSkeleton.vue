<script setup lang="ts">
interface Props {
  /** 骨架項目數量 */
  count?: number
}

const props = withDefaults(defineProps<Props>(), {
  count: 3
})
</script>

<template>
  <div class="space-y-3">
    <div
      v-for="i in count"
      :key="i"
      class="flex items-center gap-4 p-3 rounded-lg border border-gray-200 dark:border-gray-700"
    >
      <!-- 順序編號骨架 -->
      <div class="w-8 h-8 rounded-full bg-gray-200 dark:bg-gray-700 animate-pulse flex-shrink-0" />

      <!-- 內容骨架 -->
      <div class="flex-1 space-y-2">
        <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" :style="{ width: `${60 + (i * 10)}%` }" />
        <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" style="width: 40%" />
      </div>

      <!-- 操作按鈕骨架 -->
      <div class="flex items-center gap-2 flex-shrink-0">
        <div class="w-6 h-6 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
        <div class="w-6 h-6 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
      </div>
    </div>
  </div>
</template>
