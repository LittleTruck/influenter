<script setup lang="ts">
import type { CaseEmail } from '~/types/cases'
import type { TimelineItem } from '@nuxt/ui'
import { BaseCard, BaseIcon } from '~/components/base'
import BaseCollapsible from '~/components/base/BaseCollapsible.vue'
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

// 轉換為 Timeline items
const timelineItems = computed<TimelineItem[]>(() => {
  return props.emails.map((email) => ({
    date: formatTime(email.received_at),
    title: email.from_name || email.from_email,
    description: email.subject || '',
    icon: getEmailIcon(email.email_type),
    value: email.id,
    // 將原始 email 數據存儲在自定義屬性中，以便在 slot 中使用
    _email: email
  }))
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
        <div class="flex items-center gap-2 min-w-0">
          <span class="font-medium text-highlighted truncate">{{ item.title }}</span>
          <span v-if="item.description" class="text-sm text-muted truncate">{{ item.description }}</span>
        </div>
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
                <p>
                  <span class="font-medium">寄件者：</span>{{ (item as any)._email.from_email }}
                </p>
                <p v-if="(item as any)._email.subject">
                  <span class="font-medium">主旨：</span>{{ (item as any)._email.subject }}
                </p>
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
