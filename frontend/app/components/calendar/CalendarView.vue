<script setup lang="ts">
import FullCalendar from '@fullcalendar/vue3'
import dayGridPlugin from '@fullcalendar/daygrid'
import interactionPlugin from '@fullcalendar/interaction'
import type { EventDropArg } from '@fullcalendar/core'
import { format } from 'date-fns'
import { STATUS_LABELS, STATUS_COLOR_HEX, STATUS_COLORS } from '~/utils/caseStatus'
import type { CaseStatus } from '~/types/cases'
import type { Holiday } from '~/stores/holidays'

// 狀態篩選 Tab 選項
const statusTabs: { label: string; value: CaseStatus | null; color: string }[] = [
  { label: '全部', value: null, color: '' },
  { label: STATUS_LABELS.to_confirm, value: 'to_confirm', color: STATUS_COLORS.to_confirm },
  { label: STATUS_LABELS.in_progress, value: 'in_progress', color: STATUS_COLORS.in_progress },
  { label: STATUS_LABELS.completed, value: 'completed', color: STATUS_COLORS.completed },
  { label: STATUS_LABELS.cancelled, value: 'cancelled', color: STATUS_COLORS.cancelled },
  { label: STATUS_LABELS.other, value: 'other', color: STATUS_COLORS.other }
]

const colorMode = useColorMode()
const isDark = computed(() => colorMode.value === 'dark')

const {
  currentDate,
  currentView,
  events,
  goToToday,
  prev,
  next,
  setView,
  handleEventDrop,
  selectedStatus
} = useCalendar()

// 國定假日資料（標註於日曆格；非事件，不佔用案件卡片空間、無需拖曳/點擊處理）
const holidaysStore = useHolidaysStore()
onMounted(() => holidaysStore.fetchHolidays())

// 日期鍵（yyyy-MM-dd）→ 假日，供 day cell hook 查詢
const holidayMap = computed(() => {
  const m = new Map<string, Holiday>()
  for (const h of holidaysStore.holidays) m.set(h.date, h)
  return m
})

// 是否有補班日（無則圖例不顯示補班）
const hasMakeup = computed(() => holidaysStore.holidays.some(h => h.is_workday))

// Header title
const headerTitle = computed(() => format(currentDate.value, 'MMMM yyyy'))

// 視圖切換選項（全部使用 dayGrid，因為案件都是全天事件）
type CalendarViewType = 'dayGridMonth' | 'dayGridWeek' | 'dayGridDay'
const viewOptions: { label: string; value: CalendarViewType }[] = [
  { label: '月', value: 'dayGridMonth' },
  { label: '週', value: 'dayGridWeek' },
  { label: '日', value: 'dayGridDay' }
]

const handleViewChange = (view: CalendarViewType) => {
  setView(view as any)
  nextTick(() => {
    if (calendarRef.value?.getApi) {
      const api = calendarRef.value.getApi()
      api.changeView(view)
    }
  })
}

