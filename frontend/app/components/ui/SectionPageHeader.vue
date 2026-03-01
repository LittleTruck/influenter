<script setup lang="ts">
import { BaseIcon } from '~/components/base'

defineProps<{
  icon: string
  title: string
  description?: string
}>()
</script>

<template>
  <div class="mb-6 pb-6 border-b border-default">
    <div class="flex items-start justify-between gap-4">
      <div class="flex items-start gap-3">
        <BaseIcon :name="icon" class="w-6 h-6 text-primary mt-0.5 shrink-0" />
        <div>
          <h1 class="text-lg font-semibold text-highlighted">{{ title }}</h1>
          <p v-if="description" class="text-sm text-muted mt-1">{{ description }}</p>
        </div>
      </div>
      <div v-if="$slots.actions" class="shrink-0">
        <slot name="actions" />
      </div>
    </div>
  </div>
</template>
