<template>
  <BaseDashboardPanel>
    <template #header>
      <BaseDashboardNavbar title="I'm working late, cause I'm an Influencer 🫠" />
    </template>

    <template #body>
        <!-- Welcome Section -->
        <div class="mb-6">
          <div class="flex items-center gap-3">
            <div>
              <h1 class="text-2xl font-bold text-highlighted">
                {{ user ? `${user.name || user.email}` : 'Influenter' }}
              </h1>
              <p class="text-sm text-muted">
                {{ todayString }}
              </p>
            </div>
          </div>
        </div>

        <!-- Stats Cards -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
          <NuxtLink to="/calendar?status=to_confirm" class="block">
            <BaseCard class="hover:ring-1 hover:ring-warning/50 transition-all cursor-pointer">
              <div class="flex items-center justify-between">
                <div>
                  <p class="text-sm text-muted">待確認</p>
                  <p class="text-2xl font-bold text-warning">{{ toConfirmCount }}</p>
                </div>
                <BaseIcon name="i-lucide-circle-alert" class="w-7 h-7 text-warning/60" />
              </div>
            </BaseCard>
          </NuxtLink>

          <NuxtLink to="/calendar?status=in_progress" class="block">
            <BaseCard class="hover:ring-1 hover:ring-primary/50 transition-all cursor-pointer">
              <div class="flex items-center justify-between">
                <div>
                  <p class="text-sm text-muted">進行中</p>
                  <p class="text-2xl font-bold text-primary">{{ inProgressCount }}</p>
                </div>
                <BaseIcon name="i-lucide-briefcase" class="w-7 h-7 text-primary/60" />
              </div>
            </BaseCard>
          </NuxtLink>

          <NuxtLink to="/emails" class="block">
            <BaseCard class="hover:ring-1 hover:ring-primary/50 transition-all cursor-pointer">
              <div class="flex items-center justify-between">
                <div>
                  <p class="text-sm text-muted">未讀郵件</p>
                  <p class="text-2xl font-bold text-highlighted">{{ unreadCount }}</p>
                </div>
                <BaseIcon name="i-lucide-mail" class="w-7 h-7 text-primary/60" />
              </div>
            </BaseCard>
          </NuxtLink>

          <BaseCard>
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-muted">本月預估收入</p>
                <p class="text-2xl font-bold text-highlighted">{{ formattedRevenue }}</p>
              </div>
              <BaseIcon name="i-lucide-dollar-sign" class="w-7 h-7 text-success/60" />
            </div>
          </BaseCard>
        </div>

        <!-- Main content: two columns -->
        <div class="grid grid-cols-1 lg:grid-cols-5 gap-6">
          <!-- Left: Action items -->
          <div class="lg:col-span-3 space-y-6">
            <!-- 待確認案件 -->
            <BaseCard v-if="toConfirmCases.length > 0">
              <template #header>
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <BaseIcon name="i-lucide-circle-alert" class="w-5 h-5 text-warning" />
                    <h3 class="font-semibold">待確認案件</h3>
                  </div>
                  <NuxtLink to="/calendar?status=to_confirm" class="text-sm text-primary hover:underline">
                    查看全部
                  </NuxtLink>
                </div>
              </template>
              <div class="divide-y divide-default">
                <NuxtLink
                  v-for="c in toConfirmCases.slice(0, 5)"
                  :key="c.id"
                  :to="`/cases/${c.id}`"
                  class="flex items-center justify-between py-3 first:pt-0 last:pb-0 hover:bg-elevated/50 -mx-2 px-2 rounded transition-colors"
                >
                  <div class="min-w-0 flex-1">
                    <p class="font-medium text-highlighted truncate">{{ c.alias || c.title || c.brand_name }}</p>
                    <p class="text-sm text-muted truncate">
                      {{ c.brand_name }}
                      <span v-if="getCaseDisplayTotal(c)"> · NT$ {{ getCaseDisplayTotal(c)!.toLocaleString() }}</span>
                    </p>
                  </div>
                  <div class="text-sm text-muted shrink-0 ml-3">
                    {{ formatRelativeDate(c.created_at) }}
                  </div>
                </NuxtLink>
              </div>
            </BaseCard>

            <!-- 近期更新案件 -->
            <BaseCard>
              <template #header>
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <BaseIcon name="i-lucide-activity" class="w-5 h-5 text-primary" />
                    <h3 class="font-semibold">近期案件</h3>
                  </div>
                  <NuxtLink to="/cases" class="text-sm text-primary hover:underline">
                    查看全部
                  </NuxtLink>
                </div>
              </template>
              <div v-if="recentCases.length > 0" class="divide-y divide-default">
                <NuxtLink
                  v-for="c in recentCases"
                  :key="c.id"
                  :to="`/cases/${c.id}`"
                  class="flex items-center justify-between py-3 first:pt-0 last:pb-0 hover:bg-elevated/50 -mx-2 px-2 rounded transition-colors"
                >
                  <div class="min-w-0 flex-1">
                    <div class="flex items-center gap-2">
                      <span
                        class="inline-block w-2 h-2 rounded-full shrink-0"
                        :style="{ backgroundColor: getStatusColorHex(c.status) }"
                      />
                      <p class="font-medium text-highlighted truncate">{{ c.alias || c.title || c.brand_name }}</p>
                    </div>
                    <p class="text-sm text-muted truncate ml-4">
                      {{ c.brand_name }}
                      <span v-if="c.deadline_date"> · 預計上線 {{ c.deadline_date }}</span>
                    </p>
                  </div>
                  <div class="text-xs text-muted shrink-0 ml-3">
                    {{ getStatusLabel(c.status) }}
                  </div>
                </NuxtLink>
              </div>
              <div v-else class="text-center py-8 text-muted">
                <BaseIcon name="i-lucide-inbox" class="w-10 h-10 mx-auto mb-2 opacity-40" />
                <p>還沒有案件</p>
                <NuxtLink to="/emails" class="text-sm text-primary hover:underline mt-1 inline-block">
                  從郵件開始建立案件
                </NuxtLink>
              </div>
            </BaseCard>
          </div>

          <!-- Right: Upcoming deadlines -->
          <div class="lg:col-span-2 space-y-6">
            <!-- 即將到期 -->
            <BaseCard>
              <template #header>
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <BaseIcon name="i-lucide-calendar-clock" class="w-5 h-5 text-primary" />
                    <h3 class="font-semibold">即將到期</h3>
                  </div>
                  <NuxtLink to="/calendar" class="text-sm text-primary hover:underline">
                    日曆
                  </NuxtLink>
                </div>
              </template>
              <div v-if="upcomingCases.length > 0" class="space-y-3">
                <NuxtLink
                  v-for="c in upcomingCases"
                  :key="c.id"
                  :to="`/cases/${c.id}`"
                  class="block p-3 rounded-lg border border-default hover:border-primary/30 transition-colors"
                >
                  <div class="flex items-start justify-between gap-2">
                    <div class="min-w-0 flex-1">
                      <p class="font-medium text-highlighted text-sm truncate">{{ c.alias || c.title || c.brand_name }}</p>
                      <p class="text-xs text-muted mt-0.5">{{ c.brand_name }}</p>
                    </div>
                    <span
                      class="text-xs font-medium px-2 py-0.5 rounded-full shrink-0"
                      :class="deadlineUrgencyClass(c.deadline_date!)"
                    >
                      {{ formatDeadline(c.deadline_date!) }}
                    </span>
                  </div>
                </NuxtLink>
              </div>
              <div v-else class="text-center py-6 text-muted">
                <BaseIcon name="i-lucide-calendar-check" class="w-8 h-8 mx-auto mb-2 opacity-40" />
                <p class="text-sm">近期沒有到期案件</p>
              </div>
            </BaseCard>

            <!-- 案件狀態總覽 -->
            <BaseCard v-if="casesStore.cases.length > 0">
              <template #header>
                <div class="flex items-center gap-2">
                  <BaseIcon name="i-lucide-pie-chart" class="w-5 h-5 text-primary" />
                  <h3 class="font-semibold">案件狀態</h3>
                </div>
              </template>
              <div class="space-y-3">
                <div
                  v-for="(status, key) in statusSummary"
                  :key="key"
                  class="flex items-center gap-3"
                >
                  <span
                    class="w-3 h-3 rounded-full shrink-0"
                    :style="{ backgroundColor: getStatusColorHex(key as CaseStatus) }"
                  />
                  <span class="text-sm flex-1">{{ getStatusLabel(key as CaseStatus) }}</span>
                  <span class="text-sm font-medium text-highlighted">{{ status.count }}</span>
                  <div class="w-20 h-1.5 bg-muted/20 rounded-full overflow-hidden">
                    <div
                      class="h-full rounded-full transition-all"
                      :style="{
                        width: `${status.percent}%`,
                        backgroundColor: getStatusColorHex(key as CaseStatus)
                      }"
                    />
                  </div>
                </div>
              </div>
            </BaseCard>
          </div>
        </div>
      </template>
  </BaseDashboardPanel>
