/**
 * 數據分析相關型別（與後端 /analytics/collaboration-items 對齊）
 */

/** Y 軸指標：金額或專案數 */
export type AnalyticsMetric = 'amount' | 'count'

/** 堆疊系列（對應一個合作項目，或「未分類」） */
export interface AnalyticsSeries {
  id: string
  title: string
}

/** 單一月份的聚合資料 */
export interface AnalyticsMonthPoint {
  /** YYYY-MM */
  month: string
  /** seriesId -> 金額 */
  amounts: Record<string, number>
  /** seriesId -> 專案數 */
  counts: Record<string, number>
  /** 當月金額合計 */
  total_amount: number
  /** 當月專案數合計 */
  total_count: number
}

/** 合作項目數據分析回應 */
export interface CollaborationItemAnalyticsResponse {
  series: AnalyticsSeries[]
  months: AnalyticsMonthPoint[]
}

/** 未綁定任何合作項目的案件所歸入的虛擬系列 ID（需與後端一致） */
export const UNCATEGORIZED_SERIES_ID = 'uncategorized'
