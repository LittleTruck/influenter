<script setup lang="ts">
import { BaseCard, BaseIcon, BaseCollapsible, BaseTooltip } from '~/components/base'

interface Props {
  /** 標題 */
  title: string
  /** 描述（會在 tooltip 中顯示） */
  description?: string
  /** 是否可收縮 */
  collapsible?: boolean
  /** 預設展開狀態（僅在 collapsible 為 true 時有效） */
  defaultOpen?: boolean
  /** 變體：card 或 plain */
  variant?: 'card' | 'plain'
}

const props = withDefaults(defineProps<Props>(), {
  collapsible: false,
  defaultOpen: true,
  variant: 'card'
})

const isOpen = ref(props.defaultOpen)
</script>

<template>
  <div>
    <BaseCollapsible
      v-if="collapsible"
      v-model:open="isOpen"
      :default-open="defaultOpen"
      class="collapsible-section"
    >
      <BaseCard v-if="variant === 'card'">
        <template #header>
          <div class="flex items-center justify-between w-full cursor-pointer" @click="isOpen = !isOpen">
            <div class="flex items-center gap-2">
              <h2 class="text-lg font-semibold">{{ title }}</h2>
              <BaseTooltip v-if="description" :text="description">
                <BaseIcon
                  name="i-lucide-info"
                  class="w-4 h-4 text-gray-400 dark:text-gray-500"
                />
              </BaseTooltip>
            </div>
            <div class="flex items-center gap-2" @click.stop>
              <slot name="actions" />
            </div>
          </div>
        </template>
        <slot />
        <template v-if="$slots.footer" #footer>
          <slot name="footer" />
        </template>
      </BaseCard>
      <div v-else class="rounded-md border border-gray-200 dark:border-neutral-700 p-4">
        <div class="flex items-center justify-between w-full cursor-pointer mb-4" @click="isOpen = !isOpen">
          <div class="flex items-center gap-2">
            <h2 class="text-lg font-semibold">{{ title }}</h2>
            <BaseTooltip v-if="description" :text="description">
              <BaseIcon
                name="i-lucide-info"
                class="w-4 h-4 text-gray-400 dark:text-gray-500"
              />
            </BaseTooltip>
          </div>
          <div class="flex items-center gap-2" @click.stop>
            <slot name="actions" />
          </div>
        </div>
        <div v-if="isOpen">
          <slot />
        </div>
        <div v-if="$slots.footer && isOpen" class="mt-4">
          <slot name="footer" />
        </div>
      </div>
    </BaseCollapsible>
    <BaseCard v-else-if="variant === 'card'">
      <template #header>
        <div class="flex items-center justify-between w-full">
          <div class="flex items-center gap-2">
            <h2 class="text-lg font-semibold">{{ title }}</h2>
            <BaseTooltip v-if="description" :text="description">
              <BaseIcon
                name="i-lucide-info"
                class="w-4 h-4 text-gray-400 dark:text-gray-500"
              />
            </BaseTooltip>
          </div>
          <div class="flex items-center gap-2">
            <slot name="actions" />
          </div>
        </div>
      </template>
      <slot />
      <template v-if="$slots.footer" #footer>
        <slot name="footer" />
      </template>
    </BaseCard>
    <div v-else class="rounded-md border border-gray-200 dark:border-neutral-700 p-4">
      <div class="flex items-center justify-between w-full mb-4">
        <div class="flex items-center gap-2">
          <h2 class="text-lg font-semibold">{{ title }}</h2>
          <BaseTooltip v-if="description" :text="description">
            <BaseIcon
              name="i-lucide-info"
              class="w-4 h-4 text-gray-400 dark:text-gray-500"
            />
          </BaseTooltip>
        </div>
        <div class="flex items-center gap-2">
          <slot name="actions" />
        </div>
      </div>
      <slot />
      <div v-if="$slots.footer" class="mt-4">
        <slot name="footer" />
      </div>
    </div>
  </div>
</template>

