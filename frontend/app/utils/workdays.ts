/**
 * 工作日計算工具（前端鏡像後端 internal/utils/workdays.go 的語意）。
 *
 * 統一規則：
 *  - 工期 N 個工作日採「含起始日」語意，結束日 = addWorkingDays(start, N - 1)。
 *  - start 會先正規化到工作日（ensureWorkingDay）。
 *  - 串聯下一階段起點 = nextWorkingDay(prevEnd)。
 *  - 「工作日」= 非週末且非國定假日；補班日例外仍算工作日。
 *  - calendar 模式退化為日曆天，未來可由帳號設定切換。
 *
 * 這些是純函式、不依賴任何 store；元件請透過 useWorkdays() 取得已綁定假日資料的版本。
 */
import { parseISO, format, addDays, isValid } from 'date-fns'

export type DayCalcMode = 'working' | 'calendar'

export interface DayCalcConfig {
  mode: DayCalcMode
  /** 視為休假的星期（0=週日 … 6=週六），預設 [0, 6] */
  weekends: number[]
  /** 國定假日（休假）日期集合，格式 "yyyy-MM-dd" */
  holidays: Set<string>
  /** 補班日（即使落在週末仍須上班）日期集合，格式 "yyyy-MM-dd" */
  makeupWorkdays: Set<string>
}

export const DEFAULT_WEEKENDS: number[] = [0, 6]

/** 全域預設設定：工作日模式、週末為六日、無假日資料（需由 useHolidays 載入）。 */
export function defaultDayCalcConfig(): DayCalcConfig {
  return {
    mode: 'working',
    weekends: [...DEFAULT_WEEKENDS],
    holidays: new Set<string>(),
    makeupWorkdays: new Set<string>()
  }
}

/** 正規化為當天零點，避免時分與時區干擾日期判斷。 */
function toDay(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate())
}

function toKey(d: Date): string {
  return format(d, 'yyyy-MM-dd')
}

export function isWorkingDay(date: Date, cfg: DayCalcConfig): boolean {
  if (cfg.mode === 'calendar') return true
  const d = toDay(date)
  const key = toKey(d)
  if (cfg.makeupWorkdays.has(key)) return true
  if (cfg.holidays.has(key)) return false
  return !cfg.weekends.includes(d.getDay())
}

/** 若 date 不是工作日則順延到下一個工作日；否則回傳正規化後的當天。 */
export function ensureWorkingDay(date: Date, cfg: DayCalcConfig): Date {
  let d = toDay(date)
  if (cfg.mode === 'calendar') return d
  while (!isWorkingDay(d, cfg)) d = addDays(d, 1)
  return d
}

/** 嚴格晚於 date 的第一個工作日。 */
export function nextWorkingDay(date: Date, cfg: DayCalcConfig): Date {
  let d = addDays(toDay(date), 1)
  if (cfg.mode === 'calendar') return d
  while (!isWorkingDay(d, cfg)) d = addDays(d, 1)
  return d
}

/**
 * 從 start 起算往後加 n 個工作日。
 * start 會先正規化到工作日；含起始日語意（工期 N 的結束日 = addWorkingDays(start, N - 1)）。
 * n 應為非負；calendar 模式退化為日曆天。
 */
export function addWorkingDays(start: Date, n: number, cfg: DayCalcConfig): Date {
  if (cfg.mode === 'calendar') return addDays(toDay(start), n)
  let d = ensureWorkingDay(start, cfg)
  let remaining = n
  while (remaining > 0) {
    d = addDays(d, 1)
    if (isWorkingDay(d, cfg)) remaining--
  }
  return d
}

/**
 * a 到 b 之間的工作日數（不含 a 當天、含 b 當天）；b 早於 a 回傳負值。
 * 用於「距離截止還有幾個工作日」的倒數。calendar 模式等同日曆天差。
 */
export function workingDaysBetween(a: Date, b: Date, cfg: DayCalcConfig): number {
  let da = toDay(a)
  let db = toDay(b)
  if (da.getTime() === db.getTime()) return 0
  let sign = 1
  if (db.getTime() < da.getTime()) {
    const tmp = da
    da = db
    db = tmp
    sign = -1
  }
  let count = 0
  let d = da
  while (d.getTime() < db.getTime()) {
    d = addDays(d, 1)
    if (isWorkingDay(d, cfg)) count++
  }
  return sign * count
}

/**
 * 由開始日字串與工期（工作日數）計算結束日字串。
 * 統一語意：含起始日，start 先正規化到工作日。回傳 "yyyy-MM-dd"；無效輸入回傳 ''。
 */
export function calculateEndDate(startDate: string, durationDays: number, cfg: DayCalcConfig): string {
  if (!startDate || durationDays < 1) return ''
  const parsed = parseISO(startDate)
  if (!isValid(parsed)) return ''
  const start = ensureWorkingDay(parsed, cfg)
  return format(addWorkingDays(start, durationDays - 1, cfg), 'yyyy-MM-dd')
}
