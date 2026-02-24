<script setup lang="ts">
import type { CaseEmail } from '~/types/cases'
import type { TimelineItem } from '@nuxt/ui'
import { BaseCard, BaseIcon } from '~/components/base'
import BaseCollapsible from '~/components/base/BaseCollapsible.vue'
import { format } from 'date-fns'

interface Props {
  /** 郵件列表 */
  emails: CaseEmail[]
  /** 案件 ID（用於返回導航） */
  caseId?: string
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
const getEmailIcon = (emailType?: string, direction?: string) => {
  if (direction === 'outgoing') {
    return 'i-lucide-send'
  }
  const icons: Record<string, string> = {
    sent: 'i-lucide-send',
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

// 轉換為 Timeline items
const timelineItems = computed<TimelineItem[]>(() => {
  return props.emails.map((email) => {
    const isOutgoing = email.direction === 'outgoing'
    const title = isOutgoing
      ? `寄給 ${email.to_email || '—'}`
      : (email.from_name || email.from_email)
    return {
      date: formatTime(email.received_at),
      title,
      description: email.subject || '',
      icon: getEmailIcon(email.email_type, email.direction),
      value: email.id,
      _email: email
    }
  })
})

// 計算當前活動的郵件（最後一封）
const activeEmailIndex = computed(() => {
  return props.emails.length > 0 ? props.emails.length - 1 : undefined
})
</script>

<template>
  <div class="case-emails-timeline">
    <UTimeline
      v-if="emails.length > 0"
      :items="timelineItems"
      :default-value="activeEmailIndex"
      color="primary"
    >
      <template #title="{ item }">
        <NuxtLink
          :to="caseId ? `/emails/${(item as any)._email.id}?from_case=${caseId}` : `/emails/${(item as any)._email.id}`"
          class="flex items-center gap-2 min-w-0 group"
        >
          <span class="font-medium text-highlighted truncate group-hover:text-primary group-hover:underline">{{ item.title }}</span>
          <span v-if="item.description" class="text-sm text-muted truncate">{{ item.description }}</span>
          <BaseIcon name="i-lucide-external-link" class="w-4 h-4 text-muted shrink-0 opacity-0 group-hover:opacity-100 transition-opacity" />
        </NuxtLink>
      </template>
      
      <template #date="{ item }">
        <div class="flex items-center justify-between gap-4 flex-shrink-0">
          <span class="text-xs text-dimmed whitespace-nowrap">{{ item.date }}</span>
          <BaseIcon
            v-if="expandedEmails.includes((item as any)._email.id)"
            name="i-lucide-chevron-down"
            class="w-4 h-4 text-gray-400 cursor-pointer"
            @click.stop="toggleEmail((item as any)._email.id)"
          />
          <BaseIcon
            v-else
            name="i-lucide-chevron-right"
            class="w-4 h-4 text-gray-400 cursor-pointer"
            @click.stop="toggleEmail((item as any)._email.id)"
          />
        </div>
      </template>

      <template #description="{ item }">
        <BaseCollapsible
          :open="expandedEmails.includes((item as any)._email.id)"
          @update:open="() => toggleEmail((item as any)._email.id)"
          :ui="{ content: 'pb-0' }"
        >
          <template #content>
            <BaseCard class="mt-2">
              <div class="text-sm text-gray-600 dark:text-gray-300 space-y-2">
                <p v-if="(item as any)._email.direction === 'outgoing'">
                  <span class="font-medium">寄給：</span>{{ (item as any)._email.to_email || '—' }}
                </p>
                <p v-else>
                  <span class="font-medium">寄件者：</span>{{ (item as any)._email.from_email }}
                </p>
                <p v-if="(item as any)._email.subject">
                  <span class="font-medium">主旨：</span>{{ (item as any)._email.subject }}
                </p>
                <NuxtLink
                  :to="caseId ? `/emails/${(item as any)._email.id}?from_case=${caseId}` : `/emails/${(item as any)._email.id}`"
                  class="inline-flex items-center gap-1.5 mt-2 text-primary hover:underline"
                >
                  <BaseIcon name="i-lucide-external-link" class="w-4 h-4" />
                  前往郵件詳情
                </NuxtLink>
              </div>
            </BaseCard>
          </template>
        </BaseCollapsible>
      </template>
    </UTimeline>

    <!-- 空狀態 -->
    <div v-else class="text-center py-8 text-muted">
      <p class="mb-1">還沒有郵件</p>
    </div>
  </div>
</template>

<style scoped>
.case-emails-timeline {
  padding: 1rem 0;
}
</style>
