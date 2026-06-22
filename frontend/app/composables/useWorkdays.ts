import { parseISO } from 'date-fns'
import {
  calculateEndDate as calcEndDate,
  workingDaysBetween as betweenWorkdays,
  ensureWorkingDay as ensureWorkday,
  nextWorkingDay as nextWorkday,
  isWorkingDay as isWorkday,
  type DayCalcConfig
} from '~/utils/workdays'

/**
 * 工作日計算 composable：把純函式（utils/workdays）與已載入的國定假日資料綁定。
 *
 * 首次使用時自動觸發假日載入（client 端）。所有回傳函式在呼叫當下讀取 store 的
 * reactive dayCalcConfig，因此放在元件 computed 內會在假日載入後自動重算。
 *
 * 用詞規則（消除「日」歧義）：
 *  - 工期 / 距離截止 → 「工作日」（本 composable 計算）。
 *  - 「多久以前」的過去時間戳 → 維持日曆天（formatters.ts 既有邏輯）。
 */
export const useWorkdays = () => {
  const store = useHolidaysStore()

  if (import.meta.client && !store.loaded && !store.loading) {
    store.fetchHolidays()
  }

  const cfg = computed<DayCalcConfig>(() => store.dayCalcConfig)

  return {
    cfg,
    loaded: computed(() => store.loaded),
    /** 由開始日字串 + 工期（工作日）計算結束日字串 "yyyy-MM-dd"。 */
    calculateEndDate: (startDate: string, durationDays: number) =>
      calcEndDate(startDate, durationDays, store.dayCalcConfig),
    /** a 到 b 之間的工作日數（不含 a、含 b）；用於期限倒數。 */
    workingDaysBetween: (a: Date, b: Date) => betweenWorkdays(a, b, store.dayCalcConfig),
    /** 若非工作日則順延到下一個工作日。 */
    ensureWorkingDay: (d: Date) => ensureWorkday(d, store.dayCalcConfig),
    /** 嚴格晚於 d 的下一個工作日（串聯下一階段起點）。 */
    nextWorkingDay: (d: Date) => nextWorkday(d, store.dayCalcConfig),
    /** d 是否為工作日。 */
    isWorkingDay: (d: Date) => isWorkday(d, store.dayCalcConfig),
    /** 截止日是否緊急：距今少於 3 個工作日且尚未過期。 */
    isDeadlineUrgent: (deadlineStr?: string): boolean => {
      if (!deadlineStr) return false
      const days = betweenWorkdays(new Date(), parseISO(deadlineStr), store.dayCalcConfig)
      return days >= 0 && days < 3
    }
  }
}
