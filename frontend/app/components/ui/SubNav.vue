<script setup lang="ts">
defineProps<{
  items: Array<{
    label: string
    icon: string
    to: string
  }>
}>()
</script>

<template>
  <nav class="py-4 space-y-1 px-3">
    <NuxtLink
      v-for="item in items"
      :key="item.to"
      :to="item.to"
      class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium text-muted hover:text-highlighted hover:bg-elevated/50 transition-colors"
      active-class="!text-primary bg-primary-50 dark:bg-primary-900/20 hover:!text-primary hover:bg-primary-50 dark:hover:bg-primary-900/20"
    >
      <UIcon :name="item.icon" class="w-4.5 h-4.5 shrink-0" />
      <span>{{ item.label }}</span>
    </NuxtLink>
  </nav>
</template>
