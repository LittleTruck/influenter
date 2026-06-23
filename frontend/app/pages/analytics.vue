<script setup lang="ts">
import { BaseDashboardPanel, BaseDashboardNavbar, BaseCard, BaseButton, BaseIcon } from '~/components/base'
import { useAnalytics } from '~/composables/useAnalytics'
import { seriesColor } from '~/utils/analyticsColors'
import CollaborationStackedBarChart from '~/components/analytics/CollaborationStackedBarChart.vue'
import CollaborationBreakdownTable from '~/components/analytics/CollaborationBreakdownTable.vue'
import type { AnalyticsMetric, AnalyticsMonthPoint, AnalyticsSeries, CollaborationItemAnalyticsResponse } from '~/types/analytics'

definePageMeta({
  middleware: 'auth'
})

const { loading, error, fetchCollaborationAnalytics } = useAnalytics()

const metric = ref<AnalyticsMetric>('amount')
const analytics = ref<CollaborationItemAnalyticsResponse>({ series: [], months: [] })
const initialized = ref(false)

const metricOptions: { label: string; value: AnalyticsMetric; icon: string }[] = [
  { label: '金額', value: 'amount', icon: 'i-lucide-circle-dollar-sign' },
  { label: '專案數', value: 'count', icon: 'i-lucide-hash' }
]

// 時間區間
type RangeKey = 'last6' | 'last12' | 'thisYear' | 'all'
const range = ref<RangeKey>('last12')
const rangeOptions: { label: string; value: RangeKey }[] = [
  { label: '近 6 個月', value: 'last6' },
  { label: '近 12 個月', value: 'last12' },
  { label: '今年', value: 'thisYear' },
  { label: '全部', value: 'all' }
]

const pad2 = (n: number) => String(n).padStart(2, '0')
const monthKey = (y: number, m: number) => `${y}-${pad2(m)}`

// 選取區間對應的連續月份清單（含無資料的空月份）；'all' 回傳 null 代表沿用後端的資料月份
const windowMonths = computed<string[] | null>(() => {
  const now = new Date()
  const y = now.getFullYear()
  const m = now.getMonth() + 1 // 1-12
  const lastN = (count: number) => {
    const keys: string[] = []
    for (let i = count - 1; i >= 0; i--) {
      let ty = y
      let tm = m - i
      while (tm <= 0) { tm += 12; ty -= 1 }
      keys.push(monthKey(ty, tm))
    }
    return keys
  }
  if (range.value === 'last6') return lastN(6)
  if (range.value === 'last12') return lastN(12)
  if (range.value === 'thisYear') return Array.from({ length: 12 }, (_, i) => monthKey(y, i + 1))
  return null
})

const monthMap = computed<Record<string, AnalyticsMonthPoint>>(() => {
  const map: Record<string, AnalyticsMonthPoint> = {}
  for (const mo of analytics.value.months) map[mo.month] = mo
  return map
})

// 依選取區間產生要顯示的月份（空月份補 0，讓 X 軸為連續時間軸）
const displayMonths = computed<AnalyticsMonthPoint[]>(() => {
  const wm = windowMonths.value
  if (!wm) return analytics.value.months
  return wm.map((k) => monthMap.value[k] ?? { month: k, amounts: {}, counts: {}, total_amount: 0, total_count: 0 })
})

// 全集 series 的固定顏色對照（不受區間篩選影響，確保顏色穩定）
const colorMap = computed<Record<string, string>>(() => {
  const map: Record<string, string> = {}
  analytics.value.series.forEach((s, i) => { map[s.id] = seriesColor(s.id, i) })
  return map
})

// 區間內實際有資料的 series（圖例/堆疊/表格欄位用）
const displaySeries = computed<AnalyticsSeries[]>(() => {
  const present = new Set<string>()
  for (const mo of displayMonths.value) {
    for (const [id, v] of Object.entries(mo.amounts)) if (v) present.add(id)
    for (const [id, v] of Object.entries(mo.counts)) if (v) present.add(id)
  }
  return analytics.value.series.filter((s) => present.has(s.id))
})

