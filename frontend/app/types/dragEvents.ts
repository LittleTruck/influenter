/**
 * 拖曳事件類型定義
 * 用於 vuedraggable 組件的事件處理
 */
export interface DragChangeEvent<T = unknown> {
  added?: {
    element: T
    newIndex: number
  }
  removed?: {
    element: T
    oldIndex: number
  }
  moved?: {
    element: T
    oldIndex: number
    newIndex: number
  }
}