// FullCalendar options
const calendarOptions = computed(() => ({
  plugins: [dayGridPlugin, interactionPlugin],
  initialView: currentView.value,
  initialDate: currentDate.value,
  headerToolbar: false as const,
  events: events.value,
  editable: true,
  droppable: false,
  eventStartEditable: true,
  eventDurationEditable: false,
  eventDrop: (dropInfo: EventDropArg) => {
    handleEventDrop(dropInfo)
  },
  eventClick: (clickInfo: any) => {
    const caseId = clickInfo.event.extendedProps?.case?.id || clickInfo.event.id
    navigateTo(`/cases/${caseId}`)
  },
  height: '100%',
  firstDay: 0,
  dayMaxEvents: false,
  fixedWeekCount: false,
  eventDisplay: 'block',
  eventDragMinDistance: 5,
  dragScroll: true,
  // 非工作日標註：整格底色 class。優先序 補班 > 國定假日 > 週末
  // （補班日雖落在週末仍是工作日，故排除於週末標註之外）
  dayCellClassNames: (arg: { date: Date }) => {
    const h = holidayMap.value.get(format(arg.date, 'yyyy-MM-dd'))
    if (h) return [h.is_workday ? 'inf-makeup' : 'inf-holiday']
    const dow = arg.date.getDay()
    return dow === 0 || dow === 6 ? ['inf-weekend'] : []
  },
  // 國定假日／補班日標註：在日期數字列左側放假日名稱
  dayCellContent: (arg: { date: Date; dayNumberText: string; isToday: boolean }) => {
    const h = holidayMap.value.get(format(arg.date, 'yyyy-MM-dd'))
    if (!h) return { html: String(arg.dayNumberText) }
    const label = h.is_workday ? `${h.name}·補班` : h.name
    const nameEl = document.createElement('span')
    nameEl.className = 'inf-holi-name'
    nameEl.textContent = label
    nameEl.title = label
    const nodes: HTMLElement[] = [nameEl]
    // 週/日視圖的 dayNumberText 為空（日期顯示於欄首），略過數字 span
    if (arg.dayNumberText) {
      const numEl = document.createElement('span')
      // 同時是今天時，把「今天」綠色圓圈套在數字上（而非整列）
      numEl.className = arg.isToday ? 'inf-holi-num inf-num-today' : 'inf-holi-num'
      numEl.textContent = String(arg.dayNumberText)
      nodes.push(numEl)
    }
    return { domNodes: nodes }
  },
  // Custom event rendering
  eventContent: (arg: any) => {
    const props = arg.event.extendedProps
    const caseItem = props.case
    const phase = props.phase
    const type = props.type

    const status = (caseItem?.status || 'other') as CaseStatus
    const statusLabel = STATUS_LABELS[status] || status
    const statusColor = STATUS_COLOR_HEX[status] || '#6b7280'
    const caseTitle = caseItem?.title || ''
    const brandName = caseItem?.brand_name || ''

    // Build DOM nodes (inline styles 確保所有視圖一致)
    const dark = isDark.value
    const cardBg = dark ? 'rgb(31 41 55)' : 'rgb(249 250 251)'
    const cardBgHover = dark ? 'rgb(55 65 81)' : 'rgb(243 244 246)'
    const cardBorder = dark ? 'rgb(55 65 81)' : 'rgb(229 231 235)'
    const titleColor = dark ? 'rgb(243 244 246)' : 'rgb(17 24 39)'
    const brandColor = dark ? 'rgb(107 114 128)' : 'rgb(156 163 175)'

    const container = document.createElement('div')
    Object.assign(container.style, {
      background: cardBg,
      border: `1px solid ${cardBorder}`,
      borderLeft: `3px solid ${statusColor}`,
      borderRadius: '6px',
      padding: '5px 8px',
      cursor: 'pointer',
      overflow: 'hidden',
      transition: 'all 0.15s ease'
    })
    container.onmouseenter = () => {
      container.style.background = cardBgHover
      container.style.boxShadow = '0 1px 3px rgba(0,0,0,0.06)'
    }
    container.onmouseleave = () => {
      container.style.background = cardBg
      container.style.boxShadow = 'none'
    }

    // Top line: status badge + brand/case name
    const topLine = document.createElement('div')
    Object.assign(topLine.style, {
      display: 'flex',
      alignItems: 'center',
      gap: '4px',
      lineHeight: '1'
    })

    const badge = document.createElement('span')
    Object.assign(badge.style, {
      fontSize: '0.625rem',
      fontWeight: '600',
      padding: '1px 4px',
      border: `1px solid ${statusColor}`,
      borderRadius: '3px',
      whiteSpace: 'nowrap',
      lineHeight: '1.4',
      color: statusColor
    })
    badge.textContent = statusLabel
    topLine.appendChild(badge)

    const topLabel = brandName || caseTitle
    if (topLabel) {
      const brand = document.createElement('span')
      Object.assign(brand.style, {
        fontSize: '0.68rem',
        color: brandColor,
        whiteSpace: 'nowrap',
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        lineHeight: '1.4'
      })
      brand.textContent = topLabel
      topLine.appendChild(brand)
    }

    container.appendChild(topLine)

    // Title line
    const titleLine = document.createElement('div')
    Object.assign(titleLine.style, {
      fontSize: '0.78rem',
      fontWeight: '600',
      color: titleColor,
      whiteSpace: 'nowrap',
      overflow: 'hidden',
      textOverflow: 'ellipsis',
      lineHeight: '1.5',
      marginTop: '1px'
    })
    if (type === 'phase') {
      titleLine.textContent = `${phase?.name || ''} DL`
    } else {
      titleLine.textContent = caseTitle || arg.event.title || ''
    }
    container.appendChild(titleLine)

    return { domNodes: [container] }
  },
  views: {
    dayGridMonth: {
      fixedWeekCount: false,
      dayMaxEvents: false,
      eventDisplay: 'block'
    },
    dayGridWeek: {
      eventDisplay: 'block'
    },
    dayGridDay: {
      eventDisplay: 'block'
    }
  }
}))

