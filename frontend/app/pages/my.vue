<script setup lang="ts">
import { BaseDashboardPanel, BaseDashboardNavbar } from '~/components/base'
import SubNav from '~/components/ui/SubNav.vue'

const navItems = [
  { label: '合作項目', icon: 'i-lucide-package', to: '/my/collaboration-items' },
  { label: '流程範本', icon: 'i-lucide-list-checks', to: '/my/workflows' },
  { label: 'AI 助理設定', icon: 'i-lucide-brain', to: '/my/ai-instructions' },
]
</script>

<template>
  <BaseDashboardPanel :ui="{ body: 'p-0 sm:p-0' }">
    <template #header>
      <BaseDashboardNavbar title="合作管理" />
    </template>
    <template #body>
      <div class="flex h-full bg-[#f6f5ff] dark:bg-[#0f0d1a]">
        <SubNav :items="navItems" label="管理" />
        <div class="flex-1 overflow-y-auto p-5 sm:p-6">
          <NuxtPage />
        </div>
      </div>
    </template>
  </BaseDashboardPanel>
</template>
