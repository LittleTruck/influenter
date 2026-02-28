<script setup lang="ts">
import { BaseButton, BaseIcon, BaseBadge, BaseDashboardPanel, BaseDashboardNavbar, BaseDashboardSidebarCollapse } from '~/components/base'
import AppSection from '~/components/ui/AppSection.vue'

const colorMode = useColorMode()

definePageMeta({
  middleware: 'auth'
})

const authStore = useAuthStore()
const emailsStore = useEmailsStore()
const toast = useToast()
const router = useRouter()

// 載入 Gmail 狀態
onMounted(async () => {
  await emailsStore.fetchGmailStatus()
})

// 觸發初次同步
const triggerFirstSync = async () => {
  try {
    await emailsStore.triggerSync()
    toast.add({
      title: '同步成功',
      description: '郵件已成功同步，正在更新列表...'
    })
    // 導航到郵件頁面
    setTimeout(() => {
      router.push('/emails')
    }, 1500)
  } catch (e: any) {
    toast.add({
      title: '同步失敗',
      description: e?.message || '同步過程中發生錯誤，請稍後再試',
      color: 'error'
    })
  }
}

// 斷開連接
const disconnectGmail = async () => {
  try {
    await emailsStore.disconnectGmail()
    toast.add({
      title: 'Gmail 已斷開連接'
    })
  } catch (e) {
    // error 已在 store 處理
  }
}

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

// 同步狀態顯示
const syncStatusText = computed(() => {
  const status = emailsStore.gmailStatus?.sync_status
  switch (status) {
    case 'active': return '正常同步中'
    case 'paused': return '已暫停'
    case 'error': return '同步錯誤'
    default: return '未知'
  }
})

const syncStatusColor = computed(() => {
  const status = emailsStore.gmailStatus?.sync_status
  switch (status) {
    case 'active': return 'success'
    case 'paused': return 'warning'
    case 'error': return 'error'
    default: return 'neutral'
  }
})
</script>

