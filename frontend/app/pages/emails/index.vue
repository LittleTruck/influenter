<script setup lang="ts">
import type { Email, EmailDetail } from '~/stores/emails'
import { BaseButton, BaseInput, BaseIcon, BaseBadge, BaseAvatar, BaseDropdownMenu } from '~/components/base'

definePageMeta({
  middleware: 'auth'
})

const emailsStore = useEmailsStore()
const toast = useToast()
const router = useRouter()

// 當前選中的郵件
const selectedEmailId = ref<string | null>(null)
const selectedEmail = ref<EmailDetail | null>(null)
const loadingDetail = ref(false)

// 載入郵件和 Gmail 狀態
onMounted(async () => {
  await emailsStore.fetchGmailStatus()
  // 依目前分頁載入（預設收件匣）
  await emailsStore.fetchEmails({
    direction: filterDirection.value === 'all' ? undefined : filterDirection.value,
    page: 1
  })
  
  // 重新載入 Gmail 狀態以獲取最新的統計資料
  await emailsStore.fetchGmailStatus()
  
  // // 如果有郵件，自動選中第一封
  // if (emailsStore.emails.length > 0) {
  //   await selectEmail(emailsStore.emails[0].id)
  // }
  
  // 自動同步：如果 token 過期，自動觸發同步來刷新 token
  if (emailsStore.gmailStatus?.connected && emailsStore.gmailStatus?.token_expired) {
    try {
      console.log('Token expired, triggering auto-sync to refresh token...')
      // triggerSync 會自動刷新郵件列表和狀態
      await emailsStore.triggerSync()
      
      toast.add({
        title: 'Token 已刷新',
        description: '郵件同步已完成'
      })
    } catch (error: any) {
      // 靜默失敗，不顯示錯誤通知（避免打擾使用者）
      // 如果是冷卻期間，也不顯示錯誤
      if (!error.message?.includes('cooldown')) {
        console.warn('Auto-sync failed:', error)
      }
    }
  }
})

// 郵件列表容器 ref
const emailListRef = ref<HTMLElement | null>(null)

// 計算要顯示的分頁頁碼
const paginationPages = computed(() => {
  const current = emailsStore.pagination.page
  const total = emailsStore.pagination.total_pages
  const delta = 2 // 當前頁前後各顯示幾頁
  
  const pages: number[] = []
  for (let i = Math.max(1, current - delta); i <= Math.min(total, current + delta); i++) {
    pages.push(i)
  }
  
  return pages
})

// 處理分頁變更
const handlePageChange = async (page: number) => {
  try {
    await emailsStore.fetchEmails({ page })
    
    // 切換頁面後滾動到列表頂部
    if (emailListRef.value) {
      emailListRef.value.scrollTop = 0
    }
    // 清空選中的郵件
    selectedEmailId.value = null
    selectedEmail.value = null
  } catch (e) {
    console.error('Failed to change page:', e)
  }
}

// 選擇郵件
const selectEmail = async (emailId: string) => {
  // 保存當前滾動位置
  const scrollPosition = emailListRef.value?.scrollTop || 0
  
  // 恢復滾動位置的函數
  const restoreScroll = () => {
    if (emailListRef.value) {
      emailListRef.value.scrollTop = scrollPosition
    }
  }
  
  selectedEmailId.value = emailId
  loadingDetail.value = true
  
  try {
    await emailsStore.fetchEmail(emailId)
    selectedEmail.value = emailsStore.currentEmail
    
    // 標記為已讀（只對未讀郵件）
    if (selectedEmail.value && !selectedEmail.value.is_read) {
      // 立即恢復滾動位置
      restoreScroll()
      
      await emailsStore.markAsRead(emailId, true)
    }
    
    // 無論是否標記已讀，都要恢復滾動位置
    await nextTick()
    restoreScroll()
  } catch (e: any) {
    toast.add({
      title: '載入郵件失敗',
      description: e.message,
      color: 'error'
    })
  } finally {
    loadingDetail.value = false
  }
}

// 格式化時間
const formatTime = (dateStr: string) => {
  const date = new Date(dateStr)
  const now = new Date()
  const diffDays = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24))
  
  if (diffDays === 0) {
    return date.toLocaleTimeString('zh-TW', { hour: '2-digit', minute: '2-digit' })
  } else if (diffDays < 7) {
    return date.toLocaleDateString('zh-TW', { weekday: 'short' })
  } else {
    return date.toLocaleDateString('zh-TW', { month: 'short', day: 'numeric' })
  }
}

