<script setup lang="ts">
import type { CollaborationItem, CollaborationItemType, CreateCollaborationItemRequest, UpdateCollaborationItemRequest } from '~/types/collaborationItems'
import { useCollaborationItems } from '~/composables/useCollaborationItems'
import { useWorkflowTemplates } from '~/composables/useWorkflowTemplates'
import { useFormModal } from '~/composables/useFormModal'
import { useErrorHandler } from '~/composables/useErrorHandler'
import { formatAmount } from '~/utils/formatters'
import { BaseModal, BaseButton, BaseInput, BaseTextarea, BaseFormField, BaseIcon, BaseSelect } from '~/components/base'

interface Props {
  /** 是否顯示 */
  modelValue: boolean
  /** 項目資料（編輯模式） */
  item?: CollaborationItem | null
  /** 預設類型 */
  defaultType?: CollaborationItemType
}

const props = withDefaults(defineProps<Props>(), {
  item: null,
  defaultType: 'individual'
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'submit': []
}>()

const { individualItems, createItem, updateItem } = useCollaborationItems()
const { workflowTemplates, fetchWorkflows } = useWorkflowTemplates()
const { submitForm, isSubmitting } = useFormModal<CreateCollaborationItemRequest | UpdateCollaborationItemRequest>()
const { handleError } = useErrorHandler()

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const isEditMode = computed(() => !!props.item)

// 表單資料
const formData = reactive<{
  title: string
  description: string
  notes: string
  price: number | null
  type: CollaborationItemType
  workflow_id: string | null
  bundle_item_ids: string[]
}>({
  title: '',
  description: '',
  notes: '',
  price: null,
  type: 'individual',
  workflow_id: null,
  bundle_item_ids: []
})

// 載入流程列表（在 Modal 打開時載入）
watch(() => props.modelValue, async (isOpen) => {
  if (isOpen) {
    await fetchWorkflows()
  }
}, { immediate: true })

// 轉換為選項格式
const workflowOptions = computed(() => {
  if (!workflowTemplates.value || workflowTemplates.value.length === 0) {
    return []
  }
  return workflowTemplates.value.map(w => ({
    label: w.name,
    value: w.id
  }))
})

// 可選的單項列表（排除正在編輯的自己）
const availableIndividualItems = computed(() => {
  return individualItems.value.filter(item => item.id !== props.item?.id)
})

// 類型選項
const typeOptions = [
  { label: '單項', value: 'individual' as CollaborationItemType },
  { label: '組合', value: 'bundle' as CollaborationItemType }
]

// 初始化表單資料
const initForm = () => {
  if (props.item) {
    formData.title = props.item.title
    formData.description = props.item.description || ''
    formData.notes = props.item.notes || ''
    formData.price = props.item.price
    formData.type = props.item.type
    formData.workflow_id = props.item.workflow_id || null
    formData.bundle_item_ids = props.item.bundle_items?.map(ref => ref.item_id) || []
  } else {
    formData.title = ''
    formData.description = ''
    formData.notes = ''
    formData.price = null
    formData.type = props.defaultType || 'individual'
    formData.workflow_id = null
    formData.bundle_item_ids = []
  }
}

// 監聽 modal 打開時初始化表單
watch(() => props.modelValue, (isOpen) => {
  if (isOpen) {
    initForm()
  }
})

// 切換類型時清除不相容的欄位
watch(() => formData.type, (newType) => {
  if (newType === 'bundle') {
    formData.workflow_id = null
  } else {
    formData.bundle_item_ids = []
  }
})

// 切換 bundle 內含項目的選中狀態
const toggleBundleItem = (itemId: string) => {
  const idx = formData.bundle_item_ids.indexOf(itemId)
  if (idx > -1) {
    formData.bundle_item_ids.splice(idx, 1)
  } else {
    formData.bundle_item_ids.push(itemId)
  }
}

// 提交表單
const handleSubmit = async () => {
  if (!formData.title?.trim()) {
    handleError('請輸入項目名稱', '驗證錯誤', { showToast: true, log: false })
    return
  }

  if (formData.price === null || formData.price === undefined) {
    handleError('請輸入價格', '驗證錯誤', { showToast: true, log: false })
    return
  }

  if (formData.price < 0) {
    handleError('價格必須大於或等於 0', '驗證錯誤', { showToast: true, log: false })
    return
  }

  if (formData.type === 'bundle' && formData.bundle_item_ids.length === 0) {
    handleError('組合至少要包含一個單項', '驗證錯誤', { showToast: true, log: false })
    return
  }

  const success = await submitForm(
    async () => {
      if (isEditMode.value && props.item) {
        await updateItem(props.item.id, {
          title: formData.title.trim(),
          description: formData.description?.trim() || undefined,
          notes: formData.notes?.trim() || undefined,
          price: formData.price!,
          workflow_id: formData.type === 'individual' ? formData.workflow_id : null,
          bundle_item_ids: formData.type === 'bundle' ? formData.bundle_item_ids : undefined
        })
        return '項目已更新'
      } else {
        await createItem({
          title: formData.title.trim(),
          description: formData.description?.trim() || undefined,
          notes: formData.notes?.trim() || undefined,
          price: formData.price!,
          type: formData.type,
          workflow_id: formData.type === 'individual' ? formData.workflow_id : null,
          bundle_item_ids: formData.type === 'bundle' ? formData.bundle_item_ids : undefined
        })
        return '項目已建立'
      }
    },
    isEditMode.value ? '項目已更新' : '項目已建立'
  )

  if (success) {
    emit('submit')
    isOpen.value = false
  }
}

const handleCancel = () => {
  isOpen.value = false
}
</script>

<template>
  <BaseModal
    v-model="isOpen"
    :title="isEditMode ? '編輯合作項目' : '新增合作項目'"
    size="md"
  >
    <template #body>
      <form class="space-y-4" @submit.prevent="handleSubmit">
        <!-- 類型選擇（僅新增時可選） -->
        <BaseFormField v-if="!isEditMode" label="類型" name="type" required>
          <div class="flex gap-2">
            <button
              v-for="opt in typeOptions"
              :key="opt.value"
              type="button"
              :class="[
                'px-4 py-2 text-sm rounded-lg border transition-colors',
                formData.type === opt.value
                  ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20 text-primary-700 dark:text-primary-300 font-medium'
                  : 'border-default text-muted hover:bg-subtle'
              ]"
              @click="formData.type = opt.value"
            >
              {{ opt.label }}
            </button>
          </div>
        </BaseFormField>

        <!-- 編輯模式顯示類型提示 -->
        <div v-if="isEditMode" class="p-3 bg-subtle rounded-lg border border-default">
          <div class="flex items-center gap-2">
            <BaseIcon name="i-lucide-info" class="w-4 h-4 text-muted flex-shrink-0" />
            <p class="text-sm text-muted">
              類型：<span class="font-semibold">{{ props.item?.type === 'bundle' ? '組合' : '單項' }}</span>
            </p>
          </div>
        </div>

        <BaseFormField label="項目名稱" name="title" required>
          <BaseInput
            v-model="formData.title"
            placeholder="請輸入項目名稱"
            class="w-full"
          />
        </BaseFormField>

        <BaseFormField label="描述" name="description">
          <BaseTextarea
            v-model="formData.description"
            placeholder="請輸入項目描述（選填）"
            :rows="3"
            class="w-full"
          />
        </BaseFormField>

        <BaseFormField label="注意事項" name="notes">
          <BaseTextarea
            v-model="formData.notes"
            placeholder="請輸入注意事項（選填），例如修改次數限制、解約費用比例、報價有效期等"
            :rows="6"
            class="w-full"
          />
        </BaseFormField>

        <BaseFormField label="價格" name="price" required>
          <BaseInput
            v-model.number="formData.price"
            type="number"
            placeholder="0"
            class="w-full"
          />
        </BaseFormField>

        <!-- 流程選擇（僅 individual） -->
        <BaseFormField v-if="formData.type === 'individual'" label="流程" name="workflow_id">
          <BaseSelect
            v-model="formData.workflow_id"
            :items="workflowOptions"
            value-key="value"
            placeholder="選擇流程（選填）"
            class="w-full"
          />
        </BaseFormField>

        <!-- Bundle 內含項目選擇（僅 bundle） -->
        <BaseFormField v-if="formData.type === 'bundle'" label="包含的單項" name="bundle_items" required>
          <div class="max-h-48 overflow-y-auto border border-default rounded-lg p-2 space-y-1">
            <div v-if="availableIndividualItems.length === 0" class="text-sm text-muted p-4 text-center">
              尚無可選的單項
            </div>
            <div
              v-for="indItem in availableIndividualItems"
              :key="indItem.id"
              :class="[
                'flex items-center gap-2 p-2 rounded hover:bg-subtle cursor-pointer',
                formData.bundle_item_ids.includes(indItem.id) && 'bg-primary-50 dark:bg-primary-900/20'
              ]"
              @click="toggleBundleItem(indItem.id)"
            >
              <input
                type="checkbox"
                :checked="formData.bundle_item_ids.includes(indItem.id)"
                class="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                @click.stop="toggleBundleItem(indItem.id)"
              />
              <span class="flex-1 text-sm text-highlighted">{{ indItem.title }}</span>
              <span class="text-xs text-muted">{{ formatAmount(indItem.price) }}</span>
            </div>
          </div>
        </BaseFormField>

        <div class="flex justify-end gap-2 pt-2">
          <BaseButton
            color="neutral"
            variant="ghost"
            @click="handleCancel"
          >
            取消
          </BaseButton>
          <BaseButton
            type="submit"
            :loading="isSubmitting"
          >
            {{ isEditMode ? '更新' : '建立' }}
          </BaseButton>
        </div>
      </form>
    </template>
  </BaseModal>
</template>
