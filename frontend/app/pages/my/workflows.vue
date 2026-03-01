<script setup lang="ts">
import { BaseButton } from '~/components/base'
import SectionPageHeader from '~/components/ui/SectionPageHeader.vue'
import WorkflowManagement from '~/components/settings/collaboration-items/WorkflowManagement.vue'

definePageMeta({
  middleware: 'auth'
})

const workflowRef = ref<InstanceType<typeof WorkflowManagement> | null>(null)
</script>

<template>
  <div>
    <SectionPageHeader
      icon="i-lucide-list-checks"
      title="流程範本"
      description="建立可重複使用的合作流程，定義各階段名稱和預估天數"
    >
      <template #actions>
        <BaseButton
          icon="i-lucide-plus"
          size="sm"
          @click="workflowRef?.handleAddWorkflow()"
        >
          新增流程
        </BaseButton>
      </template>
    </SectionPageHeader>

    <WorkflowManagement ref="workflowRef" />
  </div>
</template>
