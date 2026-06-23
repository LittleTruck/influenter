<script setup lang="ts">
import { BaseIcon } from '~/components/base'
import { formatAmount } from '~/utils/formatters'
import type { AnalyticsSeries, AnalyticsMonthPoint } from '~/types/analytics'

const props = defineProps<{
  series: AnalyticsSeries[]
  months: AnalyticsMonthPoint[]
  /** 系列 ID → 顏色（由父層提供，與圖表共用，確保顏色一致） */
  colorMap: Record<string, string>
}>()

// 顏色對照（由父層提供）
const colorOf = (id: string): string => props.colorMap[id] ?? '#9CA3AF'

// 月份由新到舊排列；只顯示有資料的月份
const rows = computed(() =>
  [...props.months]
    .filter((m) => m.total_amount > 0 || m.total_count > 0)
    .sort((a, b) => (a.month < b.month ? 1 : -1))
)

// 每月展開後的合作項目明細（依金額大到小）
const breakdownOf = (m: AnalyticsMonthPoint) =>
  props.series
    .map((s) => ({ id: s.id, title: s.title, amount: m.amounts[s.id] ?? 0, count: m.counts[s.id] ?? 0 }))
    .filter((b) => b.amount > 0 || b.count > 0)
    .sort((a, b) => b.amount - a.amount)

const grandTotalAmount = computed(() => props.months.reduce((sum, m) => sum + m.total_amount, 0))
const grandTotalCount = computed(() => props.months.reduce((sum, m) => sum + m.total_count, 0))

const monthLabel = (month: string): string => {
  const [yy, mm] = month.split('-')
  return `${yy}年${Number(mm)}月`
}
const fmtCount = (n: number): string => (n > 0 ? String(n) : '-')

// 展開狀態（預設全部收合）
const expanded = ref<Record<string, boolean>>({})
const toggle = (month: string) => {
  expanded.value[month] = !expanded.value[month]
}
</script>

<template>
  <div class="overflow-x-auto">
    <table class="w-full border-separate border-spacing-0 text-sm">
      <thead>
        <tr class="text-muted">
          <th class="border-b border-default px-3 py-2 text-left font-medium">月份</th>
          <th class="border-b border-default px-3 py-2 text-right font-medium">金額</th>
          <th class="border-b border-default px-3 py-2 text-right font-medium">專案數</th>
          <th class="border-b border-default px-3 py-2 w-10" />
        </tr>
      </thead>
      <tbody>
        <template v-for="m in rows" :key="m.month">
          <!-- 月份彙總列（可點擊展開） -->
          <tr
            class="cursor-pointer transition-colors hover:bg-elevated/50"
            @click="toggle(m.month)"
          >
            <td class="border-b border-default px-3 py-2 font-medium text-highlighted">
              <div class="flex items-center gap-2">
                <BaseIcon
                  name="i-lucide-chevron-right"
                  :class="`size-4 text-muted transition-transform ${expanded[m.month] ? 'rotate-90' : ''}`"
                />
                {{ monthLabel(m.month) }}
              </div>
            </td>
            <td class="border-b border-default px-3 py-2 text-right tabular-nums">
              {{ formatAmount(m.total_amount) }}
            </td>
            <td class="border-b border-default px-3 py-2 text-right tabular-nums">
              {{ fmtCount(m.total_count) }}
            </td>
            <td class="border-b border-default px-3 py-2" />
          </tr>

          <!-- 展開後的合作項目明細（用真正的儲存格對齊上方欄位） -->
          <template v-if="expanded[m.month]">
            <tr
              v-for="b in breakdownOf(m)"
              :key="`${m.month}-${b.id}`"
              class="bg-elevated/30"
            >
              <td class="border-b border-default py-1.5 pl-10 pr-3">
                <div class="flex items-center gap-2">
                  <span
                    class="inline-block size-2.5 shrink-0 rounded-full"
                    :style="{ backgroundColor: colorOf(b.id) }"
                  />
                  <span class="truncate text-default">{{ b.title }}</span>
                </div>
              </td>
              <td class="border-b border-default px-3 py-1.5 text-right tabular-nums text-muted">
                {{ formatAmount(b.amount) }}
              </td>
              <td class="border-b border-default px-3 py-1.5 text-right tabular-nums text-muted">
                {{ fmtCount(b.count) }}
              </td>
              <td class="border-b border-default" />
            </tr>
          </template>
        </template>
      </tbody>
      <tfoot v-if="rows.length">
        <tr class="font-semibold text-highlighted">
          <td class="px-3 py-2">合計</td>
          <td class="px-3 py-2 text-right tabular-nums">{{ formatAmount(grandTotalAmount) }}</td>
          <td class="px-3 py-2 text-right tabular-nums">{{ fmtCount(grandTotalCount) }}</td>
          <td class="px-3 py-2" />
        </tr>
      </tfoot>
    </table>
  </div>
</template>
