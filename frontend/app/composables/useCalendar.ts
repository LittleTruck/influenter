import type { Case, CaseDetail, CasePhase } from '~/types/cases'
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
   * 取得階段狀態顏色
   */
  const getPhaseStatusColor = (phase: CasePhase): string => {
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    
    const startDate = new Date(phase.start_date)
    startDate.setHours(0, 0, 0, 0)
    
    const endDate = new Date(phase.end_date)
    endDate.setHours(23, 59, 59, 999)

    if (today > endDate) {
      return '#10b981' // success/green - 已完成
    } else if (today >= startDate && today <= endDate) {
      return '#3b82f6' // primary/blue - 進行中
    } else {
      return '#6b7280' // neutral/gray - 待開始
    }
  }

  /**
   * 將案件轉換為日曆事件
   * 包含案件截止日期和階段事件
   */
  const casesToEvents = (casesList: Case[], view?: CalendarView): EventInput[] => {
    const events: EventInput[] = []

    casesList.forEach(caseItem => {
      const caseDetail = caseItem as CaseDetail
      
      // 如果有階段，顯示階段事件
      if (caseDetail.phases && caseDetail.phases.length > 0) {
        caseDetail.phases.forEach((phase: CasePhase) => {
          const startDate = new Date(phase.start_date)
          const endDate = new Date(phase.end_date)
          
          events.push({
            id: `phase-${phase.id}`,
            title: `${caseItem.title} - ${phase.name}`,
            start: startDate.toISOString(),
            end: new Date(endDate.getTime() + 86400000).toISOString(), // 加一天因為 end_date 是包含的
            allDay: true,
            backgroundColor: getPhaseStatusColor(phase),
            borderColor: getPhaseStatusColor(phase),
            extendedProps: {
              case: caseItem,
              phase: phase,
              type: 'phase'
            }
          } as EventInput)
        })
      } else if (caseItem.deadline_date) {
        // 如果沒有階段但有截止日期，顯示截止日期事件
        const deadlineDate = new Date(caseItem.deadline_date)
        
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
            case 'other':
              return '#6b7280' // neutral/gray
            default:
              return '#3b82f6'
          }
        }

        events.push({
          id: caseItem.id,
          title: `${caseItem.title} - ${caseItem.brand_name}`,
          start: deadlineDate.toISOString(),
          allDay: true,
          backgroundColor: getEventColor(caseItem.status),
          borderColor: getEventColor(caseItem.status),
          extendedProps: {
            case: caseItem,
            type: 'case'
          }
        } as EventInput)
      }
    })

    return events
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
    const eventId = dropInfo.event.id
    const newDate = dropInfo.event.start
    const eventType = dropInfo.event.extendedProps?.type

    // 如果是階段事件，不支援拖曳（階段日期應該在案件詳情頁面編輯）
    if (eventType === 'phase') {
      handleError(new Error('階段日期請在案件詳情頁面編輯'), '無法拖曳')
      return
    }

    // 處理案件截止日期拖曳
    const caseId = eventId
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