// 格式化完整日期
const formatFullDate = (dateStr: string) => {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-TW', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 加入為案件（AI 分析後建立案件）
const addingAsCase = ref(false)
const addAsCase = async () => {
  const emailId = selectedEmailId.value
  if (!emailId || !selectedEmail.value) return
  addingAsCase.value = true
  toast.add({ title: '正在建立案件，請稍候…', color: 'info' })
  try {
    const { caseId } = await emailsStore.createCaseFromEmail(emailId)
    await emailsStore.fetchEmail(emailId)
    selectedEmail.value = emailsStore.currentEmail
    toast.add({
      title: '已建立案件',
      color: 'success'
    })
    router.push(`/cases/${caseId}`)
  } catch (e: any) {
    toast.add({
      title: '建立案件失敗',
      description: e?.data?.message || e?.message || '請稍後再試',
      color: 'error'
    })
  } finally {
    addingAsCase.value = false
  }
}

// 使用郵件清理 composable
const { sanitizeHtml } = useEmailSanitizer()

// 顯示內容
const displayContent = computed(() => {
  if (!selectedEmail.value) return ''
  const content = selectedEmail.value.body_html || selectedEmail.value.body_text || selectedEmail.value.snippet || ''
  
  // 如果是 HTML 內容，進行清理
  if (selectedEmail.value.body_html) {
    return sanitizeHtml(content)
  }
  
  return content
})

const contentType = computed(() => {
  if (!selectedEmail.value) return 'text'
  return selectedEmail.value.body_html ? 'html' : 'text'
})

// 篩選相關
const filterIsRead = ref<'all' | 'read' | 'unread'>('all')
const filterDirection = ref<'all' | 'incoming' | 'outgoing'>('incoming') // 收件匣 / 寄件匣 / 全部
const searchQuery = ref('')

// 監聽篩選變更
watch([filterIsRead, filterDirection, searchQuery], async () => {
  const params: any = {}
  
  // 收件/寄件篩選
  if (filterDirection.value === 'incoming') {
    params.direction = 'incoming'
  } else if (filterDirection.value === 'outgoing') {
    params.direction = 'outgoing'
  } else {
    params.direction = undefined
  }
  
  // 根據篩選狀態設定 is_read 參數
  if (filterIsRead.value === 'unread') {
    params.is_read = false
  } else if (filterIsRead.value === 'all') {
    params.is_read = undefined
  }
  
  if (searchQuery.value) {
    params.subject = searchQuery.value
  }
  
  params.page = 1
  
  await emailsStore.fetchEmails(params)
})

// 刷新郵件列表
const refreshing = ref(false)
const refreshEmails = async () => {
  refreshing.value = true
  try {
    await emailsStore.fetchEmails()
    toast.add({
      title: '郵件列表已更新',
      icon: 'i-lucide-check-circle'
    })
  } catch (e: any) {
    toast.add({
      title: '更新失敗',
      description: e.message || '無法載入郵件列表',
      color: 'error',
      icon: 'i-lucide-alert-circle'
    })
  } finally {
    refreshing.value = false
  }
}

// 觸發同步
const handleSync = async () => {
  try {
    await emailsStore.triggerSync()
    toast.add({
      title: '同步成功',
      description: '郵件已成功同步，列表已更新',
      icon: 'i-lucide-check-circle'
    })
  } catch (e: any) {
    toast.add({
      title: '同步失敗',
      description: e.message || '同步過程中發生錯誤，請稍後再試',
      color: 'error',
      icon: 'i-lucide-alert-circle'
    })
  }
}

// 標記已讀/未讀
const toggleRead = async (email: EmailDetail, isRead: boolean) => {
  try {
    await emailsStore.markAsRead(email.id, isRead)
    if (selectedEmail.value) {
      selectedEmail.value.is_read = isRead
    }
    toast.add({
      title: isRead ? '已標記為已讀' : '已標記為未讀',
      color: 'success',
      icon: 'i-lucide-check-circle'
    })
  } catch (e: any) {
    toast.add({
      title: '操作失敗',
      color: 'error'
    })
  }
}

</script>

<template>
  <div class="flex flex-col flex-1 h-full">
    <!-- Header -->
    <div class="flex items-center justify-between px-6 py-4 border-b border-default">
      <div class="flex items-center gap-3">
        <h1 class="text-2xl font-bold text-highlighted">郵件</h1>
        
        <!-- 未讀數量標記 (參考 template) -->
        <BaseBadge
          size="sm"
          :color="emailsStore.unreadCount > 0 ? 'primary' : 'neutral'"
          variant="solid"
        >
          {{ emailsStore.unreadCount }}
        </BaseBadge>
      </div>

      <div class="flex items-center gap-2">
        <!-- 同步按鈕 -->
        <BaseButton
          v-if="emailsStore.isConnected"
          icon="i-lucide-refresh-cw"
          color="neutral"
          variant="outline"
          size="sm"
          :loading="emailsStore.syncing"
          :disabled="!emailsStore.canSync"
          @click="handleSync"
        >
          {{ emailsStore.syncing ? '同步中...' : '同步郵件' }}
        </BaseButton>

        <!-- 重新整理 -->
        <BaseButton
          color="neutral"
          variant="ghost"
          size="sm"
          :loading="refreshing"
          :disabled="refreshing"
          @click="refreshEmails"
          aria-label="重新整理"
        >
          <template #leading>
            <BaseIcon 
              name="i-lucide-rotate-cw" 
              :class="{ 'animate-spin': refreshing }"
            />
          </template>
        </BaseButton>
      </div>
    </div>

    <!-- Split View Container -->
    <div class="flex flex-1 overflow-hidden">
      <!-- Left Side: Email List -->
      <div class="w-96 flex flex-col border-r border-default">
        <!-- Search & Filters -->
        <div class="p-4 border-b border-default bg-elevated/50 space-y-3">
          <!-- 收件匣 / 寄件匣 / 全部 -->
          <div class="flex items-center gap-0.5 bg-gray-100 dark:bg-gray-800 rounded-lg p-0.5">
            <BaseButton
              :variant="filterDirection === 'incoming' ? 'solid' : 'ghost'"
              :color="filterDirection === 'incoming' ? 'primary' : 'neutral'"
              size="sm"
              @click="filterDirection = 'incoming'"
            >
              收件匣
            </BaseButton>
            <BaseButton
              :variant="filterDirection === 'outgoing' ? 'solid' : 'ghost'"
              :color="filterDirection === 'outgoing' ? 'primary' : 'neutral'"
              size="sm"
              icon="i-lucide-send"
              @click="filterDirection = 'outgoing'"
            >
              寄件匣
            </BaseButton>
            <BaseButton
              :variant="filterDirection === 'all' ? 'solid' : 'ghost'"
              :color="filterDirection === 'all' ? 'primary' : 'neutral'"
              size="sm"
              @click="filterDirection = 'all'"
            >
              全部
            </BaseButton>
          </div>
          <div class="flex items-center gap-2">
            <BaseInput
              v-model="searchQuery"
              icon="i-lucide-search"
              placeholder="搜尋主旨..."
              size="sm"
              class="flex-1"
            />
            <div class="flex items-center gap-0.5 bg-gray-100 dark:bg-gray-800 rounded-lg p-0.5">
              <BaseButton
                :variant="filterIsRead === 'all' ? 'solid' : 'ghost'"
                :color="filterIsRead === 'all' ? 'primary' : 'neutral'"
                size="sm"
                @click="filterIsRead = 'all'"
              >
                全部
              </BaseButton>
              <BaseButton
                :variant="filterIsRead === 'unread' ? 'solid' : 'ghost'"
                :color="filterIsRead === 'unread' ? 'primary' : 'neutral'"
                size="sm"
                @click="filterIsRead = 'unread'"
              >
                未讀
              </BaseButton>
            </div>
          </div>
        </div>

        <!-- Email List -->
        <div ref="emailListRef" class="flex-1 overflow-y-auto">
          <!-- Loading State -->
          <div v-if="emailsStore.loading" class="flex items-center justify-center py-12">
            <BaseIcon name="i-lucide-loader-2" class="w-6 h-6 animate-spin text-primary" />
          </div>

          <!-- Empty State -->
          <div 
            v-else-if="emailsStore.emails.length === 0"
            class="flex flex-col items-center justify-center py-12 px-4 text-center"
          >
            <BaseIcon 
              :name="filterDirection === 'outgoing' ? 'i-lucide-send' : 'i-lucide-inbox'" 
              class="w-12 h-12 mb-3 text-muted" 
            />
            <h3 class="text-sm font-semibold text-highlighted mb-1">
              {{ emailsStore.isConnected 
                ? (filterDirection === 'outgoing' ? '目前沒有寄出的信件' : '目前沒有郵件') 
                : '尚未連接 Gmail' 
              }}
            </h3>
            <p class="text-xs text-muted">
              {{ emailsStore.isConnected 
                ? (filterDirection === 'outgoing' ? '從案件或郵件詳情頁寄出回信後會顯示於此' : '嘗試同步郵件或調整篩選條件')
                : '請先在設定中連接您的 Gmail 帳號' 
              }}
            </p>
          </div>

          <!-- Email Items -->
          <div v-else class="divide-y divide-default">
            <div
              v-for="email in emailsStore.emails"
              :key="email.id"
              class="p-4 cursor-pointer transition-colors hover:bg-gray-50 dark:hover:bg-gray-800/50"
              :class="{
                'bg-primary-50 dark:bg-primary-900/20 border-l-2 border-l-primary-500': selectedEmailId === email.id
              }"
              @click="selectEmail(email.id)"
            >
              <div class="flex items-start gap-3">
                <!-- 未讀指示器 (小綠點，參考 template) -->
                <div class="relative">
                  <BaseAvatar
                    :alt="email.from_name || email.from_email"
                    size="sm"
                  />
                  <!-- 未讀小綠點 -->
                  <span 
                    v-if="!email.is_read"
                    class="absolute -top-0.5 -right-0.5 w-2.5 h-2.5 bg-primary-500 rounded-full border-2 border-white dark:border-gray-900"
                  />
                </div>
                
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2 mb-1">
                    <span 
                      class="text-sm truncate"
                      :class="email.is_read ? 'text-muted font-normal' : 'text-highlighted font-semibold'"
                    >
                      {{ email.direction === 'outgoing' ? `寄給 ${email.to_email || '—'}` : (email.from_name || email.from_email) }}
                    </span>
                    <span class="text-xs text-muted shrink-0">
                      {{ formatTime(email.received_at) }}
                    </span>
                  </div>
                  
                  <div 
                    class="text-sm mb-1 truncate"
                    :class="email.is_read ? 'text-muted' : 'text-highlighted font-medium'"
                  >
                    {{ email.subject || '(無主旨)' }}
                  </div>
                  
                  <div class="text-xs text-muted truncate">
                    {{ email.snippet }}
                  </div>

                  <!-- Badges -->
                  <div v-if="email.has_attachments || email.labels?.includes('STARRED') || email.case_id" class="flex items-center gap-1 mt-2">
                    <BaseBadge
                      v-if="email.has_attachments"
                      size="xs"
                      color="neutral"
                      variant="subtle"
                      icon="i-lucide-paperclip"
                    />
                    <BaseBadge
                      v-if="email.labels?.includes('STARRED')"
                      size="xs"
                      color="warning"
                      variant="subtle"
                      icon="i-lucide-star"
                    />
                    <BaseBadge
                      v-if="email.case_id"
                      size="xs"
                      color="info"
                      variant="subtle"
                    >
                      已關聯
                    </BaseBadge>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Pagination -->
        <div 
          v-if="emailsStore.pagination.total_pages > 1" 
          class="p-3 border-t border-default flex justify-center gap-1"
        >
          <BaseButton
            icon="i-lucide-chevrons-left"
            size="sm"
            color="neutral"
            variant="ghost"
            :disabled="emailsStore.pagination.page === 1"
            @click="handlePageChange(1)"
            aria-label="第一頁"
          />
          <BaseButton
            icon="i-lucide-chevron-left"
            size="sm"
            color="neutral"
            variant="ghost"
            :disabled="emailsStore.pagination.page === 1"
            @click="handlePageChange(emailsStore.pagination.page - 1)"
            aria-label="上一頁"
          />
          
          <template v-for="page in paginationPages" :key="page">
            <BaseButton
              size="sm"
              :color="page === emailsStore.pagination.page ? 'primary' : 'neutral'"
              :variant="page === emailsStore.pagination.page ? 'solid' : 'ghost'"
              @click="handlePageChange(page)"
            >
              {{ page }}
            </BaseButton>
          </template>
          
          <BaseButton
            icon="i-lucide-chevron-right"
            size="sm"
            color="neutral"
            variant="ghost"
            :disabled="emailsStore.pagination.page === emailsStore.pagination.total_pages"
            @click="handlePageChange(emailsStore.pagination.page + 1)"
            aria-label="下一頁"
          />
          <BaseButton
            icon="i-lucide-chevrons-right"
            size="sm"
            color="neutral"
            variant="ghost"
            :disabled="emailsStore.pagination.page === emailsStore.pagination.total_pages"
            @click="handlePageChange(emailsStore.pagination.total_pages)"
            aria-label="最後一頁"
          />
        </div>
      </div>

      <!-- Right Side: Email Detail -->
      <div class="flex-1 flex flex-col overflow-hidden">
        <!-- Loading State -->
        <div v-if="loadingDetail" class="flex items-center justify-center h-full">
          <BaseIcon name="i-lucide-loader-2" class="w-8 h-8 animate-spin text-primary" />
        </div>

        <!-- Empty State -->
        <div 
          v-else-if="!selectedEmail"
          class="flex flex-col items-center justify-center h-full text-center px-4"
        >
          <BaseIcon name="i-lucide-mail" class="w-16 h-16 mb-4 text-muted" />
          <h3 class="text-lg font-semibold text-highlighted mb-2">選擇一封郵件</h3>
          <p class="text-muted">從左側列表選擇郵件以查看內容</p>
        </div>

        <!-- Email Detail -->
        <div v-else class="flex-1 flex flex-col overflow-hidden">
          <!-- Detail Header -->
          <div class="p-6 border-b border-default">
            <div class="flex items-center justify-between mb-4">
              <h2 class="text-xl font-bold text-highlighted flex-1">
                {{ selectedEmail.subject || '(無主旨)' }}
              </h2>

              <div class="flex items-center gap-2">
                <BaseButton
                  v-if="selectedEmail.case_id"
                  icon="i-lucide-external-link"
                  color="primary"
                  size="sm"
                  @click="router.push(`/cases/${selectedEmail.case_id}`)"
                >
                  顯示案件
                </BaseButton>
                <BaseButton
                  v-else
                  icon="i-lucide-briefcase"
                  color="primary"
                  size="sm"
                  :loading="addingAsCase"
                  :disabled="addingAsCase"
                  @click="addAsCase"
                >
                  加入為案件
                </BaseButton>
                <BaseButton
                  :icon="selectedEmail.is_read ? 'i-lucide-mail' : 'i-lucide-mail-open'"
                  color="neutral"
                  variant="ghost"
                  size="sm"
                  @click="toggleRead(selectedEmail, !selectedEmail.is_read)"
                >
                  {{ selectedEmail.is_read ? '標記未讀' : '標記已讀' }}
                </BaseButton>

                <BaseDropdownMenu
                  :items="[[
                    {
                      label: '產生回覆',
                      icon: 'i-lucide-reply',
                      onSelect: () => {
                        toast.add({ title: '功能開發中', color: 'info' })
                      }
                    }
                  ]]"
                >
                  <BaseButton
                    icon="i-lucide-ellipsis-vertical"
                    color="neutral"
                    variant="ghost"
                    size="sm"
                  />
                </BaseDropdownMenu>
              </div>
            </div>

            <!-- Sender/Recipient Info -->
            <div class="flex items-start gap-3">
              <BaseAvatar
                :alt="(selectedEmail as any).direction === 'outgoing' ? (selectedEmail.to_email || '') : (selectedEmail.from_name || selectedEmail.from_email)"
                size="lg"
              />
              <div class="flex-1">
                <div class="flex items-center gap-2 mb-1">
                  <span class="font-semibold text-highlighted">
                    {{ (selectedEmail as any).direction === 'outgoing' ? `寄給 ${selectedEmail.to_email || '—'}` : (selectedEmail.from_name || selectedEmail.from_email) }}
                  </span>
                  <BaseBadge
                    v-if="(selectedEmail as any).direction === 'outgoing'"
                    size="xs"
                    color="neutral"
                    variant="subtle"
                  >
                    已寄出
                  </BaseBadge>
                  <BaseBadge
                    v-else-if="!selectedEmail.is_read"
                    size="xs"
                    color="primary"
                  >
                    未讀
                  </BaseBadge>
                </div>
                <div class="text-sm text-muted">
                  {{ (selectedEmail as any).direction === 'outgoing' ? selectedEmail.to_email : selectedEmail.from_email }}
                </div>
                <div class="text-xs text-muted mt-1">
                  {{ formatFullDate(selectedEmail.received_at) }}
                </div>
              </div>
            </div>

            <!-- Labels -->
            <div 
              v-if="selectedEmail.labels && selectedEmail.labels.length > 0" 
              class="flex flex-wrap gap-2 mt-4"
            >
              <BaseBadge
                v-for="label in selectedEmail.labels"
                :key="label"
                size="sm"
                color="neutral"
              >
                {{ label }}
              </BaseBadge>
            </div>

            <!-- Badges -->
            <div class="flex items-center gap-2 mt-4">
              <BaseBadge
                v-if="selectedEmail.has_attachments"
                icon="i-lucide-paperclip"
                color="neutral"
                size="sm"
              >
                有附件
              </BaseBadge>
              <BaseBadge
                v-if="selectedEmail.ai_analyzed"
                icon="i-lucide-sparkles"
                color="success"
                size="sm"
              >
                已AI分析
              </BaseBadge>
            </div>
          </div>

          <!-- Detail Content -->
          <div class="flex-1 overflow-y-auto p-6">
            <div class="max-w-4xl mx-auto">
              <div class="prose prose-sm max-w-none dark:prose-invert">
                <!-- HTML Content -->
                <div
                  v-if="contentType === 'html'"
                  v-html="displayContent"
                  class="email-content isolate"
                />
                
                <!-- Plain Text Content -->
                <pre
                  v-else
                  class="whitespace-pre-wrap font-sans text-sm text-muted"
                >{{ displayContent }}</pre>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 郵件內容隔離樣式 */
