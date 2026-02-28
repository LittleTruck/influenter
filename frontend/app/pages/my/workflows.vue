<script setup lang="ts">
import { BaseDashboardPanel, BaseDashboardNavbar, BaseDashboardSidebarCollapse } from '~/components/base'
import WorkflowManagement from '~/components/settings/collaboration-items/WorkflowManagement.vue'

definePageMeta({
  middleware: 'auth'
})
</script>

<template>
  <BaseDashboardPanel>
    <template #header>
      <BaseDashboardNavbar title="流程範本">
        <template #leading>
          <BaseDashboardSidebarCollapse />
        </template>
      </BaseDashboardNavbar>
    </template>

    <template #body>
      <WorkflowManagement />
    </template>
  </BaseDashboardPanel>
</template>
