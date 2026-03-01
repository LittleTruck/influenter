<script setup lang="ts">
import { BaseDashboardPanel, BaseDashboardNavbar, BaseDashboardSidebarCollapse } from '~/components/base'
import SubNav from '~/components/ui/SubNav.vue'

const navItems = [
  { label: '帳號資訊', icon: 'i-lucide-user', to: '/settings/account' },
  { label: 'Gmail 整合', icon: 'i-lucide-mail', to: '/settings/gmail' },
  { label: '外觀設定', icon: 'i-lucide-palette', to: '/settings/appearance' },
]
</script>

<template>
  <BaseDashboardPanel :ui="{ body: 'p-0 sm:p-0' }">
    <template #header>
      <BaseDashboardNavbar title="設定">
        <template #leading>
          <BaseDashboardSidebarCollapse />
        </template>
      </BaseDashboardNavbar>
    </template>
    <template #body>
      <div class="flex h-full">
        <SubNav :items="navItems" class="w-56 shrink-0 border-r border-default" />
        <div class="flex-1 overflow-y-auto p-4 sm:p-6">
          <NuxtPage />
        </div>
      </div>
    </template>
  </BaseDashboardPanel>
</template>
