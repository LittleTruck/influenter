<script setup lang="ts">
import type { CasePhase, FlowLayout } from '~/types/cases'
import draggable from 'vuedraggable'
import { BaseModal, BaseButton, BaseIcon } from '~/components/base'
import { format, addDays, parseISO, isBefore, isEqual } from 'date-fns'

interface PhaseRow {
  id: string
  name: string
  start_date: string
  duration_days: number
  end_date: string
  collaboration_item_id?: string
  /** 標記為待刪除 */
  _deleted?: boolean
  /** 標記為新增 */
  _new?: boolean
}

interface Props {
  modelValue: boolean
  phases: CasePhase[]
  caseId: string
  /** collaboration_item_id → 名稱 的對照表 */
  itemNameMap?: Record<string, string>
  /** 目前的流程排列模式 */
  flowLayout?: FlowLayout
}

const props = withDefaults(defineProps<Props>(), {})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'saved': []
}>()

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const toast = useToast()
const config = useRuntimeConfig()
const authStore = useAuthStore()
const apiHeaders = computed(() => ({
  Authorization: `Bearer ${authStore.token}`
}))

const saving = ref(false)

// 流程排列模式（本地編輯用）
const localFlowLayout = ref<FlowLayout>('parallel')

// 串聯模式下的日期衝突檢測
const sequentialConflicts = computed(() => {
  if (localFlowLayout.value !== 'sequential') return []
  const conflicts: Array<{ index: number; name: string; prevName: string }> = []
  const visibleRows = rows.value
  for (let i = 1; i < visibleRows.length; i++) {
    const prev = visibleRows[i - 1]!
    const curr = visibleRows[i]!
    if (!prev.end_date || !curr.start_date) continue
    const prevEnd = parseISO(prev.end_date)
    const currStart = parseISO(curr.start_date)
    // 衝突：下一個階段的開始日期 <= 上一個階段的結束日期
    if (isBefore(currStart, prevEnd) || isEqual(currStart, prevEnd)) {
      conflicts.push({ index: i, name: curr.name || `階段 ${i + 1}`, prevName: prev.name || `階段 ${i}` })
    }
  }
  return conflicts
})

// 合作項目顏色對照（依出現順序分配不同顏色）
const itemColorPalette = [
  { bg: 'bg-blue-500/10', text: 'text-blue-600 dark:text-blue-400' },
  { bg: 'bg-amber-500/10', text: 'text-amber-600 dark:text-amber-400' },
  { bg: 'bg-emerald-500/10', text: 'text-emerald-600 dark:text-emerald-400' },
  { bg: 'bg-purple-500/10', text: 'text-purple-600 dark:text-purple-400' },
  { bg: 'bg-rose-500/10', text: 'text-rose-600 dark:text-rose-400' },
  { bg: 'bg-cyan-500/10', text: 'text-cyan-600 dark:text-cyan-400' },
  { bg: 'bg-orange-500/10', text: 'text-orange-600 dark:text-orange-400' },
  { bg: 'bg-indigo-500/10', text: 'text-indigo-600 dark:text-indigo-400' },
]
const itemColorMap = computed(() => {
  const map: Record<string, (typeof itemColorPalette)[number]> = {}
  if (!props.itemNameMap) return map
  let idx = 0
  for (const id of Object.keys(props.itemNameMap)) {
    map[id] = itemColorPalette[idx % itemColorPalette.length]!
    idx++
  }
  return map
})

// 編輯用的本地副本（只存可見項目，已刪除的另外存）
const rows = ref<PhaseRow[]>([])
const deletedRows = ref<PhaseRow[]>([])

const calculateEndDate = (startDate: string, days: number): string => {
  if (!startDate || days < 1) return ''
  const start = parseISO(startDate)
  const end = addDays(start, days - 1)
  return format(end, 'yyyy-MM-dd')
}

// 打開時初始化
watch(isOpen, (open) => {
  if (open) {
    rows.value = props.phases
      .slice()
      .sort((a, b) => a.order - b.order)
      .map((p) => ({
        id: p.id,
        name: p.name,
        start_date: p.start_date,
        duration_days: p.duration_days,
        end_date: p.end_date,
        collaboration_item_id: p.collaboration_item_id
      }))
    deletedRows.value = []
    localFlowLayout.value = props.flowLayout || 'parallel'
  }
})

// 拖曳選項
const dragOptions = computed(() => ({
  animation: 200,
  ghostClass: 'drag-ghost',
  chosenClass: 'drag-chosen',
  dragClass: 'drag-dragging',
  handle: '.drag-handle'
}))

