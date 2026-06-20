<script setup lang="ts">
import type { CaseTotalAdjustment } from '~/types/cases'
import { BaseButton, BaseIcon } from '~/components/base'
import { useCases } from '~/composables/useCases'
import { useErrorHandler } from '~/composables/useErrorHandler'
import { formatAmount, formatFullDate } from '~/utils/formatters'

interface Props {
  modelValue: boolean
  caseId: string
  /** 合作項目加總（原始總價） */
  originalTotal: number
  /** 目前已套用的調整後總價（未調整時為 undefined） */
  currentAdjusted?: number
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

const { adjustCaseTotal, clearCaseTotalAdjustment, fetchCaseTotalAdjustments } = useCases()
const { handleError, handleSuccess } = useErrorHandler()

const newTotal = ref<string>('')
const reason = ref<string>('')
const saving = ref(false)
const clearing = ref(false)

const history = ref<CaseTotalAdjustment[]>([])
const loadingHistory = ref(false)

const hasAdjustment = computed(() => props.currentAdjusted != null)

const loadHistory = async () => {
  loadingHistory.value = true
  try {
    history.value = await fetchCaseTotalAdjustments(props.caseId)
  } catch {
    history.value = []
  } finally {
    loadingHistory.value = false
  }
}

// 開啟時初始化表單與歷史
watch(isOpen, (open) => {
  if (open) {
    newTotal.value = String(props.currentAdjusted ?? props.originalTotal)
    reason.value = ''
    loadHistory()
  }
})

const save = async () => {
  const parsed = parseFloat(newTotal.value)
  if (isNaN(parsed) || parsed < 0) {
    handleError(new Error('請輸入有效的金額'), '金額格式錯誤')
    return
  }
  saving.value = true
  try {
    await adjustCaseTotal(props.caseId, { adjusted_total: parsed, reason: reason.value.trim() })
    handleSuccess('總價已調整')
    emit('saved')
    isOpen.value = false
  } catch (error: unknown) {
    handleError(error, '調整總價失敗')
  } finally {
    saving.value = false
  }
}

const clear = async () => {
  clearing.value = true
  try {
    await clearCaseTotalAdjustment(props.caseId)
    handleSuccess('已恢復為合作項目加總')
    emit('saved')
    isOpen.value = false
  } catch (error: unknown) {
    handleError(error, '恢復總價失敗')
  } finally {
    clearing.value = false
  }
}
</script>

<template>
  <UModal v-model:open="isOpen">
    <template #content>
      <div class="p-5 space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold text-highlighted">編輯案件總價</h3>
          <BaseButton icon="i-lucide-x" size="xs" variant="ghost" @click="isOpen = false" />
        </div>

        <!-- 原始總價（合作項目加總） -->
        <div class="bg-subtle rounded-lg p-3 flex items-center justify-between">
          <span class="text-sm text-muted">合作項目加總（原始總價）</span>
          <span class="text-sm font-semibold text-highlighted">{{ formatAmount(originalTotal) }}</span>
        </div>

        <!-- 調整後總價 -->
        <div>
          <label class="block text-sm font-medium text-highlighted mb-1">調整後總價</label>
          <div class="flex items-center gap-1.5">
            <span class="text-sm text-muted">$</span>
            <input
              v-model="newTotal"
              type="number"
              min="0"
              step="100"
              class="flex-1 text-sm font-semibold border border-default rounded-md px-3 py-2 bg-default text-highlighted focus:outline-none focus:ring-1 focus:ring-primary-500"
              placeholder="輸入調整後的合作費用"
            />
          </div>
          <p class="text-xs text-dimmed mt-1">單一廠商殺價時，可在此微調此案件的合作費用。</p>
        </div>

        <!-- 調整原因 -->
        <div>
          <label class="block text-sm font-medium text-highlighted mb-1">調整原因 <span class="text-dimmed font-normal">（選填）</span></label>
          <textarea
            v-model="reason"
            rows="3"
            class="w-full text-sm border border-default rounded-md px-3 py-2 bg-default text-highlighted focus:outline-none focus:ring-1 focus:ring-primary-500"
            placeholder="例如：廠商預算有限，議價後調降費用"
          />
        </div>

        <!-- 調整歷史 -->
        <div v-if="loadingHistory || history.length > 0" class="border-t border-default pt-3">
          <div class="text-xs font-medium text-dimmed uppercase mb-2">調整記錄</div>
          <div v-if="loadingHistory" class="text-sm text-muted">載入中…</div>
          <div v-else class="space-y-2 max-h-48 overflow-y-auto">
            <div
              v-for="record in history"
              :key="record.id"
              class="text-sm rounded-lg bg-subtle px-3 py-2"
            >
              <div class="flex items-center gap-2 font-medium text-highlighted">
                <span class="line-through text-muted">{{ formatAmount(record.original_total) }}</span>
                <BaseIcon name="i-lucide-arrow-right" class="w-3.5 h-3.5 text-muted" />
                <span class="text-primary-600 dark:text-primary-400">{{ formatAmount(record.adjusted_total) }}</span>
              </div>
              <p v-if="record.reason" class="text-muted mt-0.5 whitespace-pre-wrap">{{ record.reason }}</p>
              <p v-else class="text-dimmed mt-0.5 italic">未填寫原因</p>
              <p class="text-xs text-dimmed mt-0.5">{{ formatFullDate(record.created_at) }}</p>
            </div>
          </div>
        </div>

        <!-- 動作 -->
        <div class="flex items-center justify-between gap-2 pt-1">
          <BaseButton
            v-if="hasAdjustment"
            variant="ghost"
            color="neutral"
            icon="i-lucide-rotate-ccw"
            :loading="clearing"
            @click="clear"
          >
            恢復原始總價
          </BaseButton>
          <div class="flex items-center gap-2 ml-auto">
            <BaseButton variant="outline" @click="isOpen = false">取消</BaseButton>
            <BaseButton :loading="saving" @click="save">儲存</BaseButton>
          </div>
        </div>
      </div>
    </template>
  </UModal>
</template>
