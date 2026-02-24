<script setup lang="ts">
import { BaseCard, BaseButton, BaseIcon, BaseBadge } from '~/components/base'
import AppSection from '~/components/ui/AppSection.vue'

definePageMeta({
  middleware: 'auth'
})

const authStore = useAuthStore()
const emailsStore = useEmailsStore()
const toast = useToast()
const router = useRouter()
const config = useRuntimeConfig()

// AI 注意事項
const aiInstructions = ref(authStore.user?.ai_instructions || '')
const savingAiInstructions = ref(false)

const saveAiInstructions = async () => {
  savingAiInstructions.value = true
  try {
    await $fetch(`${config.public.apiBase}/api/v1/auth/ai-instructions`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${authStore.token}` },
      body: { ai_instructions: aiInstructions.value || null },
    })
    // 同步到 store
    if (authStore.user) {
      authStore.user.ai_instructions = aiInstructions.value || undefined
    }
    toast.add({ title: '已儲存', description: 'AI 注意事項已更新' })
  } catch (e: any) {
    toast.add({ title: '儲存失敗', description: e?.message || '請稍後再試', color: 'error' })
  } finally {
    savingAiInstructions.value = false
  }
}

// 載入 Gmail 狀態
onMounted(async () => {
  await emailsStore.fetchGmailStatus()
})

// 連接 Gmail（透過 Google OAuth）
const connectGmail = () => {
  // 由於已經透過 Google 登入，OAuth tokens 應該已經在後端
  // 這裡我們需要呼叫一個 API 來授權 Gmail 存取
  // 目前系統設計是登入時就授予了 Gmail 權限
  toast.add({
    title: '提示',
    description: '您已透過 Google 登入，Gmail 連接已自動啟用',
    color: 'info'
  })
}

// 觸發初次同步
const triggerFirstSync = async () => {
  try {
    await emailsStore.triggerSync()
    alert('同步成功')
    toast.add({
      title: '同步成功',
      description: '郵件已成功同步，正在更新列表...'
    })
    // 導航到郵件頁面
    setTimeout(() => {
      router.push('/emails')
    }, 1500)
  } catch (e: any) {
    alert(`同步失敗: ${e?.message || ''}`)
    toast.add({
      title: '同步失敗',
      description: e.message || '同步過程中發生錯誤，請稍後再試',
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
  <div class="flex flex-col flex-1 h-full">
    <!-- Header -->
    <div class="flex items-center justify-between px-6 py-4 border-b border-default">
      <div class="flex items-center gap-4">
        <BaseButton
          icon="i-lucide-arrow-left"
          color="neutral"
          variant="ghost"
          @click="router.back()"
          aria-label="返回"
        />
        <h1 class="text-2xl font-bold text-highlighted">設定</h1>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto p-6">
      <div class="max-w-4xl mx-auto space-y-6">
        
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
                  :class="emailsStore.isConnected ? 'bg-primary-50 dark:bg-primary-900/20' : 'bg-gray-100 dark:bg-gray-800'"
                >
                  <BaseIcon 
                    :name="emailsStore.isConnected ? 'i-lucide-check-circle' : 'i-lucide-circle'"
                    class="w-5 h-5"
                    :class="emailsStore.isConnected ? 'text-primary' : 'text-gray-400'"
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

        <!-- AI 注意事項 -->
        <AppSection>
          <template #header>
            <div class="flex items-center gap-3">
              <BaseIcon name="i-lucide-brain" class="w-5 h-5 text-primary" />
              <h2 class="text-lg font-semibold text-highlighted">AI 注意事項</h2>
            </div>
          </template>

          <div class="space-y-4">
            <p class="text-sm text-muted">
              在此輸入您的合作注意事項（例如：修改規則、授權範圍、報價標準等），AI 擬信時會自動參考這些說明。
            </p>
            <textarea
              v-model="aiInstructions"
              class="w-full min-h-[120px] p-3 rounded-lg border border-default bg-elevated/50 text-highlighted placeholder-muted text-sm resize-y focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary"
              placeholder="例如：&#10;- 影片修改次數上限為 2 次&#10;- 不接受買斷授權&#10;- 報價以粉絲數 × 0.5 為基準"
            />
            <div class="flex justify-end">
              <BaseButton
                color="primary"
                :loading="savingAiInstructions"
                @click="saveAiInstructions"
              >
                儲存
              </BaseButton>
            </div>
          </div>
        </AppSection>

        <!-- 快速操作 -->
        <AppSection>
          <template #header>
            <div class="flex items-center gap-3">
              <BaseIcon name="i-lucide-zap" class="w-5 h-5 text-primary" />
              <h2 class="text-lg font-semibold text-highlighted">快速操作</h2>
            </div>
          </template>

          <div class="space-y-2">
            <BaseButton
              block
              color="neutral"
              variant="outline"
              icon="i-lucide-mail"
              to="/emails"
            >
              前往郵件管理
            </BaseButton>
            
            <BaseButton
              block
              color="neutral"
              variant="outline"
              icon="i-lucide-briefcase"
              to="/cases"
            >
              前往案件管理
            </BaseButton>
          </div>
        </AppSection>

      </div>
    </div>
  </div>
</template>