</template>

<script setup lang="ts">
import { parseISO } from 'date-fns'
import { BaseDashboardPanel, BaseDashboardNavbar, BaseCard, BaseIcon } from '~/components/base'
import { getStatusLabel, getStatusColorHex } from '~/utils/caseStatus'
import { getCaseDisplayTotal } from '~/utils/caseCalculations'
import type { CaseStatus } from '~/types/cases'

definePageMeta({
  middleware: ['auth'],
})

const authStore = useAuthStore()
const casesStore = useCasesStore()
const emailsStore = useEmailsStore()

const user = computed(() => authStore.user)

// 載入資料
onMounted(async () => {
  await Promise.all([
    casesStore.fetchCases({ per_page: 50 }),
    emailsStore.fetchGmailStatus(),
  ])
})

// 今天日期
const todayString = computed(() => {
  const d = new Date()
  return d.toLocaleDateString('zh-TW', { year: 'numeric', month: 'long', day: 'numeric', weekday: 'long' })
})

// 統計數字
const toConfirmCount = computed(() => casesStore.casesByStatus.to_confirm.length)
const inProgressCount = computed(() => casesStore.casesByStatus.in_progress.length)
const unreadCount = computed(() => emailsStore.unreadCount)

const monthlyRevenue = computed(() => {
  const now = new Date()
  const thisMonth = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  return casesStore.cases
    .filter(c => (c.status === 'in_progress' || c.status === 'completed') && c.deadline_date?.startsWith(thisMonth))
    .reduce((sum, c) => sum + (getCaseDisplayTotal(c) || 0), 0)
})

