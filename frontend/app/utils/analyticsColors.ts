import { UNCATEGORIZED_SERIES_ID } from '~/types/analytics'

/**
 * 數據分析堆疊圖的固定調色盤（柔和色系），依系列索引循環取色。
 * 圖表與表格共用，確保同一合作項目顏色一致。
 */
const PALETTE = [
  '#C8915B', // 焦糖橘
  '#D9745B', // 陶土紅
  '#7FB77E', // 草綠
  '#5B8DEF', // 天藍
  '#E3C56B', // 芥末黃
  '#A78BC4', // 薰衣草紫
  '#5BC0BE', // 青綠
  '#E0828E', // 玫瑰粉
  '#8FA98C', // 灰綠
  '#C97FA8'  // 莓紫
]

/** 「未分類」固定使用中性灰 */
const UNCATEGORIZED_COLOR = '#9CA3AF'

/** 依系列在清單中的索引取得固定顏色；未分類固定為灰色 */
export function seriesColor(id: string, index: number): string {
  if (id === UNCATEGORIZED_SERIES_ID) return UNCATEGORIZED_COLOR
  return PALETTE[index % PALETTE.length] as string
}
