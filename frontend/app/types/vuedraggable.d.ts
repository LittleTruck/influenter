declare module 'vuedraggable' {
  import { Component, ComponentPublicInstance } from 'vue'
  
  interface DraggableProps {
    modelValue?: any[]
    list?: any[]
    tag?: string | Component
    componentData?: Record<string, any>
    clone?: (element: any) => any
    move?: (evt: any) => boolean
    group?: string | { name: string; pull?: boolean | string | Function; put?: boolean | string | Function }
    sort?: boolean
    delay?: number
    delayOnTouchStart?: boolean
    touchStartThreshold?: number
    disabled?: boolean
    animation?: number
    handle?: string
    filter?: string
    preventOnFilter?: boolean
    draggable?: string
    ghostClass?: string
    chosenClass?: string
    dragClass?: string
    swapThreshold?: number
    invertSwap?: boolean
    invertedSwapThreshold?: number
    direction?: 'horizontal' | 'vertical'
    forceFallback?: boolean
    fallbackClass?: string
    fallbackOnBody?: boolean
    fallbackTolerance?: number
    fallbackOffset?: { x: number; y: number }
    supportPointer?: boolean
    emptyInsertThreshold?: number
    setData?: (dataTransfer: DataTransfer, dragEl: HTMLElement) => void
    [key: string]: any
  }

  interface DraggableEmits {
    (e: 'update:modelValue', value: any[]): void
    (e: 'end', event: any): void
    (e: 'start', event: any): void
    (e: 'add', event: any): void
    (e: 'remove', event: any): void
    (e: 'update', event: any): void
    (e: 'sort', event: any): void
    (e: 'filter', event: any): void
    (e: 'move', event: any, originalEvent: any): void | -1 | 1
    [key: string]: any
  }

  const Draggable: Component<DraggableProps, {}, any, any, any, any, any, DraggableEmits>
  export default Draggable
}

