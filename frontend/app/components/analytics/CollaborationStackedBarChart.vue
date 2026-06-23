<script setup lang="ts">
import { VisXYContainer, VisStackedBar, VisAxis, VisBulletLegend, VisTooltip } from '@unovis/vue'
import { StackedBar } from '@unovis/ts'
import type { AnalyticsSeries, AnalyticsMonthPoint, AnalyticsMetric } from '~/types/analytics'

const props = defineProps<{
  series: AnalyticsSeries[]
  months: AnalyticsMonthPoint[]
  metric: AnalyticsMetric
  /** 系列 ID → 顏色（由父層提供，確保切換區間時顏色穩定） */
  colorMap: Record<string, string>
}>()

const colorOf = (id: string): string => props.colorMap[id] ?? '#9CA3AF'

type Row = AnalyticsMonthPoint & { _index: number }

const data = computed<Row[]>(() => props.months.map((m, i) => ({ ...m, _index: i })))

const valueOf = (m: AnalyticsMonthPoint, id: string): number =>
  props.metric === 'amount' ? (m.amounts[id] ?? 0) : (m.counts[id] ?? 0)

// X 軸用月份索引（VisStackedBar 的 x 需為數值）
const x = (d: Row) => d._index
// Y 傳一個 accessor 陣列 → 每個合作項目一個堆疊片段。
// 明確讀取 props.metric，使切換金額/專案數時 computed 重算並回傳「新的」函式陣列參考，
// 觸發 @unovis 的 prop 差異比較而重繪（否則切換 metric 圖表不會更新）。
const y = computed(() => {
  const metric = props.metric
  return props.series.map((s) => (d: Row) =>
    metric === 'amount' ? (d.amounts[s.id] ?? 0) : (d.counts[s.id] ?? 0)
  )
})
// 傳陣列 y 時，unovis 以「系列索引」呼叫 color
const color = (_d: Row, i: number) => colorOf(props.series[i]?.id ?? '')

// 當月最高堆疊總額（依目前指標），用來在 Y 軸頂端留白，避免最高的長條貼齊頂端被截斷
const maxTotal = computed(() => {
  const vals = props.months.map((m) => (props.metric === 'amount' ? m.total_amount : m.total_count))
  return vals.length ? Math.max(0, ...vals) : 0
})
const yDomain = computed<[number, number]>(() => [0, maxTotal.value > 0 ? maxTotal.value * 1.15 : 1])
// X 軸兩端各留半格，讓長條置中成「柱狀」而非貼齊邊緣（月份很少時尤其明顯）
const xDomain = computed<[number, number]>(() => [-0.5, Math.max(props.months.length - 0.5, 0.5)])

const monthLabel = (index: number): string => {
  const m = props.months[Math.round(index)]
  if (!m) return ''
  const [yy, mm] = m.month.split('-')
  return `${yy}/${mm}`
}

// 控制 X 軸標籤密度，月份過多時避免擠在一起
const labelStep = computed(() => Math.max(1, Math.ceil(props.months.length / 12)))
const xTickValues = computed(() => props.months.map((_, i) => i))
const xTickFormat = (index: number): string => {
  const idx = Math.round(index)
  return idx % labelStep.value === 0 ? monthLabel(idx) : ''
}

const fmtCurrency = (v: number): string =>
  new Intl.NumberFormat('zh-TW', {
    style: 'currency',
    currency: 'TWD',
    maximumFractionDigits: 0,
    notation: 'compact'
  }).format(v)
// 用 computed 包裝，使切換 metric 時回傳新的格式化函式參考，讓 Y 軸刻度一併重繪
const yTickFormat = computed(() => {
  const metric = props.metric
  return (v: number): string => (metric === 'amount' ? fmtCurrency(v) : String(Math.round(v)))
})

const legendItems = computed(() =>
  props.series.map((s) => ({ name: s.title, color: colorOf(s.id) }))
)

const triggers = {
  [StackedBar.selectors.bar]: (d: Row): string => {
    const lines = props.series
      .map((s, i) => ({ s, i, v: valueOf(d, s.id) }))
      .filter((r) => r.v > 0)
      .map((r) => {
        const val = props.metric === 'amount' ? fmtCurrency(r.v) : String(r.v)
        return `<div style="display:flex;align-items:center;gap:6px;white-space:nowrap">
          <span style="width:8px;height:8px;border-radius:9999px;background:${colorOf(r.s.id)};display:inline-block"></span>
          <span style="flex:1">${r.s.title}</span>
          <span style="margin-left:12px;font-variant-numeric:tabular-nums">${val}</span>
        </div>`
      })
      .join('')
    const total = props.metric === 'amount' ? fmtCurrency(d.total_amount) : String(d.total_count)
    return `<div style="padding:8px 10px;min-width:180px;font-size:12px">
      <div style="font-weight:600;margin-bottom:6px">${monthLabel(d._index)}</div>
      ${lines || '<div style="opacity:.6">無資料</div>'}
      <div style="border-top:1px solid rgba(0,0,0,.12);margin-top:6px;padding-top:6px;display:flex;justify-content:space-between;font-weight:600">
        <span>合計</span><span style="font-variant-numeric:tabular-nums">${total}</span>
      </div>
    </div>`
  }
}
</script>

<template>
  <div class="w-full analytics-chart">
    <VisBulletLegend :items="legendItems" class="mb-3 flex flex-wrap gap-x-4 gap-y-1 text-sm" />
    <VisXYContainer
      :data="data"
      :height="360"
      :margin="{ left: 16, right: 16, top: 12, bottom: 8 }"
      :x-domain="xDomain"
      :y-domain="yDomain"
    >
      <VisStackedBar :x="x" :y="y" :color="color" :rounded-corners="3" :bar-padding="0.2" :bar-max-width="56" />
      <VisAxis type="x" :tick-values="xTickValues" :tick-format="xTickFormat" :grid-line="false" />
      <VisAxis type="y" :tick-format="yTickFormat" :grid-line="true" :domain-line="false" />
      <VisTooltip :triggers="triggers" />
    </VisXYContainer>
  </div>
</template>

<style scoped>
.analytics-chart {
  --vis-axis-tick-label-color: var(--ui-text-muted);
  --vis-axis-grid-color: var(--ui-border);
  --vis-tooltip-background-color: var(--ui-bg);
  --vis-tooltip-text-color: var(--ui-text);
  --vis-tooltip-border-color: var(--ui-border-accented);
}
</style>
