<script setup lang="ts">
import FullCalendar from '@fullcalendar/vue3'
import dayGridPlugin from '@fullcalendar/daygrid'
import timeGridPlugin from '@fullcalendar/timegrid'
import interactionPlugin from '@fullcalendar/interaction'
import type { CalendarView } from '~/composables/useCalendar'
import type { EventDropArg } from '@fullcalendar/core'
import { format } from 'date-fns'
import { zhTW } from 'date-fns/locale'
import { BaseButton, BaseIcon } from '~/components/base'

const {
  currentView,
  currentDate,
  events,
  loading,
  setView,
  goToToday,
  prev,
  next,
  handleEventDrop
} = useCalendar()

// FullCalendar 選項
const calendarOptions = computed(() => ({
  plugins: [dayGridPlugin, timeGridPlugin, interactionPlugin],
  initialView: currentView.value,
  initialDate: currentDate.value,
  headerToolbar: false, // 我們自己實作工具列
  events: events.value,
  editable: true,
  droppable: false,
  eventStartEditable: true,
  eventDurationEditable: false, // 不允許調整事件長度
  eventDrop: (dropInfo: EventDropArg) => {
    handleEventDrop(dropInfo)
  },
  eventClick: (clickInfo: any) => {
    // 點擊事件時導航到案件詳情
    const caseId = clickInfo.event.extendedProps?.case?.id || clickInfo.event.id
    navigateTo(`/cases/${caseId}`)
  },
  height: 'auto', // 讓 FullCalendar 自動計算高度
  contentHeight: 'auto', // 自動填滿容器
  firstDay: 1, // 週一為第一天
  dayMaxEvents: 10, // 增加顯示的事件數量
  moreLinkClick: 'popover', // 點擊「更多」時顯示彈窗
  eventDisplay: 'block',
  eventTimeFormat: {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false
  },
  // 拖曳相關設定
  eventDragMinDistance: 5, // 最小拖曳距離
  dragScroll: true, // 啟用拖曳時自動滾動
  // 樣式相關
  eventClassNames: 'calendar-event',
  dayCellClassNames: 'calendar-day-cell',
  // 週視圖和日視圖設定 - 顯示完整時間軸（0:00-24:00），填滿空間
  allDaySlot: true, // 顯示全天事件區域
  allDayText: '全天',
  // 時間軸設定
  slotLabelFormat: {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false
  },
  // 確保週/日視圖顯示時間軸並填滿空間
  views: {
    timeGridWeek: {
      slotMinTime: '00:00:00',
      slotMaxTime: '24:00:00',
      slotDuration: '01:00:00', // 每個小時一個 slot
      slotLabelInterval: '01:00:00', // 每小時顯示一次標籤
      allDaySlot: true,
      allDayText: '全天',
      height: 'auto', // 自動填滿
      contentHeight: 'auto' // 自動計算內容高度
    },
    timeGridDay: {
      slotMinTime: '00:00:00',
      slotMaxTime: '24:00:00',
      slotDuration: '01:00:00', // 每個小時一個 slot
      slotLabelInterval: '01:00:00',
      allDaySlot: true,
      allDayText: '全天',
      height: 'auto',
      contentHeight: 'auto'
    },
    dayGridMonth: {
      fixedWeekCount: false, // 不固定週數，讓月視圖完整顯示所有週
      dayMaxEvents: 10,
      moreLinkClick: 'popover',
      height: 'auto',
      contentHeight: 'auto'
    }
  }
}))

// 視圖選項
const viewOptions: Array<{ value: CalendarView; label: string; icon: string }> = [
  { value: 'dayGridMonth', label: '月', icon: 'i-lucide-calendar' },
  { value: 'timeGridWeek', label: '週', icon: 'i-lucide-calendar-days' },
  { value: 'timeGridDay', label: '日', icon: 'i-lucide-calendar-days' }
]

// 視圖切換
const handleViewChange = (view: CalendarView) => {
  setView(view)
}

// 監聽視圖變化，更新日曆
watch(currentView, () => {
  // FullCalendar 會自動處理視圖切換
})