const hasData = computed(() =>
  displaySeries.value.length > 0 && displayMonths.value.some((m) => m.total_amount > 0 || m.total_count > 0)
)

onMounted(async () => {
  analytics.value = await fetchCollaborationAnalytics()
  initialized.value = true
})
</script>

<template>
  <BaseDashboardPanel>
    <template #header>
      <BaseDashboardNavbar title="數據分析" />
    </template>

    <template #body>
      <!-- 圖表卡片（BaseDashboardPanel 的 body 本身已是可捲動的 flex 容器，
           卡片用 shrink-0 維持原本高度，避免展開明細時壓縮圖表） -->
      <BaseCard class="shrink-0">
          <template #header>
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h2 class="text-base font-semibold text-highlighted">每月合作{{ metric === 'amount' ? '金額' : '專案數' }}</h2>
                <p class="text-sm text-muted">依建立月份統計，堆疊顏色代表合作項目</p>
              </div>
              <div class="flex flex-wrap items-center gap-2">
                <!-- 時間區間 -->
                <div class="flex items-center gap-1 bg-muted p-1 rounded-lg shrink-0">
                  <BaseButton
                    v-for="opt in rangeOptions"
                    :key="opt.value"
                    :color="range === opt.value ? 'primary' : 'neutral'"
                    :variant="range === opt.value ? 'solid' : 'ghost'"
                    size="xs"
                    @click="range = opt.value"
                  >
                    {{ opt.label }}
                  </BaseButton>
                </div>
                <!-- Y 軸切換 -->
                <div class="flex items-center gap-1 bg-muted p-1 rounded-lg shrink-0">
                  <BaseButton
                    v-for="opt in metricOptions"
                    :key="opt.value"
                    :color="metric === opt.value ? 'primary' : 'neutral'"
                    :variant="metric === opt.value ? 'solid' : 'ghost'"
                    size="xs"
                    :icon="opt.icon"
                    @click="metric = opt.value"
                  >
                    {{ opt.label }}
                  </BaseButton>
                </div>
              </div>
            </div>
          </template>

          <!-- 載入中 -->
          <div v-if="loading || !initialized" class="flex h-[360px] items-center justify-center text-muted">
            <BaseIcon name="i-lucide-loader-circle" class="size-6 animate-spin" />
          </div>
          <!-- 錯誤 -->
          <div v-else-if="error" class="flex h-[360px] flex-col items-center justify-center gap-2 text-muted">
            <BaseIcon name="i-lucide-circle-alert" class="size-6 text-error" />
            <span>{{ error }}</span>
          </div>
          <!-- 空狀態 -->
          <div v-else-if="!hasData" class="flex h-[360px] flex-col items-center justify-center gap-2 text-muted">
            <BaseIcon name="i-lucide-chart-column" class="size-8" />
            <span>此區間尚無可分析的合作資料</span>
          </div>
          <!-- 圖表（僅在 client 端渲染） -->
          <ClientOnly v-else>
            <CollaborationStackedBarChart
              :key="`${range}-${metric}`"
              :series="displaySeries"
              :months="displayMonths"
              :metric="metric"
              :color-map="colorMap"
            />
            <template #fallback>
              <div class="flex h-[360px] items-center justify-center text-muted">
                <BaseIcon name="i-lucide-loader-circle" class="size-6 animate-spin" />
              </div>
            </template>
          </ClientOnly>
        </BaseCard>

      <!-- 明細表格卡片 -->
      <BaseCard v-if="hasData && !loading" class="shrink-0">
          <template #header>
            <h2 class="text-base font-semibold text-highlighted">每月明細</h2>
            <p class="text-sm text-muted">點擊月份可展開各合作項目的金額與專案數</p>
          </template>
          <CollaborationBreakdownTable
            :series="displaySeries"
            :months="displayMonths"
            :color-map="colorMap"
          />
      </BaseCard>
    </template>
  </BaseDashboardPanel>
</template>