// Calendar instance ref
const calendarRef = ref<InstanceType<typeof FullCalendar>>()

// Sync date changes to FullCalendar
watch(currentDate, () => {
  nextTick(() => {
    if (calendarRef.value?.getApi) {
      const calendarApi = calendarRef.value.getApi()
      calendarApi.gotoDate(currentDate.value)
    }
  })
})

// 假日資料載入後強制重繪：day cell hook 在渲染當下才讀 holidayMap，
// 故 options 不會因假日載入而改變，需主動重繪一次套上標註。
watch(() => holidaysStore.loaded, (loaded) => {
  if (!loaded) return
  nextTick(() => calendarRef.value?.getApi?.().render())
})
</script>

<template>
  <div class="flex flex-col w-full h-full min-h-0">
    <!-- Header toolbar -->
    <div class="flex items-center justify-between px-1 pb-3 shrink-0">
      <div class="flex items-center gap-2">
        <h2 class="text-lg font-semibold text-highlighted">
          {{ headerTitle }}
        </h2>
        <div class="flex items-center gap-0.5">
          <button
            class="p-1.5 rounded-md hover:bg-muted text-dimmed transition-colors"
            @click="prev"
          >
            <UIcon name="i-lucide-chevron-left" class="w-4 h-4" />
          </button>
          <button
            class="px-3 py-1 text-sm font-medium text-muted hover:bg-muted rounded-md transition-colors"
            @click="goToToday"
          >
            今天
          </button>
          <button
            class="p-1.5 rounded-md hover:bg-muted text-dimmed transition-colors"
            @click="next"
          >
            <UIcon name="i-lucide-chevron-right" class="w-4 h-4" />
          </button>
        </div>
      </div>
      <!-- 視圖切換 -->
      <div class="flex items-center bg-muted rounded-lg p-0.5">
        <button
          v-for="opt in viewOptions"
          :key="opt.value"
          class="px-3 py-1 text-xs font-medium rounded-md transition-colors"
          :class="currentView === opt.value
            ? 'bg-elevated text-highlighted shadow-sm'
            : 'text-muted hover:text-highlighted'"
          @click="handleViewChange(opt.value)"
        >
          {{ opt.label }}
        </button>
      </div>
    </div>

    <!-- 狀態篩選 Tabs -->
    <div class="flex items-center gap-1 px-1 pb-3 shrink-0 flex-wrap">
      <button
        v-for="tab in statusTabs"
        :key="String(tab.value)"
        class="px-3 py-1 text-xs font-medium rounded-full border transition-colors"
        :class="selectedStatus === tab.value
          ? 'shadow-sm'
          : 'border-transparent opacity-60 hover:opacity-100'"
        :style="{
          color: selectedStatus === tab.value ? 'white' : (tab.value ? STATUS_COLOR_HEX[tab.value] : '#000000'),
          backgroundColor: selectedStatus === tab.value ? (tab.value ? STATUS_COLOR_HEX[tab.value] : '#000000') : 'transparent',
          borderColor: tab.value ? STATUS_COLOR_HEX[tab.value] : '#000000'
        }"
        @click="selectedStatus = tab.value"
      >
        {{ tab.label }}
      </button>

      <!-- 圖例：說明日曆格標註 -->
      <div class="ml-auto flex items-center gap-3 pl-2 text-xs text-dimmed">
        <span class="inline-flex items-center gap-1.5">
          <span class="inline-block w-2.5 h-2.5 rounded-sm holiday-legend-dot" />
          國定假日
        </span>
        <span class="inline-flex items-center gap-1.5">
          <span class="inline-block w-2.5 h-2.5 rounded-sm weekend-legend-dot" />
          週末
        </span>
        <span v-if="hasMakeup" class="inline-flex items-center gap-1.5">
          <span class="inline-block w-2.5 h-2.5 rounded-sm makeup-legend-dot" />
          補班日
        </span>
      </div>
    </div>

    <!-- Calendar grid -->
    <div class="calendar-wrapper flex-1 min-h-0">
      <FullCalendar
        ref="calendarRef"
        :options="calendarOptions"
      />
    </div>
  </div>
