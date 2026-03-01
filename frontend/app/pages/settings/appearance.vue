<script setup lang="ts">
import { BaseButton } from '~/components/base'
import SectionPageHeader from '~/components/ui/SectionPageHeader.vue'

const colorMode = useColorMode()

definePageMeta({
  middleware: 'auth'
})
</script>

<template>
  <div>
    <SectionPageHeader
      icon="i-lucide-palette"
      title="外觀設定"
      description="自訂介面顯示偏好"
    />

    <div class="space-y-4">
      <div class="flex items-center justify-between">
        <div>
          <h3 class="font-medium text-highlighted">深色模式</h3>
          <p class="text-sm text-muted">切換介面的淺色或深色主題</p>
        </div>
        <div class="flex items-center gap-1 bg-muted p-1 rounded-lg">
          <BaseButton
            icon="i-lucide-sun"
            :color="colorMode.preference === 'light' ? 'primary' : 'neutral'"
            :variant="colorMode.preference === 'light' ? 'solid' : 'ghost'"
            size="xs"
            @click="colorMode.preference = 'light'"
          >
            淺色
          </BaseButton>
          <BaseButton
            icon="i-lucide-moon"
            :color="colorMode.preference === 'dark' ? 'primary' : 'neutral'"
            :variant="colorMode.preference === 'dark' ? 'solid' : 'ghost'"
            size="xs"
            @click="colorMode.preference = 'dark'"
          >
            深色
          </BaseButton>
          <BaseButton
            icon="i-lucide-monitor"
            :color="colorMode.preference === 'system' ? 'primary' : 'neutral'"
            :variant="colorMode.preference === 'system' ? 'solid' : 'ghost'"
            size="xs"
            @click="colorMode.preference = 'system'"
          >
            系統
          </BaseButton>
        </div>
      </div>
    </div>
  </div>
</template>
