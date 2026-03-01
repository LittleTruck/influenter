<script setup lang="ts">
import { BaseIcon } from '~/components/base'
import SectionPageHeader from '~/components/ui/SectionPageHeader.vue'

definePageMeta({
  middleware: 'auth'
})

const authStore = useAuthStore()

// 格式化日期
const formatDate = (dateString: string | null) => {
  if (!dateString) return '從未'
  const date = new Date(dateString)
  return date.toLocaleString('zh-TW', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<template>
  <div>
    <SectionPageHeader
      icon="i-lucide-user"
      title="帳號資訊"
      description="您的個人資料和帳號資訊"
    />

    <div class="space-y-4">
      <div class="flex items-center gap-4">
        <img
          v-if="authStore.user?.profile_picture_url"
          :src="authStore.user.profile_picture_url"
          :alt="authStore.user.name"
          class="w-16 h-16 rounded-full"
        />
        <div
          v-else
          class="w-16 h-16 rounded-full bg-primary-50 dark:bg-primary-900/20 flex items-center justify-center"
        >
          <BaseIcon name="i-lucide-user" class="w-8 h-8 text-primary" />
        </div>

        <div class="flex-1">
          <h3 class="font-medium text-highlighted">{{ authStore.user?.name }}</h3>
          <p class="text-sm text-muted">{{ authStore.user?.email }}</p>
        </div>
      </div>

      <div class="grid grid-cols-2 gap-4 pt-4 border-t border-default">
        <div>
          <div class="text-sm text-muted mb-1">註冊時間</div>
          <div class="text-sm font-medium text-highlighted">
            {{ formatDate(authStore.user?.createdAt || null) }}
          </div>
        </div>
        <div>
          <div class="text-sm text-muted mb-1">帳號 ID</div>
          <div class="text-xs font-mono text-highlighted">
            {{ authStore.user?.id }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
