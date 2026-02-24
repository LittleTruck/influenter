<script setup lang="ts">
import { BaseButton, BaseIcon, BaseBadge, BaseAvatar, BaseDropdownMenu, BaseTextarea } from '~/components/base'
import AppSection from '~/components/ui/AppSection.vue'

definePageMeta({
  middleware: 'auth'
})

const config = useRuntimeConfig()
const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const emailsStore = useEmailsStore()
const toast = useToast()

const emailId = route.params.id as string

// 回覆框內容（使用者可審視、修改）
const replyBody = ref('')
const draftInstruction = ref('')
const draftLoading = ref(false)
const sendLoading = ref(false)
const replySectionRef = ref<HTMLElement | null>(null)

// 載入郵件詳情
onMounted(async () => {
  await emailsStore.fetchEmail(emailId)
  if (route.query.reply === '1') {
    await nextTick()
    replySectionRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    // 保留 from_case 參數，移除 reply
    const { reply, ...rest } = route.query
    router.replace({ path: route.path, query: rest })
  }
})

// 當前郵件
const email = computed(() => emailsStore.currentEmail)

// 標記已讀
const markAsRead = async (isRead: boolean) => {
  try {
    await emailsStore.markAsRead(emailId, isRead)
    toast.add({
      title: isRead ? '已標記為已讀' : '已標記為未讀',
      color: 'success'
    })
  } catch (e) {
    // error 已在 store 中處理
  }
}

// 返回列表（若從案件進來，返回該案件）
const fromCaseId = route.query.from_case as string | undefined
const goBack = () => {
  if (fromCaseId) {
    router.push(`/cases/${fromCaseId}`)
  } else {
    router.push('/emails')
  }
}

// 格式化日期
const formatDate = (dateStr: string) => {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-TW', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 使用郵件清理 composable
const { sanitizeHtml } = useEmailSanitizer()

// 顯示 HTML 或文字內容
const displayContent = computed(() => {
  if (!email.value) return ''
  const content = email.value.body_html || email.value.body_text || email.value.snippet || ''
  
  // 如果是 HTML 內容，進行清理
  if (email.value.body_html) {
    return sanitizeHtml(content)
  }
  
  return content
})

const contentType = computed(() => {
  if (!email.value) return 'text'
  return email.value.body_html ? 'html' : 'text'
})

// 關聯案件對話框
const showLinkDialog = ref(false)

const handleLinked = async (_caseId: string) => {
  // 重新載入郵件詳情
  await emailsStore.fetchEmail(emailId)
  toast.add({
    title: '案件關聯成功',
    color: 'success'
  })
}

// 加入為案件（AI 分析後建立案件）
const addingAsCase = ref(false)
const addAsCase = async () => {
  if (!email.value) return
  addingAsCase.value = true
  toast.add({ title: '正在建立案件，請稍候…', color: 'info' })
  try {
    const { caseId } = await emailsStore.createCaseFromEmail(emailId)
    await emailsStore.fetchEmail(emailId)
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

// AI 產生草稿：呼叫擬信 API，將草稿填入回覆框
const generateDraft = async () => {
  const em = email.value
  if (!em) return
  if (!em.case_id) {
    toast.add({
      title: '請先將此郵件關聯到案件',
      description: '關聯後即可使用 AI 產生回信草稿',
      color: 'warning'
    })
    return
  }

  draftLoading.value = true
  try {
    const res = await $fetch<{ draft: string }>(
      `${config.public.apiBase}/api/v1/cases/${em.case_id}/draft-reply`,
      {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ email_id: emailId, instruction: draftInstruction.value || undefined })
      }
    )
    replyBody.value = res.draft ?? ''
    toast.add({ title: '草稿已填入回覆框', color: 'success' })
  } catch (e: any) {
    const msg = e?.data?.message || e?.message || '產生草稿失敗'
    toast.add({ title: msg, color: 'error' })
  } finally {
    draftLoading.value = false
  }
}

const copyReplyBody = async () => {
  if (!replyBody.value) return
  try {
    await navigator.clipboard.writeText(replyBody.value)
    toast.add({ title: '已複製到剪貼簿', color: 'success' })
  } catch {
    toast.add({ title: '複製失敗', color: 'error' })
  }
}

// 寄出回信
const sendReply = async () => {
  const body = replyBody.value?.trim()
  if (!body) {
    toast.add({ title: '請輸入回信內容', color: 'warning' })
    return
  }

  sendLoading.value = true
  try {
    await $fetch(`${config.public.apiBase}/api/v1/emails/${emailId}/send-reply`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${authStore.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ body })
    })
    replyBody.value = ''
    toast.add({ title: '回信已寄出', color: 'success' })
    await emailsStore.fetchEmail(emailId)
    // 若有關聯案件，重新載入該案件郵件，案件詳情時間軸會顯示剛寄出的信
    const em = emailsStore.currentEmail
    if (em?.case_id) {
      const casesStore = useCasesStore()
      await casesStore.fetchCaseEmails(em.case_id).catch(() => {})
    }
  } catch (e: any) {
    const msg = e?.data?.message || e?.message || '寄出失敗'
    toast.add({ title: msg, color: 'error' })
  } finally {
    sendLoading.value = false
  }
}
</script>