</template>

<style scoped>
/* ========== Calendar Wrapper ========== */
.calendar-wrapper {
  width: 100%;
  height: 100%;
  border: 1px solid rgb(229 231 235); /* gray-200 */
  border-radius: 8px;
  overflow: hidden;
}

/* ========== FullCalendar Base ========== */
:deep(.fc) {
  font-family: inherit;
  width: 100%;
  height: 100%;
}

:deep(.fc-header-toolbar) {
  display: none;
}

/* ========== Grid Structure — 移除 fc 自帶外框，由 wrapper 控制 ========== */
:deep(.fc-scrollgrid) {
  border: none !important;
}

/* ========== Column Headers ========== */
:deep(.fc-col-header-cell) {
  padding: 10px 0;
  font-size: 0.8rem;
  font-weight: 500;
  color: rgb(107 114 128);
  background: rgb(249 250 251); /* gray-50 */
  text-transform: capitalize;
}

:deep(.fc-col-header-cell-cushion) {
  text-decoration: none !important;
  color: inherit !important;
}

/* ========== Day Cells ========== */
:deep(.fc-daygrid-day) {
  min-height: auto;
  vertical-align: top;
}

:deep(.fc-daygrid-day-frame) {
  min-height: 130px;
  padding: 0;
}

/* Day number - right aligned */
:deep(.fc-daygrid-day-top) {
  display: flex;
  justify-content: flex-end;
  padding: 6px 8px 4px;
}

:deep(.fc-daygrid-day-number) {
  font-size: 0.85rem;
  color: rgb(107 114 128);
  text-decoration: none !important;
  padding: 0;
  line-height: 1;
}

/* Other month days - dimmed */
:deep(.fc-day-other .fc-daygrid-day-number) {
  color: rgb(209 213 219);
}

/* ========== Today Indicator (green primary) ========== */
:deep(.fc-day-today) {
  background: rgb(240 253 244) !important;
}

:deep(.fc-day-today .fc-daygrid-day-number) {
  background: rgb(22 163 74);
  color: white !important;
  border-radius: 50%;
  width: 26px;
  height: 26px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8rem;
  font-weight: 600;
}

/* ========== Events Container ========== */
:deep(.fc-daygrid-day-events) {
  padding: 2px 4px 4px !important;
  margin: 0 !important;
  display: flex !important;
  flex-direction: column !important;
  gap: 4px !important;
}

/* 日視圖加大 padding 與間距 */
:deep(.fc-dayGridDay-view .fc-daygrid-day-events) {
  padding: 8px 16px 12px !important;
  gap: 8px !important;
}

:deep(.fc-daygrid-event-harness) {
  margin: 0 !important;
}

