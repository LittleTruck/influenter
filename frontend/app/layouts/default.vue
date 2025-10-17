<template>
  <UDashboardGroup>
    <UDashboardSidebar collapsible resizable>
      <template #header="{ collapsed }">
        <div v-if="!collapsed" class="flex items-center gap-2">
          <UIcon name="i-lucide-sparkles" class="w-6 h-6 text-primary" />
          <span class="font-bold text-lg">Influenter</span>
        </div>
        <UIcon v-else name="i-lucide-sparkles" class="w-6 h-6 text-primary mx-auto" />
      </template>

      <template #default="{ collapsed }">
        <UButton
          :label="collapsed ? undefined : '搜尋...'"
          icon="i-lucide-search"
          color="neutral"
          variant="outline"
          block
          :square="collapsed"
          class="mb-4"
        />

        <UNavigationMenu
          :collapsed="collapsed"
          :items="navigationItems"
          orientation="vertical"
        />
      </template>

      <template #footer="{ collapsed }">
        <UserDropdown :collapsed="collapsed" />
      </template>
    </UDashboardSidebar>

    <slot />
  </UDashboardGroup>
</template>

<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'

const navigationItems: NavigationMenuItem[][] = [[{
  label: '首頁',
  icon: 'i-lucide-house',
  to: '/',
  active: true
}, {
  label: '案件管理',
  icon: 'i-lucide-briefcase',
  to: '/cases',
  badge: '即將推出'
}, {
  label: '郵件',
  icon: 'i-lucide-mail',
  to: '/emails'
}, {
  label: '日曆',
  icon: 'i-lucide-calendar',
  to: '/calendar',
  badge: '即將推出'
}], [{
  label: '設定',
  icon: 'i-lucide-settings',
  to: '/settings'
}]]
</script>


