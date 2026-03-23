<script setup lang="ts">
import type { CasePhase } from '~/types/cases'
import draggable from 'vuedraggable'
import { BaseModal, BaseButton, BaseIcon } from '~/components/base'
import { format, addDays, parseISO } from 'date-fns'

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
        <!-- 空狀態 -->
        <div v-if="rows.length === 0 && deletedRows.length === 0" class="text-center py-8 text-muted">
          <BaseIcon name="i-lucide-list-x" class="w-10 h-10 mx-auto mb-2 opacity-40" />
          <p class="text-sm">尚未設定流程階段</p>
          <p class="text-xs text-dimmed mt-1">點擊下方「新增階段」開始建立</p>
        </div>

        <!-- 可拖曳的階段列表 -->
        <draggable
          v-model="rows"
          v-bind="dragOptions"
          item-key="id"
          class="space-y-2"
        >
          <template #item="{ element: row, index }">
            <div class="flex items-center gap-2 p-3 rounded-lg border border-default bg-subtle transition-shadow">
              <!-- 拖曳把手 -->
              <div class="drag-handle flex items-center self-stretch px-0.5">
                <UIcon name="i-lucide-grip-vertical" class="w-4 h-4 text-dimmed" />
              </div>

              <!-- 序號 -->
              <span class="text-xs font-semibold text-muted w-5 text-center flex-shrink-0">{{ index + 1 }}</span>

              <!-- 合作項目標籤 -->
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

              <!-- 欄位 -->
              <div class="flex-1 grid grid-cols-[1fr_auto_auto_auto] gap-2 items-center">
                <!-- 名稱 -->
                <input
                  v-model="row.name"
                  data-phase-name
                  placeholder="階段名稱"
                  class="w-full rounded-md border border-default bg-default px-2.5 py-1.5 text-sm text-highlighted shadow-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
                />

                <!-- 開始日期 -->
                <input
                  v-model="row.start_date"
                  type="date"
                  class="w-[140px] rounded-md border border-default bg-default px-2 py-1.5 text-sm text-highlighted shadow-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
                  @change="updateEndDate(row)"
                />

                <!-- 天數 -->
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

                <!-- 結束日期（唯讀） -->
                <span class="text-xs text-dimmed whitespace-nowrap min-w-[80px] text-right">
                  → {{ row.end_date ? format(parseISO(row.end_date), 'MM/dd') : '-' }}
                </span>
              </div>

              <!-- 刪除 -->
              <button
                class="text-dimmed hover:text-red-500 transition-colors flex-shrink-0"
                @click="markDelete(row)"
              >
                <UIcon name="i-lucide-x" class="w-4 h-4" />
              </button>
            </div>
          </template>
        </draggable>

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
