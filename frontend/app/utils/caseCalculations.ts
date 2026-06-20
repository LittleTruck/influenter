/**
 * 案件金額計算工具
 * 統一案件「總價」的計算邏輯，避免各處各自實作而不一致。
 */
import type { Case, CaseDetail, CaseCollaborationItem } from '~/types/cases'

/**
 * 取得單一合作項目在案件中的實際價格
 * 案件個別議價（cci.price）優先，否則使用合作項目預設價格。
 */
export const getCaseItemPrice = (cci: CaseCollaborationItem): number => {
  if (cci.price != null) return cci.price
  return cci.collaboration_item?.price ?? 0
}

/**
 * 合作項目價格加總（案件需包含 case_collaboration_items 才有值）
 */
export const getCollaborationItemsTotal = (c?: Case | CaseDetail | null): number => {
  const items = c?.case_collaboration_items ?? []
  return items.reduce((sum, cci) => sum + getCaseItemPrice(cci), 0)
}

/**
 * 取得案件顯示用總價（數值）。
 * 優先序：手動調整總價 adjusted_total ＞ 合作項目加總（若有）＞ 預估報價 quoted_amount。
 * 回傳 undefined 表示沒有可顯示的金額。
 */
export const getCaseDisplayTotal = (c?: Case | CaseDetail | null): number | undefined => {
  if (!c) return undefined
  if (c.adjusted_total != null) return c.adjusted_total
  const itemsTotal = getCollaborationItemsTotal(c)
  if (itemsTotal > 0) return itemsTotal
  return c.quoted_amount
}
