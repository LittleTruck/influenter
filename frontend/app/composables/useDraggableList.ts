import type { DragChangeEvent } from '~/types/dragEvents'

interface UseDraggableListOptions {
  /** 拖曳組名稱 */
  groupName: string
  /** 是否禁用拖曳 */
  disabled?: boolean
  /** 動畫時長（毫秒） */
  animation?: number
}

/**
 * 統一的 draggable 列表 composable
 * 以流程管理的實現為主
 */
export function useDraggableList<T extends { id: string }>(
  items: Ref<T[]>,
  options: UseDraggableListOptions
) {
  const {
    groupName,
    disabled = false,
    animation = 200
  } = options

  // 拖曳選項（統一使用流程管理的簡潔配置）
  const dragOptions = computed(() => ({
    animation,
    group: groupName,
    disabled,
    ghostClass: 'drag-ghost',
    chosenClass: 'drag-chosen',
    dragClass: 'drag-dragging',
    handle: '.drag-handle'
  }))

  // 處理拖曳變更
  const handleChange = (evt: DragChangeEvent<T>, onReorder?: (itemIds: string[]) => void) => {
    if (evt.moved || evt.added) {
      const itemIds = items.value.map(item => item.id)
      onReorder?.(itemIds)
    }
  }

  return {
    dragOptions,
    handleChange
  }
}

