<template>
  <BaseDashboardGroup>
    <BaseDashboardSidebar collapsible resizable >
      <template #header="{ collapsed }">
        <div v-if="!collapsed" class="flex items-center justify-between w-full">
          <div class="flex items-center gap-2">
            <span class="grid place-items-center w-6 h-6 rounded-[7px] bg-white/25">
              <BaseIcon name="i-lucide-sparkles" class="w-3.5 h-3.5 text-white" />
            </span>
            <span class="text-[13px] font-medium text-white">Influenter</span>
          </div>
          <BaseDashboardSidebarCollapse class="text-white/70 hover:text-white" />
        </div>
        <BaseDashboardSidebarCollapse v-else class="text-white/70 hover:text-white" />
      </template>

      <template #default="{ collapsed }">
        <BaseNavigationMenu
          :collapsed="collapsed"
          :items="navigationItems"
          orientation="vertical"
          tooltip
          popover
          :ui="{
            link: 'text-xs px-2.5 py-1.5 rounded-lg',
            linkLeadingIcon: 'size-4',
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

// Col 1 主導覽（依 redesign.md 扁平化：日曆提升為頂層）
const navigationItems: NavigationMenuItem[][] = [[{
  label: '首頁',
  icon: 'i-lucide-house',
  to: '/',
  exact: true
}, {
  label: '案件',
  icon: 'i-lucide-briefcase',
  to: '/cases'
}, {
  label: '日曆',
  icon: 'i-lucide-calendar-days',
  to: '/calendar'
}, {
  label: '郵件',
  icon: 'i-lucide-mail',
  to: '/emails'
}, {
  label: '合作管理',
  icon: 'i-lucide-handshake',
  to: '/my'
}, {
  label: '數據分析',
  icon: 'i-lucide-chart-column',
  to: '/analytics'
}], [{
  label: '設定',
  icon: 'i-lucide-settings',
  to: '/settings'
}]]
</script>