// 並聯模式：按合作項目分組
interface PhaseGroup {
  key: string
  itemId: string | null
  itemName: string
  color: (typeof itemColorPalette)[number] | null
  rows: PhaseRow[]
}

const groupedRows = computed((): PhaseGroup[] => {
  const groups: PhaseGroup[] = []
  const groupMap = new Map<string, PhaseRow[]>()

  for (const row of rows.value) {
    const key = row.collaboration_item_id || '__none__'
    if (!groupMap.has(key)) groupMap.set(key, [])
    groupMap.get(key)!.push(row)
  }

  for (const [key, groupRows] of groupMap) {
    const itemId = key === '__none__' ? null : key
    const itemName = itemId ? (props.itemNameMap?.[itemId] || '未知項目') : '未分類'
    const color = itemId ? (itemColorMap.value[itemId] || null) : null
    groups.push({ key, itemId, itemName, color, rows: groupRows })
  }

  return groups
})

// 並聯模式下，更新分組內的 rows 回寫到主 rows
const updateGroupRows = (groupKey: string, newGroupRows: PhaseRow[]) => {
  // 收集其他組的 rows（保持原順序）
  const otherRows = rows.value.filter(r => (r.collaboration_item_id || '__none__') !== groupKey)
  // 找到這個組在 rows 中第一次出現的位置，以此為插入點
  const firstIdx = rows.value.findIndex(r => (r.collaboration_item_id || '__none__') === groupKey)
  if (firstIdx === -1) {
    // 全部都是其他組，直接 append
    rows.value = [...otherRows, ...newGroupRows]
  } else {
    // 把 newGroupRows 插回原位
    const before = rows.value.slice(0, firstIdx).filter(r => (r.collaboration_item_id || '__none__') !== groupKey)
    const after = rows.value.slice(firstIdx).filter(r => (r.collaboration_item_id || '__none__') !== groupKey)
    rows.value = [...before, ...newGroupRows, ...after]
  }
}

// 更新結束日期
const updateEndDate = (row: PhaseRow) => {
  row.end_date = calculateEndDate(row.start_date, row.duration_days)
}

// 新增階段
const addRow = () => {
  const lastRow = rows.value[rows.value.length - 1]
  const startDate: string = lastRow?.end_date
    ? format(addDays(parseISO(lastRow.end_date), 1), 'yyyy-MM-dd')
    : new Date().toISOString().split('T')[0]!

  const newRow: PhaseRow = {
    id: `new_${Date.now()}`,
    name: '',
    start_date: startDate,
    duration_days: 7,
    end_date: calculateEndDate(startDate, 7),
    _new: true
  }
  rows.value.push(newRow)

  nextTick(() => {
    const inputs = document.querySelectorAll<HTMLInputElement>('[data-phase-name]')
    inputs[inputs.length - 1]?.focus()
  })
}

// 新增階段到指定分組
const addRowForGroup = (itemId: string | null) => {
  const groupKey = itemId || '__none__'
  const groupRows = rows.value.filter(r => (r.collaboration_item_id || '__none__') === groupKey)
  const lastRow = groupRows[groupRows.length - 1]
  const startDate: string = lastRow?.end_date
    ? format(addDays(parseISO(lastRow.end_date), 1), 'yyyy-MM-dd')
    : new Date().toISOString().split('T')[0]!

  const newRow: PhaseRow = {
    id: `new_${Date.now()}`,
    name: '',
    start_date: startDate,
    duration_days: 7,
    end_date: calculateEndDate(startDate, 7),
    collaboration_item_id: itemId || undefined,
    _new: true
  }
  rows.value.push(newRow)

  nextTick(() => {
    const inputs = document.querySelectorAll<HTMLInputElement>('[data-phase-name]')
    inputs[inputs.length - 1]?.focus()
  })
}

// 標記刪除
const markDelete = (row: PhaseRow) => {
  const idx = rows.value.indexOf(row)
  if (idx !== -1) {
    rows.value.splice(idx, 1)
    // 新增的直接丟棄，既有的放進已刪除清單
    if (!row._new) {
      row._deleted = true
      deletedRows.value.push(row)
    }
  }
}

// 復原刪除
const undoDelete = (row: PhaseRow) => {
  const idx = deletedRows.value.indexOf(row)
  if (idx !== -1) {
    deletedRows.value.splice(idx, 1)
    row._deleted = false
    rows.value.push(row)
  }
}

