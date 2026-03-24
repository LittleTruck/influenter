<script setup lang="ts">
import { BaseButton } from '~/components/base'
import SectionPageHeader from '~/components/ui/SectionPageHeader.vue'
import { THEME_COLORS, useThemeColor } from '~/composables/useThemeColor'
import type { ThemeColor } from '~/composables/useThemeColor'

const colorMode = useColorMode()
const { primaryColor, setPrimaryColor } = useThemeColor()

const colorClasses: Record<ThemeColor, string> = {
  green: 'bg-green-500',
  blue: 'bg-blue-500',
  violet: 'bg-violet-500',
  rose: 'bg-rose-500',
  amber: 'bg-amber-500',
  cyan: 'bg-cyan-500',
  indigo: 'bg-indigo-500',
  skyblue: '',
  rosepink: '',
}

const ringClasses: Record<ThemeColor, string> = {
  green: 'ring-green-500',
  blue: 'ring-blue-500',
  violet: 'ring-violet-500',
  rose: 'ring-rose-500',
  amber: 'ring-amber-500',
  cyan: 'ring-cyan-500',
  indigo: 'ring-indigo-500',
  skyblue: '',
  rosepink: '',
}

// 自訂色沒有 Tailwind utility，用 inline style
const customColorStyles: Partial<Record<ThemeColor, { bg: string; ring: string }>> = {
  skyblue: { bg: '#5693EC', ring: '#5693EC' },
  rosepink: { bg: '#F28ED4', ring: '#F28ED4' },
}

function getColorStyle(color: ThemeColor) {
  const custom = customColorStyles[color]
  if (!custom) return {}
  return { backgroundColor: custom.bg }
}

function getRingStyle(color: ThemeColor, isSelected: boolean) {
  const custom = customColorStyles[color]
  if (!custom || !isSelected) return {}
  return { boxShadow: `0 0 0 2px var(--color-white, #fff), 0 0 0 4px ${custom.ring}` }
}

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

    <div class="space-y-6">
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

      <div>
        <h3 class="font-medium text-highlighted">主題配色</h3>
        <p class="text-sm text-muted mb-3">選擇介面的主題顏色</p>
        <div class="flex items-center gap-3">
          <button
            v-for="color in THEME_COLORS"
            :key="color.value"
            :title="color.label"
            class="relative w-8 h-8 rounded-full cursor-pointer transition-transform hover:scale-110 focus:outline-none"
            :class="[
              colorClasses[color.value],
              primaryColor === color.value && ringClasses[color.value] ? `ring-2 ring-offset-2 ring-offset-white dark:ring-offset-gray-900 ${ringClasses[color.value]}` : ''
            ]"
            :style="{ ...getColorStyle(color.value), ...getRingStyle(color.value, primaryColor === color.value) }"
            @click="setPrimaryColor(color.value)"
          >
            <span
              v-if="primaryColor === color.value"
              class="absolute inset-0 flex items-center justify-center text-white"
            >
              <UIcon name="i-lucide-check" class="w-4 h-4" />
            </span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
