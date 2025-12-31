<script setup lang="ts">
import { BaseDashboardPanel, BaseDashboardNavbar, BaseDashboardSidebarCollapse, BaseButton } from '~/components/base'
import WorkflowManagement from '~/components/settings/collaboration-items/WorkflowManagement.vue'

definePageMeta({
  middleware: 'auth'
})
</script>

<template>
  <BaseDashboardPanel>
    <template #header>
      <BaseDashboardNavbar title="流程管理">
        <template #leading>
          <BaseDashboardSidebarCollapse />
        </template>

        <template #trailing>
          <BaseButton
            icon="i-lucide-arrow-left"
            variant="ghost"
            @click="navigateTo('/cases/collaboration-items')"
          >
            返回合作項目
          </BaseButton>
        </template>
      </BaseDashboardNavbar>
    </template>

    <template #body>
      <WorkflowManagement />
    </template>
  </BaseDashboardPanel>
</template>