:deep(.fc-daygrid-event-harness[style]) {
  margin-top: 0 !important;
}

:deep(.fc-daygrid-day-bottom) {
  margin: 0 !important;
  padding: 0 4px !important;
}

/* ========== Remove Default Event Styling（覆蓋全域 main.css）========== */
:deep(.fc-event),
:deep(.fc-h-event),
:deep(.fc-daygrid-event),
:deep(.fc-daygrid-block-event),
:deep(.fc-daygrid-dot-event),
:deep(.fc-daygrid-day-event),
:deep(.fc-event.fc-daygrid-event),
:deep(.fc-event.fc-daygrid-dot-event),
:deep(.fc-event.fc-h-event) {
  background: none !important;
  background-color: transparent !important;
  border: none !important;
  border-left: none !important;
  box-shadow: none !important;
  padding: 0 !important;
  margin: 0 !important;
  border-radius: 0 !important;
  cursor: pointer;
  display: block !important;
}

:deep(.fc-event:hover),
:deep(.fc-h-event:hover),
:deep(.fc-daygrid-dot-event:hover),
:deep(.fc-daygrid-event:hover) {
  background: none !important;
  background-color: transparent !important;
  box-shadow: none !important;
  transform: none !important;
}

:deep(.fc-event-main),
:deep(.fc-event-main-frame),
:deep(.fc-event-title-container) {
  padding: 0 !important;
  color: inherit !important;
}

:deep(.fc-event-title) {
  font-weight: inherit !important;
  padding: 0 !important;
}

:deep(.fc-event-selected),
:deep(.fc-event:focus) {
  box-shadow: none !important;
}

:deep(.fc-event-selected:after),
:deep(.fc-event:focus:after) {
  display: none !important;
}

:deep(.fc-daygrid-event-dot) {
  display: none !important;
}

:deep(.fc-event-time) {
  display: none !important;
}

/* ========== Custom Event Card（樣式由 JS inline style 控制，確保所有視圖一致） ========== */

/* ========== More Link ========== */
:deep(.fc-daygrid-more-link) {
  font-size: 0.75rem;
  color: rgb(22 163 74);
  font-weight: 500;
  padding: 2px 4px;
}

/* ========== Dark Mode ========== */
.dark .calendar-wrapper {
  border-color: rgb(55 65 81);
}

.dark :deep(.fc-col-header-cell) {
  color: rgb(107 114 128);
  background: rgb(31 41 55);
}

.dark :deep(.fc-scrollgrid td),
.dark :deep(.fc-scrollgrid th) {
  border-color: rgb(55 65 81) !important;
}

.dark :deep(.fc-daygrid-day-number) {
  color: rgb(156 163 175);
}

.dark :deep(.fc-day-other .fc-daygrid-day-number) {
  color: rgb(75 85 99);
}

.dark :deep(.fc-day-today) {
  background: rgba(20, 83, 45, 0.15) !important;
}

.dark :deep(.fc-day-today .fc-daygrid-day-number) {
  background: rgb(74 222 128);
  color: rgb(17 24 39) !important;
}

/* Dark mode event card 由 JS 處理 */

.dark :deep(.fc-daygrid-more-link) {
  color: rgb(74 222 128);
}

/* ========== 國定假日 / 補班日 標註 ========== */
/* 整格淡底色 */
:deep(.inf-holiday) {
  background: rgba(244, 63, 94, 0.06); /* rose */
}
:deep(.inf-makeup) {
  background: rgba(245, 158, 11, 0.08); /* amber */
}
:deep(.inf-weekend) {
  background: rgba(100, 116, 139, 0.07); /* slate */
}

/* 假日格的數字列改為左名稱、右日期 */
:deep(.inf-holiday .fc-daygrid-day-number),
:deep(.inf-makeup .fc-daygrid-day-number) {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  gap: 6px;
}

