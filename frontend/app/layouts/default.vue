<template>
  <BaseDashboardGroup>
    <BaseDashboardSidebar collapsible resizable >
      <template #header="{ collapsed }">
        <div v-if="!collapsed" class="flex items-center justify-between w-full">
          <div class="flex items-center gap-2">
            <BaseIcon name="i-lucide-sparkles" class="w-6 h-6" />
            <span class="font-bold text-lg">Influenter</span>
          </div>
          <BaseDashboardSidebarCollapse />
        </div>
        <BaseDashboardSidebarCollapse v-else />
      </template>

      <template #default="{ collapsed }">
        <BaseNavigationMenu
          :collapsed="collapsed"
          :items="navigationItems"
          orientation="vertical"
          tooltip
          popover
          :ui="{
            link: 'text-base hover:text-primary-500 dark:hover:text-primary-400',
            linkLeadingIcon: 'size-5 group-hover:text-primary-500 dark:group-hover:text-primary-400',
            childLabel: 'font-semibold text-xs text-muted uppercase tracking-wide px-2 py-1.5 border-b border-default mb-1',
            childLink: 'text-sm',
          }"
        />
      </template>

      <template #footer="{ collapsed }">
        <UserDropdown :collapsed="collapsed" />
      </template>
    </BaseDashboardSidebar>

    <slot />
  </BaseDashboardGroup>
</template>

<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'
import { BaseButton, BaseIcon, BaseDashboardGroup, BaseDashboardSidebar, BaseDashboardSidebarCollapse, BaseNavigationMenu } from '~/components/base'
import UserDropdown from '~/components/UserDropdown.vue'

const navigationItems: NavigationMenuItem[][] = [[{
  label: '首頁',
  icon: 'i-lucide-house',
  to: '/'
}, {
  label: '案件',
  icon: 'i-lucide-briefcase',
  to: '/cases',
  type: 'trigger',
  defaultOpen: false,
  children: [{
    label: '案件列表',
    to: '/cases',
    exact: true
  }, {
    label: '日曆',
    to: '/calendar'
  }]
}, {
  label: '郵件',
  icon: 'i-lucide-mail',
  to: '/emails'
}, {
  label: '合作管理',
  icon: 'i-lucide-handshake',
  to: '/my'
}], [{
  label: '設定',
  icon: 'i-lucide-settings',
  to: '/settings'
}]]
</script>