.email-content {
  /* 重置所有可能影響外部的樣式 */
  all: initial;
  display: block;
  font-family: inherit;
  line-height: 1.6;
  color: inherit;
  
  /* 確保隔離 */
  contain: style layout;
  isolation: isolate;
}

/* 重置郵件內容內的所有元素 */
.email-content :deep(*) {
  all: unset;
  display: revert;
  box-sizing: border-box;
}

.email-content :deep(body) {
  margin: 0;
  padding: 0;
  font-family: inherit;
}

.email-content :deep(html) {
  margin: 0;
  padding: 0;
}

/* 恢復必要的文字樣式 */
.email-content :deep(p),
.email-content :deep(div),
.email-content :deep(span) {
  display: block;
  margin: 0 0 1rem 0;
  color: inherit;
  font-family: inherit;
  font-size: inherit;
  line-height: inherit;
}

.email-content :deep(p:last-child),
.email-content :deep(div:last-child) {
  margin-bottom: 0;
}

/* 連結樣式 */
.email-content :deep(a) {
  color: rgb(37 99 235);
  text-decoration: none;
  display: inline;
}

.email-content :deep(a:hover) {
  text-decoration: underline;
}

/* 圖片樣式 */
.email-content :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 0.5rem;
  display: block;
  margin: 1rem 0;
}

/* 引用樣式 */
.email-content :deep(blockquote) {
  border-left: 4px solid rgb(229 231 235);
  padding-left: 1rem;
  margin: 1rem 0;
  font-style: italic;
  display: block;
}

/* 標題樣式 */
.email-content :deep(h1),
.email-content :deep(h2),
.email-content :deep(h3),
.email-content :deep(h4),
.email-content :deep(h5),
.email-content :deep(h6) {
  font-weight: bold;
  margin: 1.5rem 0 1rem 0;
  color: inherit;
  display: block;
}

.email-content :deep(h1) { font-size: 1.5rem; }
.email-content :deep(h2) { font-size: 1.25rem; }
.email-content :deep(h3) { font-size: 1.125rem; }

/* 列表樣式 */
.email-content :deep(ul),
.email-content :deep(ol) {
  margin: 1rem 0;
  padding-left: 2rem;
  display: block;
}

.email-content :deep(li) {
  display: list-item;
  margin: 0.5rem 0;
}

/* Dark mode */
:deep(.dark) .email-content :deep(a) {
  color: rgb(96 165 250);
}

:deep(.dark) .email-content :deep(blockquote) {
  border-color: rgb(31 41 55);
}
</style>