<template>
  <BaseDashboardPanel>
    <template #header>
      <BaseDashboardNavbar title="設定">
        <template #leading>
          <BaseDashboardSidebarCollapse />
        </template>
      </BaseDashboardNavbar>
    </template>

    <template #body>
      <div class="space-y-6">
        
        <!-- Gmail 整合 -->
        <AppSection>
          <template #header>
            <div class="flex items-center gap-3">
              <BaseIcon name="i-lucide-mail" class="w-5 h-5 text-primary" />
              <h2 class="text-lg font-semibold text-highlighted">Gmail 整合</h2>
            </div>
          </template>

          <div class="space-y-4">
            <!-- 連接狀態 -->
            <div class="flex items-center justify-between p-4 rounded-lg bg-elevated/50">
              <div class="flex items-center gap-3">
                <div 
                  class="w-10 h-10 rounded-full flex items-center justify-center"
                  :class="emailsStore.isConnected ? 'bg-primary-50 dark:bg-primary-900/20' : 'bg-muted'"
                >
                  <BaseIcon 
                    :name="emailsStore.isConnected ? 'i-lucide-check-circle' : 'i-lucide-circle'"
                    class="w-5 h-5"
                    :class="emailsStore.isConnected ? 'text-primary' : 'text-dimmed'"
                  />
                </div>
                <div>
                  <h3 class="font-medium text-highlighted">
                    {{ emailsStore.isConnected ? '已連接' : '未連接' }}
                  </h3>
                  <p class="text-sm text-muted">
                    {{ emailsStore.isConnected ? emailsStore.gmailStatus?.email : '尚未連接 Gmail 帳號' }}
                  </p>
                </div>
              </div>

              <BaseButton
                v-if="emailsStore.isConnected"
                color="error"
                variant="outline"
                size="sm"
                @click="disconnectGmail"
              >
                斷開連接
              </BaseButton>
            </div>

            <!-- 載入中 -->
            <div v-if="!emailsStore.isConnected && !emailsStore.gmailStatus" class="p-4 rounded-lg bg-elevated/50">
              <div class="flex items-center gap-3">
                <BaseIcon name="i-lucide-loader-2" class="w-5 h-5 text-primary animate-spin" />
                <div>
                  <p class="text-sm text-muted">正在檢查 Gmail 連接狀態...</p>
                </div>
              </div>
            </div>

            <!-- 未連接提示 -->
            <div v-else-if="!emailsStore.isConnected && emailsStore.gmailStatus" class="space-y-4">
              <div class="p-4 rounded-lg border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20">
                <div class="flex items-start gap-3">
                  <BaseIcon name="i-lucide-info" class="w-5 h-5 text-blue-500 dark:text-blue-400 mt-0.5" />
                  <div class="flex-1">
                    <h4 class="font-medium text-highlighted mb-1">Gmail 尚未連接</h4>
                    <p class="text-sm text-muted mb-3">
                      您已使用 Google 帳號登入，但需要重新授權以獲取 Gmail 存取權限。
                    </p>
                    <BaseButton
                      icon="i-lucide-mail"
                      color="primary"
                      @click="router.push('/auth/login')"
                    >
                      重新授權 Gmail
                    </BaseButton>
                  </div>
                </div>
              </div>
            </div>

            <!-- 同步資訊（已連接時顯示） -->
            <div v-if="emailsStore.isConnected" class="space-y-3">
              <div class="grid grid-cols-2 gap-4">
                <!-- 同步狀態 -->
                <div class="p-3 rounded-lg bg-elevated/50">
                  <div class="text-sm text-muted mb-1">同步狀態</div>
                  <BaseBadge :color="syncStatusColor" variant="subtle">
                    {{ syncStatusText }}
                  </BaseBadge>
                </div>

                <!-- 最後同步 -->
                <div class="p-3 rounded-lg bg-elevated/50">
                  <div class="text-sm text-muted mb-1">最後同步</div>
                  <div class="text-sm font-medium text-highlighted">
                    {{ formatDate(emailsStore.gmailStatus?.last_sync_at || null) }}
                  </div>
                </div>
              </div>

              <!-- 統計資訊 -->
              <div v-if="emailsStore.gmailStatus?.stats" class="p-4 rounded-lg bg-elevated/50">
                <h4 class="text-sm font-medium text-highlighted mb-3">郵件統計</h4>
                <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
                  <div>
                    <div class="text-xs text-muted mb-1">總郵件數</div>
                    <div class="text-lg font-semibold text-highlighted">
                      {{ emailsStore.gmailStatus.stats.total_messages?.toLocaleString() || 0 }}
                    </div>
                  </div>
                  <div>
                    <div class="text-xs text-muted mb-1">未讀郵件</div>
                    <div class="text-lg font-semibold text-highlighted">
                      {{ emailsStore.gmailStatus.stats.unread_messages?.toLocaleString() || 0 }}
                    </div>
                  </div>
                  <div>
                    <div class="text-xs text-muted mb-1">已加星號</div>
                    <div class="text-lg font-semibold text-highlighted">
                      {{ emailsStore.gmailStatus.stats.starred_messages?.toLocaleString() || 0 }}
                    </div>
                  </div>
                  <div>
                    <div class="text-xs text-muted mb-1">重要郵件</div>
                    <div class="text-lg font-semibold text-highlighted">
                      {{ emailsStore.gmailStatus.stats.important_messages?.toLocaleString() || 0 }}
                    </div>
                  </div>
                </div>
              </div>

              <!-- 同步錯誤（如果有） -->
              <div v-if="emailsStore.gmailStatus?.sync_error" 
                class="p-4 rounded-lg border border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/20"
              >
                <div class="flex items-start gap-3">
                  <BaseIcon name="i-lucide-alert-circle" class="w-5 h-5 text-red-600 dark:text-red-400 mt-0.5" />
                  <div class="flex-1">
                    <h4 class="font-medium text-red-600 dark:text-red-400 mb-1">同步錯誤</h4>
                    <p class="text-sm text-muted">
                      {{ emailsStore.gmailStatus.sync_error }}
                    </p>
                  </div>
                </div>
              </div>

              <!-- 手動同步按鈕 -->
              <div class="flex justify-end">
                <BaseButton
                  icon="i-lucide-refresh-cw"
                  color="primary"
                  variant="outline"
                  :loading="emailsStore.syncing"
                  :disabled="!emailsStore.canSync"
                  @click="triggerFirstSync"
                >
                  {{ emailsStore.syncing ? '同步中...' : '手動同步' }}
                </BaseButton>
              </div>
            </div>
          </div>
        </AppSection>

        <!-- 外觀設定 -->
        <AppSection>
          <template #header>
            <div class="flex items-center gap-3">
              <BaseIcon name="i-lucide-palette" class="w-5 h-5 text-primary" />
              <h2 class="text-lg font-semibold text-highlighted">外觀設定</h2>
            </div>
          </template>

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
        </AppSection>

        <!-- 帳號資訊 -->
        <AppSection>
          <template #header>
            <div class="flex items-center gap-3">
              <BaseIcon name="i-lucide-user" class="w-5 h-5 text-primary" />
              <h2 class="text-lg font-semibold text-highlighted">帳號資訊</h2>
            </div>
          </template>

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
                  {{ formatDate(authStore.user?.created_at || null) }}
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
        </AppSection>

      </div>
    </template>
  </BaseDashboardPanel>
</template>
