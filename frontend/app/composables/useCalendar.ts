import type { Case, CaseDetail, CasePhase, CaseStatus } from '~/types/cases'
import type { EventInput, EventDropArg } from '@fullcalendar/core'

/**
 * 日曆視圖類型
 */
export type CalendarView = 'dayGridMonth' | 'dayGridWeek' | 'dayGridDay'

/**
 * 日曆相關 composable
 * 提供日曆狀態管理和數據轉換功能
 */
export const useCalendar = () => {
  const { cases, fetchCases, fetchCase, updateCase } = useCases()
  const { handleError, handleSuccess } = useErrorHandler()

  // 當前視圖
  const currentView = ref<CalendarView>('dayGridMonth')

  // 當前日期
  const currentDate = ref<Date>(new Date())

  // 載入狀態
  const loading = ref(false)

  // 狀態篩選（null 代表「全部」）
  const selectedStatus = ref<CaseStatus | null>(null)

  // 含階段資料的案件（由 fetchAllCaseDetails 填充）
  const enrichedCases = ref<Case[]>([])

  /**
   * 載入所有案件及其階段資料
   * 先 fetchCases 取得列表，再逐一 fetchCase 取得 phases
   */
  const fetchAllCaseDetails = async () => {
    loading.value = true
    try {
      await fetchCases()
      // 先用列表資料讓日曆有東西顯示
      enrichedCases.value = [...cases.value]

      // 背景逐一取得每個案件的詳情（含 phases）
      const details = await Promise.all(
        cases.value.map(c =>
          fetchCase(c.id).catch(() => null)
        )
      )

      // 合併 phases 到 enrichedCases
      enrichedCases.value = cases.value.map((c, i) => {
        const detail = details[i] as CaseDetail | null
        if (detail?.phases && detail.phases.length > 0) {
          return { ...c, phases: detail.phases }
        }
        return c
      })
    } finally {
      loading.value = false
    }
  }

  /**
   * 將案件轉換為日曆事件
   * 包含案件截止日期和階段事件
   */
  const casesToEvents = (casesList: Case[], view?: CalendarView): EventInput[] => {
    const events: EventInput[] = []

    casesList.forEach(caseItem => {
      // 如果有階段，為每個階段的截止日（end_date）建立事件
      if (caseItem.phases && caseItem.phases.length > 0) {
        caseItem.phases.forEach((phase: CasePhase) => {
          if (!phase.end_date) return
          events.push({
            id: `phase-${phase.id}`,
            title: `${caseItem.title} - ${phase.name}`,
            start: phase.end_date,
            allDay: true,
            extendedProps: {
              case: caseItem,
              phase: phase,
              type: 'phase'
            }
          } as EventInput)
        })
      } else if (caseItem.deadline_date) {
        // 如果沒有階段但有截止日期，顯示截止日期事件
        events.push({
          id: caseItem.id,
          title: `${caseItem.title} - ${caseItem.brand_name}`,
          start: caseItem.deadline_date,
          allDay: true,
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
   * 計算事件（從含階段的案件列表，套用狀態篩選）
   */
  const events = computed(() => {
    const source = enrichedCases.value.length > 0 ? enrichedCases.value : cases.value
    if (!source) return []
    const filtered = selectedStatus.value
      ? source.filter(c => c.status === selectedStatus.value)
      : source
    return casesToEvents(filtered, currentView.value)
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
      handleSuccess('預計上線日已更新')
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
    } else if (currentView.value === 'dayGridWeek') {
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
    } else if (currentView.value === 'dayGridWeek') {
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
    selectedStatus,

    // 方法
    setView,
    goToDate,
    goToToday,
    prev,
    next,
    handleEventDrop,
    fetchAllCaseDetails,
    casesToEvents
  }
}
