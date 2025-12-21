<script setup lang="ts" generic="T extends Record<string, any>">
import { computed, onMounted, nextTick, ref } from 'vue'
import { useScroll } from '@vueuse/core'
import BaseButton from './BaseButton.vue'

/**
 * BaseTable - 表格元件的基礎封裝
 * 封裝 UTable，提供統一的表格介面
 */

defineOptions({ inheritAttrs: false })

interface Column {
  key: string
  label: string
  sortable?: boolean
  class?: string
}

interface TableUi {
  wrapper?: string
  base?: string
  divide?: string
  [key: string]: any
}

interface Props {
  /** 表格數據 (rows 模式) */
  rows?: T[]
  /** 表格數據 (data 模式，TanStack Table) */
  data?: any[]
  /** 表格列定義 */
  columns?: Column[] | any[]
  /** 是否載入中 */
  loading?: boolean
  /** 空狀態文字 */
  emptyState?: string
  /** 自定義 UI 配置 */
  ui?: TableUi
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  emptyState: '暫無資料',
})

const emit = defineEmits<{
  'select': [row: any]
  'edit': [row: any]
}>()

const handleEditClick = (row: any) => {
  emit('edit', row)
}

const handleSelect = (row: any) => {
  emit('select', row)
}

// 直接返回原始數據，不進行填充
const paddedData = computed(() => {
  return props.data
})

const paddedRows = computed(() => {
  return props.rows
})

// 檢查是否有 actions 欄位
const hasActionsColumn = computed(() => {
  if (!props.columns) return false
  return props.columns.some((col: any) => 
    col.id === 'actions' || col.key === 'actions' || col.accessorKey === 'actions'
  )
})

// 初始化 column-pinning
const columnPinning = computed(() => {
  if (hasActionsColumn.value) {
    return {
      left: [],
      right: ['actions']
    }
  }
  return { left: [], right: [] }
})

// 偵測是否滾動到最右邊（使用 VueUse）
const scrollContainer = ref<HTMLElement>()
const scrollTarget = ref<HTMLElement | null>(null)
const { arrivedState } = useScroll(scrollTarget)
const isScrolledToRight = computed(() => !!arrivedState.right)

onMounted(async () => {
  await nextTick()
  if (!scrollContainer.value) return
  // 尋找實際可水平滾動的子元素
  const nodes = Array.from(scrollContainer.value.querySelectorAll<HTMLElement>('*'))
  const candidate = nodes.find(el => (el.scrollWidth - el.clientWidth) > 1)
  // 若找不到，退回用容器本身
  scrollTarget.value = candidate ?? scrollContainer.value
})

// 規範化為 Nuxt UI v4 / TanStack Table 欄位格式
const normalizedColumns = computed<any>(() => {
  if (!props.columns) return []
  
  // 如果是已經格式化的 columns (data 模式)
  if (props.columns.length > 0 && ('accessorKey' in props.columns[0] || 'id' in props.columns[0])) {
    return props.columns
  }
  
  // 否則轉換為標準格式 (rows 模式)
  return props.columns?.map((col: any) => ({
    id: col.key,
    accessorKey: col.key,
    header: col.label,
    meta: { th: col.class || '', td: col.class || '' }
  })) as any
})
</script>

<template>
  <div class="base-table-wrapper">
    <div ref="scrollContainer" class="base-table-scroll-container" :class="{ 'scrolled-to-right': isScrolledToRight }">
      <UTable
        v-bind="$attrs"
        v-if="data"
        v-model:column-pinning="columnPinning"
        :data="paddedData"
        :columns="normalizedColumns"
        :loading="loading"
        :empty-state="{ icon: 'i-heroicons-circle-stack-20-solid', label: emptyState }"
        class="base-table"
        pinned
        :ui="props.ui || {
          base: 'table-fixed border-separate border-spacing-0 min-w-full',
          thead: '[&>tr]:bg-default [&>tr]:after:content-none sticky top-0 z-10',
          tbody: '',
          th: 'border-b border-default px-4 py-2 whitespace-nowrap min-w-[120px] [&[data-pinned]]:bg-default',
          td: 'border-b border-default px-4 py-2 whitespace-nowrap data-[pinned=right]:bg-default'
        }"
        @select="handleSelect"
      >
        <template v-for="(_, name) in $slots" #[name]="slotData">
          <slot :name="name" v-bind="slotData" />
        </template>
      </UTable>
      
      <UTable
        v-bind="$attrs"
        v-else
        :rows="paddedRows"
        :columns="normalizedColumns"
        :loading="loading"
        :empty-state="{ icon: 'i-heroicons-circle-stack-20-solid', label: emptyState }"
        class="base-table"
        pinned
        :ui="props.ui || {
          base: 'table-fixed border-separate border-spacing-0 min-w-full',
          thead: '[&>tr]:bg-default [&>tr]:after:content-none sticky top-0 z-10',
          tbody: '',
          th: 'border-b border-default px-4 py-2 whitespace-nowrap min-w-[120px] [&[data-pinned]]:bg-default',
          td: 'border-b border-default px-4 py-2 whitespace-nowrap [data-pinned]:bg-default'
        }"
      >
        <template v-for="(_, name) in $slots" #[name]="slotData">
          <slot :name="name" v-bind="slotData" />
        </template>
        <!-- 預設的 actions 欄位：若父層未自定義 actions-data slot，提供編輯按鈕 -->
        <template v-if="hasActionsColumn && !$slots['actions-data']" #actions-data="{ row }">
          <BaseButton
            color="primary"
            variant="ghost"
            icon="i-heroicons-pencil-square-20-solid"
            @click.stop="handleEditClick(row)"
          />
        </template>
      </UTable>
    </div>
  </div>
