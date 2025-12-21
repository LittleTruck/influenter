<script setup lang="ts">
import { BaseDashboardPanel, BaseDashboardNavbar, BaseDashboardSidebarCollapse, BaseIcon } from '~/components/base'

definePageMeta({
  middleware: 'auth'
})
</script>

<template>
  <BaseDashboardPanel>
    <template #header>
      <BaseDashboardNavbar title="日曆">
        <template #leading>
          <BaseDashboardSidebarCollapse />
        </template>
      </BaseDashboardNavbar>
    </template>

    <template #body>
      <div class="flex items-center justify-center h-64">
        <div class="text-center">
          <BaseIcon name="i-lucide-calendar" class="w-16 h-16 text-gray-400 mx-auto mb-4" />
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">即將推出</h2>
          <p class="text-gray-500 dark:text-gray-400">日曆功能正在開發中</p>
        </div>
      </div>
    </template>
  </BaseDashboardPanel>
</template>