/* 假日名稱（左側，過長省略） */
:deep(.inf-holi-name) {
  flex: 1 1 auto;
  min-width: 0;
  text-align: left;
  font-size: 0.62rem;
  font-weight: 600;
  line-height: 1.4;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
:deep(.inf-holi-num) {
  flex-shrink: 0;
}

/* 假日（休假）配色 */
:deep(.inf-holiday .fc-daygrid-day-number),
:deep(.inf-holiday .inf-holi-name) {
  color: rgb(225 29 72); /* rose-600 */
}
/* 補班日配色 */
:deep(.inf-makeup .fc-daygrid-day-number),
:deep(.inf-makeup .inf-holi-name) {
  color: rgb(217 119 6); /* amber-600 */
}

/* 圖例色塊 */
.holiday-legend-dot {
  background: rgba(244, 63, 94, 0.55);
}
.weekend-legend-dot {
  background: rgba(100, 116, 139, 0.55);
}
.makeup-legend-dot {
  background: rgba(245, 158, 11, 0.6);
}

/* Dark mode */
.dark :deep(.inf-holiday) {
  background: rgba(244, 63, 94, 0.12);
}
.dark :deep(.inf-makeup) {
  background: rgba(245, 158, 11, 0.14);
}
.dark :deep(.inf-weekend) {
  background: rgba(148, 163, 184, 0.10);
}
.dark :deep(.inf-holiday .fc-daygrid-day-number),
.dark :deep(.inf-holiday .inf-holi-name) {
  color: rgb(251 113 133); /* rose-400 */
}
.dark :deep(.inf-makeup .fc-daygrid-day-number),
.dark :deep(.inf-makeup .inf-holi-name) {
  color: rgb(251 191 36); /* amber-400 */
}

/* 修正：週/日視圖中，非假日格被強制渲染的空白日期列。
   （定義 dayCellContent 會讓 FullCalendar 在所有視圖強制 day-top，
    但週/日視圖的日期本就顯示在欄首，故隱藏非假日格的空白列；
    假日格保留以顯示假日名稱。） */
:deep(.fc-dayGridWeek-view .fc-daygrid-day:not(.inf-holiday):not(.inf-makeup) .fc-daygrid-day-top),
:deep(.fc-dayGridDay-view .fc-daygrid-day:not(.inf-holiday):not(.inf-makeup) .fc-daygrid-day-top) {
  display: none;
}

/* 修正：今天又是假日時，把「今天」綠圈套在數字上，避免整列被撐成綠色長條 */
:deep(.fc-day-today.inf-holiday .fc-daygrid-day-number),
:deep(.fc-day-today.inf-makeup .fc-daygrid-day-number) {
  display: flex; /* 不依賴 source order，明確覆蓋今天的 inline-flex */
  background: transparent !important;
  width: 100%;
  height: auto;
  border-radius: 0;
  color: inherit !important;
}
:deep(.inf-num-today) {
  background: rgb(22 163 74);
  color: white !important;
  border-radius: 50%;
  width: 26px;
  height: 26px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8rem;
  font-weight: 600;
}
.dark :deep(.inf-num-today) {
  background: rgb(74 222 128);
  color: rgb(17 24 39) !important;
}

/* 修正：鄰月（other-month）假日格維持淡灰，不套用假日飽和色 */
:deep(.fc-day-other.inf-holiday .fc-daygrid-day-number),
:deep(.fc-day-other.inf-makeup .fc-daygrid-day-number),
:deep(.fc-day-other.inf-holiday .inf-holi-name),
:deep(.fc-day-other.inf-makeup .inf-holi-name) {
  color: rgb(209 213 219);
}
.dark :deep(.fc-day-other.inf-holiday .fc-daygrid-day-number),
.dark :deep(.fc-day-other.inf-makeup .fc-daygrid-day-number),
.dark :deep(.fc-day-other.inf-holiday .inf-holi-name),
.dark :deep(.fc-day-other.inf-makeup .inf-holi-name) {
  color: rgb(75 85 99);
}
</style>
