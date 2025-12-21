<template>
  <BaseDashboardGroup>
    <BaseDashboardSidebar collapsible resizable>
      <template #header="{ collapsed }">
        <div v-if="!collapsed" class="flex items-center gap-2">
          <BaseIcon name="i-lucide-sparkles" class="w-6 h-6" />
          <span class="font-bold text-lg">Influenter</span>
        </div>
        <BaseIcon v-else name="i-lucide-sparkles" class="w-6 h-6 mx-auto" />
      </template>

      <template #default="{ collapsed }">
        <BaseButton
          :label="collapsed ? undefined : '搜尋...'"
          icon="i-lucide-search"
          color="neutral"
          variant="outline"
          block
          :square="collapsed"
          class="mb-4"
        />

        <BaseNavigationMenu
          :collapsed="collapsed"
          :items="navigationItems"
          orientation="vertical"
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
import { BaseButton, BaseIcon, BaseDashboardGroup, BaseDashboardSidebar, BaseNavigationMenu } from '~/components/base'

const navigationItems: NavigationMenuItem[][] = [[{
  label: '首頁',
  icon: 'i-lucide-house',
  to: '/'
}, {
  label: '案件管理',
  icon: 'i-lucide-briefcase',
  to: '/cases',
  type: 'trigger',
  defaultOpen: false,
  children: [{
    label: '案件列表',
    to: '/cases',
    exact: true
  }, {
    label: '合作項目',
    to: '/cases/collaboration-items'
  }]
}, {
  label: '郵件',
  icon: 'i-lucide-mail',
  to: '/emails'
}, {
  label: '日曆',
  icon: 'i-lucide-calendar',
  to: '/calendar'
}], [{
  label: '設定',
  icon: 'i-lucide-settings',
  to: '/settings'
}]]
</script>


