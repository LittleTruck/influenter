<template>
  <UDropdownMenu
    :items="items"
    :popper="{ strategy: 'absolute', placement: 'top' }"
    class="w-full"
  >
    <UButton
      :avatar="{
        src: authStore.user?.profile_picture_url || `https://api.dicebear.com/7.x/avataaars/svg?seed=${authStore.user?.email || 'Influenter'}`
      }"
      :label="collapsed ? undefined : (authStore.user?.name || authStore.user?.email || '使用者')"
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

const authStore = useAuthStore()
const { logout } = useAuth()

const handleLogout = async () => {
  await logout()
}

const items = computed(() => [
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
    onClick: handleLogout
  }]
])
</script>