// 監聽日期變化
watch(currentDate, () => {
  // FullCalendar 會自動處理日期導航
})

// 日曆實例引用
const calendarRef = ref<InstanceType<typeof FullCalendar>>()

// 當視圖或日期變化時，更新日曆
watch([currentView, currentDate], () => {
  nextTick(() => {
    if (calendarRef.value?.getApi) {
      const calendarApi = calendarRef.value.getApi()
      calendarApi.changeView(currentView.value)
      calendarApi.gotoDate(currentDate.value)
    }
  })
})
</script>

<template>
  <div class="flex flex-col w-full min-h-[600px]">
    <!-- 工具列和日曆連在一起 -->
    <div class="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-800 shadow-sm overflow-visible w-full flex flex-col">
      <!-- 工具列 -->
      <div class="flex items-center justify-between p-3 border-b border-gray-200 dark:border-gray-800">
        <!-- 左側：導航按鈕 -->
        <div class="flex items-center gap-2">
          <BaseButton
            color="neutral"
            variant="ghost"
            size="sm"
            square
            icon="i-lucide-chevron-left"
            @click="prev"
          />
          <BaseButton
            color="neutral"
            variant="outline"
            size="sm"
            label="今天"
            @click="goToToday"
          />
          <BaseButton
            color="neutral"
            variant="ghost"
            size="sm"
            square
            icon="i-lucide-chevron-right"
            @click="next"
          />
          <div class="ml-3 text-base font-semibold text-gray-900 dark:text-white">
            {{ format(currentDate, 'yyyy年MM月', { locale: zhTW }) }}
          </div>
        </div>

        <!-- 右側：視圖切換 -->
        <div class="flex items-center gap-1">
          <BaseButton
            v-for="option in viewOptions"
            :key="option.value"
            :color="currentView === option.value ? 'primary' : 'neutral'"
            :variant="currentView === option.value ? 'solid' : 'ghost'"
            size="sm"
            :icon="option.icon"
            :label="option.label"
            @click="handleViewChange(option.value)"
          />
        </div>
      </div>

      <!-- 日曆主體 -->
      <div class="calendar-wrapper flex-1 min-h-0 w-full">
        <FullCalendar
          ref="calendarRef"
          :options="calendarOptions"
          class="calendar-container w-full h-full"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
/* FullCalendar 樣式客製化 - 使用 CSS 變數和直接屬性 */
.calendar-wrapper {
  flex: 1; /* 佔滿剩餘空間 */
  min-height: 0; /* 允許 flex 收縮 */
  overflow: visible; /* 避免裁切 */
  position: relative;
  display: flex;
  flex-direction: column;
  padding-bottom: 0 !important;
  margin-bottom: 0 !important;
  width: 100%; /* 撐滿寬度 */
}

:deep(.fc) {
  font-family: inherit;
  height: 100% !important;
  width: 100% !important; /* 撐滿寬度 */
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  padding-bottom: 0 !important;
  margin-bottom: 0 !important;
  overflow: visible !important; /* 避免裁切 */
  max-width: 100% !important; /* 確保不超出父元素 */
}