<template>
  <div class="flex flex-col flex-1 h-full">
    <!-- Header -->
    <div class="flex items-center gap-4 px-6 py-4 border-b border-default">
      <BaseButton
        icon="i-lucide-arrow-left"
        color="neutral"
        variant="ghost"
        @click="goBack"
        aria-label="返回"
      />

      <h1 class="text-xl font-semibold text-highlighted flex-1">
        郵件詳情
      </h1>

      <div v-if="email" class="flex items-center gap-2">
        <!-- 已關聯：顯示案件；未關聯：加入為案件 -->
        <BaseButton
          v-if="email.case_id"
          icon="i-lucide-external-link"
          color="primary"
          size="sm"
          @click="router.push(`/cases/${email.case_id}`)"
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

        <!-- 標記已讀/未讀 -->
        <BaseButton
          :icon="email.is_read ? 'i-lucide-mail' : 'i-lucide-mail-open'"
          color="neutral"
          variant="outline"
          size="sm"
          @click="markAsRead(!email.is_read)"
        >
          {{ email.is_read ? '標記未讀' : '標記已讀' }}
        </BaseButton>

        <!-- 更多操作 -->
        <BaseDropdownMenu
          :items="[[
            {
              label: '關聯到案件',
              icon: 'i-lucide-link',
              onSelect: () => {
                showLinkDialog.value = true
              }
            },
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
            aria-label="更多操作"
          />
        </BaseDropdownMenu>
      </div>
    </div>

    <!-- Content -->
    <div v-if="emailsStore.loading" class="flex items-center justify-center h-full">
      <BaseIcon name="i-lucide-loader-2" class="w-8 h-8 animate-spin text-primary" />
    </div>

    <div v-else-if="email" class="flex-1 overflow-y-auto">
      <div class="max-w-4xl mx-auto p-6">
        <!-- Email Header Card -->
        <AppSection class="mb-6">
          <template #header>
            <h2 class="text-2xl font-bold text-highlighted">
              {{ email.subject || '(無主旨)' }}
            </h2>
          </template>

          <div class="space-y-4">
            <!-- 寄件者 -->
            <div class="flex items-start gap-3">
              <div class="flex-shrink-0">
                <BaseAvatar
                  :alt="email.from_name || email.from_email"
                  size="lg"
                />
              </div>
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <span class="font-semibold text-highlighted">
                    {{ email.from_name || email.from_email }}
                  </span>
                  <BaseBadge
                    v-if="!email.is_read"
                    size="xs"
                    color="primary"
                  >
                    未讀
                  </BaseBadge>
                </div>
                <div class="text-sm text-muted">{{ email.from_email }}</div>
                <div class="text-xs text-muted mt-1">
                  {{ formatDate(email.received_at) }}
                </div>
              </div>
            </div>

            <!-- 收件者 -->
            <div v-if="email.to_email" class="text-sm">
              <span class="text-muted">收件者：</span>
              <span class="text-highlighted">{{ email.to_email }}</span>
            </div>

            <!-- 標籤 -->
            <div v-if="email.labels && email.labels.length > 0" class="flex flex-wrap gap-2">
              <BaseBadge
                v-for="label in email.labels"
                :key="label"
                size="sm"
                variant="subtle"
                color="neutral"
              >
                {{ label }}
              </BaseBadge>
            </div>
          </div>
        </AppSection>

        <!-- Email Content -->
        <AppSection>
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="font-semibold text-highlighted">郵件內容</h3>
              <div class="flex items-center gap-2">
                <BaseBadge
                  v-if="email.has_attachments"
                  icon="i-lucide-paperclip"
                  variant="subtle"
                  color="neutral"
                >
                  有附件
                </BaseBadge>
                <BaseBadge
                  v-if="email.ai_analyzed"
                  icon="i-lucide-sparkles"
                  variant="subtle"
                  color="success"
                >
                  已AI分析
                </BaseBadge>
              </div>
            </div>
          </template>

          <div class="prose prose-sm max-w-none dark:prose-invert">
            <!-- HTML 內容 -->
            <div
              v-if="contentType === 'html'"
              v-html="displayContent"
              class="email-content isolate"
            />
            
            <!-- 純文字內容 -->
            <pre
              v-else
              class="whitespace-pre-wrap font-sans text-sm text-muted"
            >{{ displayContent }}</pre>
          </div>
        </AppSection>

        <!-- 回覆區：回覆框 + AI 產生草稿 -->
        <div ref="replySectionRef" class="mt-6">
        <AppSection>
          <template #header>
            <div class="flex items-center justify-between flex-wrap gap-2">
              <h3 class="font-semibold text-highlighted">回覆</h3>
              <div class="flex items-center gap-2">
                <BaseButton
                  icon="i-lucide-sparkles"
                  variant="outline"
                  size="sm"
                  :loading="draftLoading"
                  :disabled="draftLoading || sendLoading || !email?.case_id"
                  @click="generateDraft"
                >
                  AI 產生草稿
                </BaseButton>
                <BaseButton
                  v-if="replyBody"
                  icon="i-lucide-copy"
                  variant="outline"
                  size="sm"
                  :disabled="sendLoading"
                  @click="copyReplyBody"
                >
                  複製
                </BaseButton>
                <BaseButton
                  icon="i-lucide-send"
                  size="sm"
                  :loading="sendLoading"
                  :disabled="sendLoading || draftLoading || !replyBody?.trim()"
                  @click="sendReply"
                >
                  寄出
                </BaseButton>
              </div>
            </div>
          </template>

          <div class="space-y-3">
            <div>
              <label class="block text-sm font-medium text-muted mb-1">AI 補充說明（選填）</label>
              <BaseTextarea
                v-model="draftInstruction"
                :rows="2"
                placeholder="例如：拒絕這次合作，因為近期沒有合適的內容"
                :disabled="draftLoading || sendLoading"
              />
            </div>
            <BaseTextarea
              v-model="replyBody"
              :rows="12"
              placeholder="回信內容（可手動輸入或點「AI 產生草稿」填入）"
              class="font-mono text-sm"
            />
          </div>
        </AppSection>
        </div>
      </div>
    </div>

    <div v-else class="flex items-center justify-center h-full">
      <div class="text-center">
        <BaseIcon name="i-lucide-mail-x" class="w-16 h-16 mx-auto mb-4 text-muted" />
        <h3 class="text-lg font-semibold text-highlighted mb-2">找不到郵件</h3>
        <p class="text-muted mb-4">此郵件可能已被刪除</p>
        <BaseButton color="primary" @click="goBack">返回列表</BaseButton>
      </div>
    </div>

    <!-- Link to Case Dialog -->
    <EmailLinkToCaseDialog
      v-if="email"
      :email-id="emailId"
      v-model="showLinkDialog"
      @linked="handleLinked"
    />
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

