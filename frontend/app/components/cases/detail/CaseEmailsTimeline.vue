<script setup lang="ts">
import type { CaseEmail } from '~/types/cases'
import EmptyState from '~/components/common/EmptyState.vue'
import { format } from 'date-fns'

interface Props {
  /** 郵件列表 */
  emails: CaseEmail[]
}

const props = defineProps<Props>()

const expandedEmails = ref<string[]>([])

// 切換郵件展開狀態
const toggleEmail = (emailId: string) => {
  const index = expandedEmails.value.indexOf(emailId)
  if (index > -1) {
    expandedEmails.value.splice(index, 1)
  } else {
    expandedEmails.value.push(emailId)
  }
}

// 取得郵件圖示
const getEmailIcon = (emailType?: string) => {
  const icons: Record<string, string> = {
    initial_inquiry: 'i-lucide-mail',
    reply: 'i-lucide-reply',
    follow_up: 'i-lucide-mail-question'
  }
  return icons[emailType || ''] || 'i-lucide-mail'
}

// 格式化時間
const formatTime = (dateStr: string) => {
  return format(new Date(dateStr), 'yyyy/MM/dd HH:mm')
}
</script>

<template>
  <div class="email-timeline relative">
    <!-- 時間軸連接線 -->
    <div class="absolute left-4 top-0 bottom-0 w-0.5 bg-gray-200 dark:bg-gray-700"></div>

    <!-- 郵件節點 -->
    <div
      v-for="(email, index) in emails"
      :key="email.id"
      class="timeline-item relative pl-12 pb-6"
    >
      <!-- 節點圖示 -->
      <div
        class="absolute left-0 w-8 h-8 rounded-full bg-primary-500 border-4 border-white dark:border-gray-900 flex items-center justify-center z-10"
      >
        <UIcon :name="getEmailIcon(email.email_type)" class="w-4 h-4 text-white" />
      </div>

      <!-- 郵件卡片（可展開） -->
      <UCard
        class="email-card cursor-pointer transition-all duration-200 hover:shadow-md"
        @click="toggleEmail(email.id)"
      >
        <div class="flex items-start justify-between">
          <div class="flex-1 min-w-0">
            <div class="font-medium text-gray-900 dark:text-white">
              {{ email.from_name || email.from_email }}
            </div>
            <div v-if="email.subject" class="text-sm text-gray-500 dark:text-gray-400 mt-1">
              {{ email.subject }}
            </div>
          </div>
          <div class="text-xs text-gray-400 dark:text-gray-500 ml-4 flex-shrink-0">
            {{ formatTime(email.received_at) }}
          </div>
        </div>

        <!-- 展開內容（動畫） -->
        <Transition name="expand">
          <div v-if="expandedEmails.includes(email.id)" class="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
            <div class="text-sm text-gray-600 dark:text-gray-300">
              <p class="mb-2">
                <span class="font-medium">寄件者：</span>{{ email.from_email }}
              </p>
              <p v-if="email.subject" class="mb-2">
                <span class="font-medium">主旨：</span>{{ email.subject }}
              </p>
            </div>
          </div>
        </Transition>
      </UCard>
    </div>

    <!-- 空狀態 -->
    <EmptyState
      v-if="emails.length === 0"
      icon="i-lucide-mail"
      title="還沒有郵件"
      :show-icon-background="false"
    />
  </div>
</template>

<style scoped>
.expand-enter-active,
.expand-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}

.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  max-height: 0;
}

.expand-enter-to,
.expand-leave-from {
  opacity: 1;
  max-height: 1000px;
}
</style>

