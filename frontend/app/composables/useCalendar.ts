import type { Case } from '~/types/cases'
import type { EventInput, EventDropArg } from '@fullcalendar/core'

/**
 * 日曆視圖類型
 */
export type CalendarView = 'dayGridMonth' | 'timeGridWeek' | 'timeGridDay'

/**
 * 日曆相關 composable
 * 提供日曆狀態管理和數據轉換功能
 */
export const useCalendar = () => {
  const { cases, fetchCases, updateCase } = useCases()
  const { handleError, handleSuccess } = useErrorHandler()

  // 當前視圖
  const currentView = ref<CalendarView>('dayGridMonth')

  // 當前日期
  const currentDate = ref<Date>(new Date())

  // 載入狀態
  const loading = ref(false)

  /**
   * 將案件轉換為日曆事件
   */
  const casesToEvents = (casesList: Case[], view?: CalendarView): EventInput[] => {
    return casesList
      .filter(caseItem => caseItem.deadline_date)
      .map(caseItem => {
        const deadlineDate = new Date(caseItem.deadline_date!)
        
        // 根據案件狀態設定顏色
        const getEventColor = (status: Case['status']) => {
          switch (status) {
            case 'to_confirm':
              return '#f59e0b' // warning/amber
            case 'in_progress':
              return '#3b82f6' // primary/blue
            case 'completed':
              return '#10b981' // success/green
            case 'cancelled':
              return '#6b7280' // neutral/gray
            default:
              return '#3b82f6'
          }
        }

        // 所有案件都顯示為全天事件（因為目前案件還不能設定時間）
        const eventStart = deadlineDate.toISOString()
        
        return {
          id: caseItem.id,
          title: `${caseItem.title} - ${caseItem.brand_name}`,
          start: eventStart,
          allDay: true, // 所有案件都是全天事件
          backgroundColor: getEventColor(caseItem.status),
          borderColor: getEventColor(caseItem.status),
          extendedProps: {
            case: caseItem
          }
        } as EventInput
      })
  }

  /**
   * 計算事件（從案件列表）
   */
  const events = computed(() => {
    if (!cases.value) return []
    return casesToEvents(cases.value, currentView.value)
  })

  /**
   * 處理事件拖曳
   */
  const handleEventDrop = async (dropInfo: EventDropArg) => {
    const caseId = dropInfo.event.id
    const newDate = dropInfo.event.start

    // 找到對應的案件
    const caseItem = cases.value.find(c => c.id === caseId)
    if (!caseItem) {
      handleError(new Error('找不到對應的案件'), '更新失敗')
      return
    }

    // 格式化日期為 YYYY-MM-DD
    const formattedDate = newDate.toISOString().split('T')[0]

    loading.value = true
    try {
      await updateCase(caseId, {
        deadline_date: formattedDate
      })
      handleSuccess('案件截止日期已更新')
    } catch (error: any) {
      handleError(error, '更新案件日期失敗')
    } finally {
      loading.value = false
    }
  }

  /**
   * 切換視圖
   */
  const setView = (view: CalendarView) => {
    currentView.value = view
  }

  /**
   * 導航到指定日期
   */
  const goToDate = (date: Date) => {
    currentDate.value = date
  }

  /**
   * 導航到今天
   */
  const goToToday = () => {
    currentDate.value = new Date()
  }

  /**
   * 導航到上一個時間段
   */
  const prev = () => {
    const newDate = new Date(currentDate.value)
    if (currentView.value === 'dayGridMonth') {
      newDate.setMonth(newDate.getMonth() - 1)
    } else if (currentView.value === 'timeGridWeek') {
      newDate.setDate(newDate.getDate() - 7)
    } else {
      newDate.setDate(newDate.getDate() - 1)
    }
    currentDate.value = newDate
  }

  /**
   * 導航到下一個時間段
   */
  const next = () => {
    const newDate = new Date(currentDate.value)
    if (currentView.value === 'dayGridMonth') {
      newDate.setMonth(newDate.getMonth() + 1)
    } else if (currentView.value === 'timeGridWeek') {
      newDate.setDate(newDate.getDate() + 7)
    } else {
      newDate.setDate(newDate.getDate() + 1)
    }
    currentDate.value = newDate
  }

  return {
    // 狀態
    currentView,
    currentDate,
    events,
    loading,

    // 方法
    setView,
    goToDate,
    goToToday,
    prev,
    next,
    handleEventDrop,
    casesToEvents
  }
}

