<script setup lang="ts">
import { BaseDashboardPanel, BaseDashboardNavbar } from '~/components/base'
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
      <BaseDashboardNavbar title="設定" />
    </template>
    <template #body>
      <div class="flex h-full bg-[#f6f5ff] dark:bg-[#0f0d1a]">
        <SubNav :items="navItems" label="設定" />
        <div class="flex-1 overflow-y-auto p-5 sm:p-6">
          <NuxtPage />
        </div>
      </div>
    </template>
  </BaseDashboardPanel>
</template>