const formattedRevenue = computed(() => {
  if (monthlyRevenue.value === 0) return 'NT$ 0'
  return `NT$ ${monthlyRevenue.value.toLocaleString()}`
})

// 待確認案件
const toConfirmCases = computed(() => casesStore.casesByStatus.to_confirm)

// 近期案件（排除待確認，按更新時間排序，取前 5 筆）
const recentCases = computed(() => {
  return [...casesStore.cases]
    .filter(c => c.status !== 'to_confirm')
    .sort((a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime())
    .slice(0, 5)
})

// 即將到期案件（有 deadline_date 且在未來 14 天內，不含已完成/取消）
const upcomingCases = computed(() => {
  const now = new Date()
  const twoWeeksLater = new Date(now.getTime() + 14 * 24 * 60 * 60 * 1000)

  return casesStore.cases
    .filter(c => {
      if (!c.deadline_date || c.status === 'completed' || c.status === 'cancelled') return false
      const deadline = new Date(c.deadline_date)
      return deadline >= now && deadline <= twoWeeksLater
    })
    .sort((a, b) => new Date(a.deadline_date!).getTime() - new Date(b.deadline_date!).getTime())
    .slice(0, 6)
})

// 案件狀態總覽
const statusSummary = computed(() => {
  const total = casesStore.cases.length
  const result: Record<string, { count: number; percent: number }> = {}
  for (const [key, cases] of Object.entries(casesStore.casesByStatus)) {
    if (cases.length > 0) {
      result[key] = {
        count: cases.length,
        percent: total > 0 ? Math.round((cases.length / total) * 100) : 0
      }
    }
  }
  return result
})

// 日期格式化
function formatRelativeDate(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

  if (diffDays === 0) return '今天'
  if (diffDays === 1) return '昨天'
  if (diffDays < 7) return `${diffDays} 天前`
  return date.toLocaleDateString('zh-TW', { month: 'short', day: 'numeric' })
}

// 期限倒數以「工作日」計（排除週末與國定假日）
const { workingDaysBetween } = useWorkdays()

function formatDeadline(dateStr: string): string {
  const days = workingDaysBetween(new Date(), parseISO(dateStr))
  if (days <= 0) return '今天'
  if (days <= 7) return `${days} 個工作日後`
  return parseISO(dateStr).toLocaleDateString('zh-TW', { month: 'short', day: 'numeric' })
}

function deadlineUrgencyClass(dateStr: string): string {
  const days = workingDaysBetween(new Date(), parseISO(dateStr))
  if (days <= 2) return 'bg-error/10 text-error'
  if (days <= 7) return 'bg-warning/10 text-warning'
  return 'bg-primary/10 text-primary'
}
</script>
