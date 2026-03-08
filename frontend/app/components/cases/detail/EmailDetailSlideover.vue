<script setup lang="ts">
import { BaseSlideover, BaseButton, BaseIcon, BaseBadge } from '~/components/base'
import { format } from 'date-fns'

interface Props {
  modelValue: boolean
  emailId: string | null
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const emailsStore = useEmailsStore()
const loading = ref(false)

const email = computed(() => emailsStore.currentEmail)

// 載入郵件詳情（合併為單一 watcher，避免雙重觸發）
watch(
  [() => props.emailId, isOpen],
  async ([id, open]) => {
    if (id && open) {
      loading.value = true
      try {
        await emailsStore.fetchEmail(id)
      } finally {
        loading.value = false
      }
    }
  }
)

// 使用郵件清理 composable
const { sanitizeHtml } = useEmailSanitizer()

const displayContent = computed(() => {
  if (!email.value) return ''
  if (email.value.body_html) {
    return sanitizeHtml(email.value.body_html)
  }
  return email.value.body_text || ''
})

const isHtml = computed(() => !!email.value?.body_html)

const formatDate = (dateStr: string) => {
  return format(new Date(dateStr), 'yyyy/MM/dd HH:mm')
}
</script>

<template>
  <BaseSlideover
    v-model="isOpen"
    :title="email?.subject || '郵件詳情'"
    side="right"
    size="lg"
  >
    <template #body>
      <!-- 載入中 -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <BaseIcon name="i-lucide-loader-2" class="w-6 h-6 animate-spin text-primary" />
      </div>

      <!-- 郵件內容 -->
      <div v-else-if="email" class="space-y-4">
        <!-- 郵件 meta -->
        <div class="space-y-2 pb-4 border-b border-default">
          <div class="flex items-center gap-2">
            <BaseBadge
              :color="email.direction === 'outgoing' ? 'primary' : 'neutral'"
              variant="subtle"
              size="sm"
            >
              {{ email.direction === 'outgoing' ? '寄出' : '收件' }}
            </BaseBadge>
            <span class="text-xs text-dimmed">{{ formatDate(email.received_at) }}</span>
          </div>

          <div class="text-sm space-y-1">
            <div v-if="email.direction !== 'outgoing'">
              <span class="text-dimmed">寄件者：</span>
              <span class="text-highlighted font-medium">{{ email.from_name || email.from_email }}</span>
              <span v-if="email.from_name" class="text-muted ml-1">&lt;{{ email.from_email }}&gt;</span>
            </div>
            <div v-if="email.to_email">
              <span class="text-dimmed">收件者：</span>
              <span class="text-highlighted">{{ email.to_email }}</span>
            </div>
            <div v-if="email.subject">
              <span class="text-dimmed">主旨：</span>
              <span class="text-highlighted font-medium">{{ email.subject }}</span>
            </div>
          </div>
        </div>

        <!-- 郵件內文 -->
        <div class="email-body">
          <div
            v-if="isHtml"
            class="prose prose-sm dark:prose-invert max-w-none"
            v-html="displayContent"
          />
          <pre v-else class="text-sm text-highlighted whitespace-pre-wrap font-sans">{{ displayContent }}</pre>
        </div>
      </div>

      <!-- 空狀態 -->
      <div v-else class="text-center py-8 text-muted">
        <p>無法載入郵件內容</p>
      </div>
    </template>

    <template #footer>
      <div class="flex justify-end">
        <BaseButton
          color="neutral"
          variant="outline"
          @click="isOpen = false"
        >
          關閉
        </BaseButton>
      </div>
    </template>
  </BaseSlideover>
</template>

<style scoped>
.email-body :deep(img) {
  max-width: 100%;
  height: auto;
}
</style>