</template>

<style scoped>
.base-table-wrapper {
  width: 100%;
  height: 100%;
  position: relative;
  overflow: hidden;
}

.base-table-scroll-container {
  width: 100%;
  height: 100%;
  overflow: auto;
  /* 確保滾動條始終可見 */
  overflow-x: auto;
  overflow-y: auto;
  /* 滾動條固定在右下角 */
  position: relative;
}

.base-table {
  min-width: 100%;
  height: 100%;
}

/* 滾動條樣式 - 更細更透明 */
.base-table-scroll-container::-webkit-scrollbar {
  width: 4px;
  height: 4px;
}

.base-table-scroll-container::-webkit-scrollbar-button {
  display: none !important;
  width: 0 !important;
  height: 0 !important;
  background: transparent !important;
  border: none !important;
}

.base-table-scroll-container::-webkit-scrollbar-button:start:decrement,
.base-table-scroll-container::-webkit-scrollbar-button:end:increment,
.base-table-scroll-container::-webkit-scrollbar-button:single-button:vertical:decrement,
.base-table-scroll-container::-webkit-scrollbar-button:single-button:vertical:increment,
.base-table-scroll-container::-webkit-scrollbar-button:single-button:horizontal:decrement,
.base-table-scroll-container::-webkit-scrollbar-button:single-button:horizontal:increment,
.base-table-scroll-container::-webkit-scrollbar-button:double-button:vertical:start,
.base-table-scroll-container::-webkit-scrollbar-button:double-button:vertical:end,
.base-table-scroll-container::-webkit-scrollbar-button:double-button:horizontal:start,
.base-table-scroll-container::-webkit-scrollbar-button:double-button:horizontal:end {
  display: none !important;
  width: 0 !important;
  height: 0 !important;
  background: transparent !important;
  border: none !important;
}

.base-table-scroll-container::-webkit-scrollbar-track {
  background: transparent;
}

.base-table-scroll-container::-webkit-scrollbar-thumb {
  background-color: rgba(156, 163, 175, 0.2);
  border-radius: 2px;
}

.base-table-scroll-container::-webkit-scrollbar-thumb:hover {
  background-color: rgba(156, 163, 175, 0.4);
}


</style>

<style>
/* 讓陰影不被 table 的 overflow 裁切 */
.base-table table {
  overflow: visible !important;
}

/* pinned 陰影樣式暫時移除，先確保背景可被覆蓋為藍色 */

/* 非最右邊狀態：顯示右側陰影（Light：保持原設定） */
.base-table-scroll-container:not(.scrolled-to-right) [data-pinned=right],
.base-table-scroll-container:not(.scrolled-to-right) .test-shadow {
  box-shadow: 0px 0px 5px 0px rgba(0,0,0,0.1) !important;
}

/* 非最右邊狀態：顯示右側陰影（Dark：加強對比） */
.dark .base-table-scroll-container:not(.scrolled-to-right) [data-pinned=right],
.dark .base-table-scroll-container:not(.scrolled-to-right) .test-shadow {
  box-shadow: 0px 0px 5px 0px rgba(255,255,255,0.05) !important;
}

/* 滾動到最右邊時：取消右側陰影 */
.scrolled-to-right [data-pinned=right],
.scrolled-to-right .test-shadow {
  box-shadow: none !important;
}
</style>