:deep(.fc-view-harness) {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

:deep(.fc-view-harness-active) {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  height: 100% !important;
}

:deep(.fc-header-toolbar) {
  display: none;
}

/* 移除日曆的邊框和圓角，因為已經在外層容器處理 */
:deep(.fc-scrollgrid) {
  border: none !important;
  border-radius: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding-bottom: 0 !important;
  margin-bottom: 0 !important;
  width: 100% !important; /* 撐滿寬度 */
  overflow: visible !important; /* 避免裁切 */
}

:deep(.fc-scrollgrid-section) {
  flex-shrink: 0;
}

:deep(.fc-scrollgrid-section-liquid) {
  flex: 1;
  min-height: 0;
  overflow-y: auto !important;
  overflow-x: hidden !important;
}

/* 移除月視圖底部的白色區域 */
:deep(.fc-daygrid-body) {
  padding-bottom: 0 !important;
  margin-bottom: 0 !important;
}

:deep(.fc-daygrid) {
  padding-bottom: 0 !important;
  margin-bottom: 0 !important;
}

:deep(.fc-scrollgrid-sync-table) {
  margin-bottom: 0 !important;
  padding-bottom: 0 !important;
}

/* 月視圖 - 讓格子更方，完整顯示 */
:deep(.fc-daygrid-day-frame) {
  min-height: 140px;
}

:deep(.fc-daygrid-day) {
  min-height: 140px;
}

/* 事件容器調整 - 案件置頂並撐滿寬度 */
:deep(.fc-daygrid-day-events) {
  /* margin-top: 1.75rem; */
  /* margin-bottom: 0.25rem; */
  padding: 0 !important; /* 移除所有 padding */
  width: 100%;
  box-sizing: border-box;
}

:deep(.fc-daygrid-day-event) {
  width: 100% !important;
  max-width: 100% !important;
  margin: 0 !important;
  margin-left: 0 !important;
  margin-right: 0 !important;
  box-sizing: border-box !important;
}

/* 移除事件 harness 的所有 margin - 使用更具體的選擇器提高優先級 */
:deep(.fc-daygrid-day-events .fc-daygrid-event-harness) {
  margin: 0 !important;
  margin-top: 0 !important;
  margin-bottom: 0 !important;
  margin-left: 0 !important;
  margin-right: 0 !important;
  width: 100% !important;
}

/* 使用屬性選擇器覆蓋內聯樣式 - 提高優先級 */
:deep(.fc-daygrid-day-events .fc-daygrid-event-harness[style]) {
  margin: 0 !important;
  margin-top: 0 !important;
  margin-bottom: 0 !important;
  margin-left: 0 !important;
  margin-right: 0 !important;
}

:deep(.fc-daygrid-event) {
  margin: 0 !important; /* 移除所有 margin */
  width: 100% !important; /* 撐滿寬度 */
  max-width: 100% !important;
  box-sizing: border-box !important;
  vertical-align: top !important;
}

/* 移除 day-bottom 的 margin - 使用更具體的選擇器 */
:deep(.fc-daygrid-day-events .fc-daygrid-day-bottom) {
  margin: 0 !important;
  margin-top: 0 !important;
  margin-bottom: 0 !important;
}

:deep(.fc-daygrid-day-events .fc-daygrid-day-bottom[style]) {
  margin: 0 !important;
  margin-top: 0 !important;
  margin-bottom: 0 !important;
}

/* 月視圖整體 - 確保完整顯示，可以滾動 */
:deep(.fc-daygrid-body) {
  min-height: auto;
  padding-bottom: 0 !important;
  margin-bottom: 0 !important;
}

:deep(.fc-daygrid) {
  height: auto !important;
  overflow-y: auto;
  padding-bottom: 0 !important;
  margin-bottom: 0 !important;
}

/* 確保月視圖可以完整顯示所有週 */
:deep(.fc-daygrid-body .fc-scroller) {
  overflow-y: auto !important;
  height: auto !important;
  padding-bottom: 0 !important;
  margin-bottom: 0 !important;
}

/* 移除月視圖底部的白色區域 */
:deep(.fc-scrollgrid-sync-table) {
  margin-bottom: 0 !important;
  padding-bottom: 0 !important;
}

:deep(.fc-scroller) {
  padding-bottom: 0 !important;
  margin-bottom: 0 !important;
}

:deep(.fc-scroller-liquid-absolute) {
  padding-bottom: 0 !important;
  margin-bottom: 0 !important;
}

/* 移除月視圖最後一行的底部邊距 */
:deep(.fc-daygrid-body tr:last-child td) {
  border-bottom: none !important;
  padding-bottom: 0 !important;
  margin-bottom: 0 !important;
}

/* 移除月視圖容器底部的空白 */
:deep(.fc-daygrid-container) {
  padding-bottom: 0 !important;
  margin-bottom: 0 !important;
}

/* 週視圖和日視圖 - 完整時間軸顯示（0:00-24:00），填滿空間 */
:deep(.fc-timegrid-body) {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

:deep(.fc-timegrid-body .fc-scroller) {
  flex: 1;
  min-height: 0;
  overflow-y: auto !important;
  overflow-x: hidden !important;
  height: 100% !important;
  max-height: none !important; /* 移除 max-height 限制 */
  -webkit-overflow-scrolling: touch; /* iOS 平滑滾動 */
  position: relative !important;
}

/* 確保時間網格內容可以滾動 */
:deep(.fc-timegrid-body .fc-scroller-liquid-absolute) {
  position: relative !important;
  height: auto !important;
  min-height: calc(24 * 4rem) !important; /* 24小時，每個slot 4rem */
  padding-bottom: 0 !important;
  margin-bottom: 0 !important;
}

/* 確保 scroller 可以正確滾動 */
:deep(.fc-timegrid-body .fc-scroller) {
  overflow-y: auto !important;
  overflow-x: hidden !important;
  height: 100% !important;
  max-height: none !important; /* 移除 max-height 限制 */
  position: relative !important;
}

/* 確保時間網格表格有足夠高度 */
:deep(.fc-timegrid-slot-table) {
  height: auto !important;
  min-height: calc(24 * 4rem) !important;
}

/* 確保時間網格容器可以滾動 */
:deep(.fc-timegrid-body) {
  overflow: hidden !important; /* 讓 scroller 處理滾動 */
  position: relative !important;
}

:deep(.fc-timegrid-body .fc-scroller-liquid) {
  height: 100% !important;
  overflow-y: auto !important;
  position: relative !important;
}

/* 確保 scroller harness 可以正確滾動 */
:deep(.fc-scroller-harness) {
  height: 100% !important;
  overflow: hidden !important;
}

:deep(.fc-scroller-harness-liquid) {
  height: 100% !important;
  overflow: hidden !important;
}

/* 時間 slot - 每個小時一個大格 */
:deep(.fc-timegrid-slot) {
  min-height: 4rem;
  height: auto;
}

:deep(.fc-timegrid-slot-label) {
  font-size: 0.875rem; /* text-sm */
  color: rgb(100 116 139);
  font-weight: 500;
  line-height: 1.5;
  font-family: inherit;
}

.dark :deep(.fc-timegrid-slot-label) {
  color: rgb(148 163 184);
}

/* 確保時間軸可見且完整 */
:deep(.fc-timegrid-axis) {
  width: 4rem !important;
  min-width: 4rem;
  flex-shrink: 0;
}

:deep(.fc-timegrid-axis-cushion) {
  font-size: 0.875rem; /* text-sm */
  font-weight: 500;
  line-height: 1.5;
  padding: 0.5rem;
  font-family: inherit;
}

/* 週視圖和日視圖的列 - 使用 flex 填滿空間 */
:deep(.fc-timegrid-col-frame) {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

:deep(.fc-timegrid-col) {
  flex: 1;
  min-height: 0;
}

/* 時間網格表格 - 確保有足夠高度顯示完整 24 小時（每個小時一個 slot） */
:deep(.fc-timegrid-slot-table) {
  height: auto !important;
  min-height: calc(24 * 4rem); /* 24小時，每個slot 4rem */
}

:deep(.fc-timegrid-slot-table tbody) {
  height: auto !important;
  min-height: calc(24 * 4rem);
}

/* 確保每個 slot 有最小高度（每個小時一個大格） */
:deep(.fc-timegrid-slot) {
  min-height: 4rem !important;
  height: 4rem !important;
}

/* 確保 slot 的 tr 也有正確高度 */
:deep(.fc-timegrid-slot-table tbody tr) {
  height: 4rem !important;
  min-height: 4rem !important;
}

/* 全天事件區域 */
:deep(.fc-all-day) {
  border-bottom: 2px solid rgb(229 231 235);
}

.dark :deep(.fc-all-day) {
  border-bottom-color: rgb(55 65 81);
}

/* 移除底部多餘的 padding */
:deep(.fc-scrollgrid-section-footer) {
  display: none;
}

/* 確保內容可以滾動 */
:deep(.fc-scrollgrid-sync-table) {
  height: auto !important;
}
</style>