// 驗證
const validate = (): boolean => {
  for (const row of rows.value) {
    if (!row.name.trim()) {
      toast.add({ title: '請填寫所有階段名稱', color: 'error' })
      return false
    }
    if (!row.start_date) {
      toast.add({ title: `「${row.name}」缺少開始日期`, color: 'error' })
      return false
    }
    if (row.duration_days < 1) {
      toast.add({ title: `「${row.name}」天數至少為 1`, color: 'error' })
      return false
    }
  }
  return true
}

// 儲存所有變更
const handleSave = async () => {
  if (!validate()) return

  saving.value = true
  try {
    const baseUrl = `${config.public.apiBase}/api/v1/cases/${props.caseId}/phases`

    // 1. 刪除
    for (const row of deletedRows.value) {
      await $fetch(`${baseUrl}/${row.id}`, {
        method: 'DELETE',
        headers: apiHeaders.value
      })
    }

    // 2. 新增
    for (const [i, row] of rows.value.entries()) {
      if (!row._new) continue
      await $fetch(baseUrl, {
        method: 'POST',
        headers: apiHeaders.value,
        body: {
          name: row.name.trim(),
          start_date: row.start_date,
          duration_days: row.duration_days,
          collaboration_item_id: row.collaboration_item_id || undefined,
          order: i + 1
        }
      })
    }

    // 3. 更新既有階段
    for (const [i, row] of rows.value.entries()) {
      if (row._new) continue
      await $fetch(`${baseUrl}/${row.id}`, {
        method: 'PATCH',
        headers: apiHeaders.value,
        body: {
          name: row.name.trim(),
          start_date: row.start_date,
          end_date: row.end_date,
          duration_days: row.duration_days,
          order: i + 1
        }
      })
    }

    // 4. 更新流程排列模式（如有變更）
    if (localFlowLayout.value !== props.flowLayout) {
      await $fetch(
        `${config.public.apiBase}/api/v1/cases/${props.caseId}/flow-layout`,
        {
          method: 'PATCH',
          headers: apiHeaders.value,
          body: { flow_layout: localFlowLayout.value }
        }
      )
    }

    toast.add({ title: '流程已儲存', color: 'success' })
    isOpen.value = false
    emit('saved')
  } catch (error: any) {
    const msg = error?.data?.message || error?.response?._data?.message || '儲存失敗'
    toast.add({ title: msg, color: 'error' })
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <BaseModal
    v-model="isOpen"
    title="編輯專案流程"
    description="拖曳左側圖示調整順序，管理所有階段的名稱與日期"
    size="lg"
  >
    <template #body>
      <div class="space-y-3">
        <!-- 並聯/串聯切換 -->
        <div class="flex items-center justify-between px-3 py-2 rounded-lg bg-subtle border border-default">
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-highlighted">排列模式</span>
            <div class="flex items-center gap-1 bg-default rounded-lg p-0.5">
              <button
                class="text-xs px-2.5 py-1 rounded-md transition-colors"
                :class="localFlowLayout === 'parallel' ? 'bg-primary text-white shadow-sm' : 'text-muted hover:text-highlighted'"
                @click="localFlowLayout = 'parallel'"
              >
                並聯
              </button>
              <button
                class="text-xs px-2.5 py-1 rounded-md transition-colors"
                :class="localFlowLayout === 'sequential' ? 'bg-primary text-white shadow-sm' : 'text-muted hover:text-highlighted'"
                @click="localFlowLayout = 'sequential'"
              >
                串聯
              </button>
            </div>
          </div>
          <span class="text-xs text-dimmed">
            {{ localFlowLayout === 'parallel' ? '各項目同時進行' : '依序串接執行' }}
          </span>
        </div>

        <!-- 串聯日期衝突警告 -->
        <div
          v-if="sequentialConflicts.length > 0"
          class="flex items-start gap-2 px-3 py-2 rounded-lg bg-warning/10 border border-warning/30"
        >
          <UIcon name="i-lucide-alert-triangle" class="w-4 h-4 text-warning flex-shrink-0 mt-0.5" />
          <div class="text-xs text-warning">
            <p class="font-medium mb-0.5">串聯模式下有日期重疊：</p>
            <ul class="list-disc list-inside space-y-0.5">
              <li v-for="c in sequentialConflicts" :key="c.index">
                「{{ c.name }}」的開始日期與「{{ c.prevName }}」的結束日期重疊
              </li>
            </ul>
          </div>
        </div>

        <!-- 空狀態 -->
        <div v-if="rows.length === 0 && deletedRows.length === 0" class="text-center py-8 text-muted">
          <BaseIcon name="i-lucide-list-x" class="w-10 h-10 mx-auto mb-2 opacity-40" />
          <p class="text-sm">尚未設定流程階段</p>
          <p class="text-xs text-dimmed mt-1">點擊下方「新增階段」開始建立</p>
        </div>

        <!-- ====== 並聯模式：按合作項目分組 ====== -->
        <template v-if="localFlowLayout === 'parallel' && groupedRows.length > 0">
          <div
            v-for="group in groupedRows"
            :key="group.key"
            class="rounded-lg border border-default overflow-hidden"
          >
            <!-- 分組標頭 -->
            <div class="flex items-center justify-between px-3 py-2 bg-subtle border-b border-default">
              <div class="flex items-center gap-2">
                <span
                  v-if="group.color"
                  class="text-[11px] font-medium px-1.5 py-0.5 rounded truncate max-w-[120px]"
                  :class="[group.color.bg, group.color.text]"
                >
                  {{ group.itemName }}
                </span>
                <span
                  v-else
                  class="text-[11px] font-medium px-1.5 py-0.5 rounded bg-gray-500/10 text-gray-500 dark:text-gray-400"
                >
                  未分類
                </span>
                <span class="text-xs text-dimmed">{{ group.rows.length }} 個階段</span>
              </div>
              <button
                class="flex items-center gap-1 text-xs text-muted hover:text-primary transition-colors"
                @click="addRowForGroup(group.itemId)"
              >
                <UIcon name="i-lucide-plus" class="w-3.5 h-3.5" />
                新增
              </button>
            </div>

            <!-- 分組內可拖曳列表 -->
            <draggable
              :model-value="group.rows"
              @update:model-value="updateGroupRows(group.key, $event)"
              v-bind="dragOptions"
              item-key="id"
              class="divide-y divide-default"
            >
              <template #item="{ element: row, index }">
                <div class="flex items-center gap-2 px-3 py-2.5 bg-default transition-shadow">
                  <div class="drag-handle flex items-center self-stretch px-0.5">
                    <UIcon name="i-lucide-grip-vertical" class="w-4 h-4 text-dimmed" />
                  </div>
                  <span class="text-xs font-semibold text-muted w-5 text-center flex-shrink-0">{{ index + 1 }}</span>
                  <div class="flex-1 grid grid-cols-[1fr_auto_auto_auto] gap-2 items-center">
                    <input
                      v-model="row.name"
                      data-phase-name
                      placeholder="階段名稱"
                      class="w-full rounded-md border border-default bg-default px-2.5 py-1.5 text-sm text-highlighted shadow-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
                    />
                    <input
                      v-model="row.start_date"
                      type="date"
                      class="w-[140px] rounded-md border border-default bg-default px-2 py-1.5 text-sm text-highlighted shadow-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
                      @change="updateEndDate(row)"
                    />
                    <div class="flex items-center gap-1">
                      <input
                        v-model.number="row.duration_days"
                        type="number"
                        min="1"
                        max="365"
                        class="w-[64px] rounded-md border border-default bg-default px-2 py-1.5 text-sm text-highlighted text-center shadow-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
                        @input="updateEndDate(row)"
                      />
                      <span class="text-xs text-dimmed">天</span>
                    </div>
                    <span class="text-xs text-dimmed whitespace-nowrap min-w-[80px] text-right">
                      → {{ row.end_date ? format(parseISO(row.end_date), 'MM/dd') : '-' }}
                    </span>
                  </div>
                  <button
                    class="text-dimmed hover:text-red-500 transition-colors flex-shrink-0"
                    @click="markDelete(row)"
                  >
                    <UIcon name="i-lucide-x" class="w-4 h-4" />
                  </button>
                </div>
              </template>
            </draggable>
          </div>
        </template>

        <!-- ====== 串聯模式：單一列表 ====== -->
        <template v-else-if="rows.length > 0">
          <draggable
            v-model="rows"
            v-bind="dragOptions"
            item-key="id"
            class="space-y-2"
          >
            <template #item="{ element: row, index }">
              <div class="flex items-center gap-2 p-3 rounded-lg border border-default bg-subtle transition-shadow">
                <div class="drag-handle flex items-center self-stretch px-0.5">
                  <UIcon name="i-lucide-grip-vertical" class="w-4 h-4 text-dimmed" />
                </div>
                <span class="text-xs font-semibold text-muted w-5 text-center flex-shrink-0">{{ index + 1 }}</span>
                <span
                  v-if="row.collaboration_item_id && itemNameMap?.[row.collaboration_item_id]"
                  class="flex-shrink-0 text-[11px] font-medium px-1.5 py-0.5 rounded truncate max-w-[100px]"
                  :class="[itemColorMap[row.collaboration_item_id]?.bg, itemColorMap[row.collaboration_item_id]?.text]"
                  :title="itemNameMap[row.collaboration_item_id]"
                >
                  {{ itemNameMap[row.collaboration_item_id] }}
                </span>
                <span
                  v-else
                  class="flex-shrink-0 text-[11px] font-medium px-1.5 py-0.5 rounded truncate max-w-[100px] bg-gray-500/10 text-gray-500 dark:text-gray-400"
                >
                  未分類
                </span>
                <div class="flex-1 grid grid-cols-[1fr_auto_auto_auto] gap-2 items-center">
                  <input
                    v-model="row.name"
                    data-phase-name
                    placeholder="階段名稱"
                    class="w-full rounded-md border border-default bg-default px-2.5 py-1.5 text-sm text-highlighted shadow-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
                  />
                  <input
                    v-model="row.start_date"
                    type="date"
                    class="w-[140px] rounded-md border border-default bg-default px-2 py-1.5 text-sm text-highlighted shadow-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
                    @change="updateEndDate(row)"
                  />
                  <div class="flex items-center gap-1">
                    <input
                      v-model.number="row.duration_days"
                      type="number"
                      min="1"
                      max="365"
                      class="w-[64px] rounded-md border border-default bg-default px-2 py-1.5 text-sm text-highlighted text-center shadow-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
                      @input="updateEndDate(row)"
                    />
                    <span class="text-xs text-dimmed">天</span>
                  </div>
                  <span class="text-xs text-dimmed whitespace-nowrap min-w-[80px] text-right">
                    → {{ row.end_date ? format(parseISO(row.end_date), 'MM/dd') : '-' }}
                  </span>
                </div>
                <button
                  class="text-dimmed hover:text-red-500 transition-colors flex-shrink-0"
                  @click="markDelete(row)"
                >
                  <UIcon name="i-lucide-x" class="w-4 h-4" />
                </button>
              </div>
            </template>
          </draggable>
        </template>

        <!-- 新增按鈕 -->
        <button
          class="w-full flex items-center justify-center gap-1.5 py-2.5 rounded-lg border-2 border-dashed border-default text-muted hover:border-primary hover:text-primary transition-colors"
          @click="addRow"
        >
          <UIcon name="i-lucide-plus" class="w-4 h-4" />
          <span class="text-sm">新增階段</span>
        </button>

        <!-- 已刪除的項目（可復原） -->
        <div v-if="deletedRows.length > 0" class="pt-2 border-t border-default">
          <p class="text-xs text-dimmed mb-1.5">已移除（儲存後生效）</p>
          <div
            v-for="row in deletedRows"
            :key="row.id"
            class="flex items-center justify-between gap-2 py-1.5 px-2 rounded text-sm text-dimmed line-through"
          >
            <span>{{ row.name || '未命名' }}</span>
            <button
              class="text-xs text-primary hover:underline flex-shrink-0"
              @click="undoDelete(row)"
            >
              復原
            </button>
          </div>
        </div>
      </div>
    </template>

    <template #footer>
      <div class="flex items-center justify-between">
        <span class="text-xs text-dimmed">
          {{ rows.length }} 個階段
          <template v-if="deletedRows.length > 0">
            ・{{ deletedRows.length }} 個待刪除
          </template>
        </span>
        <div class="flex gap-2">
          <BaseButton
            color="neutral"
            variant="outline"
            @click="isOpen = false"
          >
            取消
          </BaseButton>
          <BaseButton
            color="primary"
            :loading="saving"
            @click="handleSave"
          >
            儲存
          </BaseButton>
        </div>
      </div>
    </template>
  </BaseModal>
</template>

<style scoped>
/* 拖曳效果 */
:deep(.drag-ghost) {
  opacity: 0.4;
  background: rgba(var(--color-primary-500) / 0.08);
}

:deep(.drag-chosen) {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  z-index: 10;
}

:deep(.drag-dragging) {
  cursor: grabbing !important;
  z-index: 1000;
}

.drag-handle {
  cursor: grab;
}

.drag-handle:active {
  cursor: grabbing;
}
</style>
