<template>
  <UDropdownMenu
    :items="items"
    :popper="{ strategy: 'absolute', placement: 'top' }"
    class="w-full"
  >
    <UButton
      :avatar="{
        src: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Influenter'
      }"
      :label="collapsed ? undefined : '使用者'"
      color="neutral"
      variant="ghost"
      class="w-full"
      :block="collapsed"
    />
  </UDropdownMenu>
</template>

<script setup lang="ts">
defineProps<{
  collapsed?: boolean
}>()

const items = [
  [{
    label: '個人資料',
    icon: 'i-lucide-user',
    to: '/settings/profile'
  }, {
    label: 'AI 設定',
    icon: 'i-lucide-sparkles',
    to: '/settings/ai'
  }], [{
    label: '登出',
    icon: 'i-lucide-log-out',
    click: () => {
      // TODO: 實作登出邏輯
      console.log('登出')
    }
  }]
]
</script>

