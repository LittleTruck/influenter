<script setup lang="ts">
import type { CollaborationItem, CreateCollaborationItemRequest, UpdateCollaborationItemRequest } from '~/types/collaborationItems'
import { useCollaborationItems } from '~/composables/useCollaborationItems'
import { useWorkflowTemplates } from '~/composables/useWorkflowTemplates'
import { useFormModal } from '~/composables/useFormModal'
import { useErrorHandler } from '~/composables/useErrorHandler'
import { BaseModal, BaseButton, BaseInput, BaseTextarea, BaseFormField, BaseIcon, BaseSelect } from '~/components/base'

interface Props {
  /** 是否顯示 */
  modelValue: boolean
  /** 項目資料（編輯模式） */
  item?: CollaborationItem | null
  /** 父項目 ID（新增子項目時） */
  parentId?: string | null
}

const props = withDefaults(defineProps<Props>(), {
  item: null,
  parentId: null
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'submit': []
}>()

const { items, flatItems, createItem, updateItem } = useCollaborationItems()
const { workflowTemplates, fetchWorkflows } = useWorkflowTemplates()
const { submitForm, isSubmitting, validateRequired } = useFormModal<CreateCollaborationItemRequest | UpdateCollaborationItemRequest>()
const { handleError } = useErrorHandler()

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const isEditMode = computed(() => !!props.item)
const isAddingChild = computed(() => !!props.parentId && !props.item)

// 表單資料
const formData = reactive<{
  title: string
  description: string
  price: number | null
  parent_id: string | null
  workflow_id: string | null
}>({
  title: '',
  description: '',
  price: null,
  parent_id: null,
  workflow_id: null
})

// 載入流程列表（在 Modal 打開時載入，確保資料是最新的）
watch(() => props.modelValue, async (isOpen) => {
  if (isOpen) {
    await fetchWorkflows()
  }
}, { immediate: true })

// 轉換為選項格式（BaseSelect 需要的格式：label 和 value）
const workflowOptions = computed(() => {
  if (!workflowTemplates.value || workflowTemplates.value.length === 0) {
    return []
  }
  return workflowTemplates.value.map(w => ({
    label: w.name,
    value: w.id
  }))
})

// 取得父項目（用於繼承流程）
const parentItem = computed(() => {
  if (props.parentId) {
    return flatItems.value.find(item => item.id === props.parentId)
  }
  return null
})

// 初始化表單資料
const initForm = () => {
  if (props.item) {
    formData.title = props.item.title
    formData.description = props.item.description || ''
    formData.price = props.item.price
    formData.parent_id = props.item.parent_id || null
    formData.workflow_id = props.item.workflow_id || null
  } else {
    formData.title = ''
    formData.description = ''
    formData.price = null
    // 如果是新增子項目，預設父項目 ID，並繼承父項目的流程
    formData.parent_id = props.parentId || null
    formData.workflow_id = parentItem.value?.workflow_id || null
  }
}

// 監聽 modal 打開時初始化表單
watch(() => props.modelValue, (isOpen) => {
  if (isOpen) {
    initForm()
  }
})

// 取得父項目名稱（用於顯示）
const parentItemName = computed(() => {
  if (!props.parentId) return null
  const parent = flatItems.value.find(item => item.id === props.parentId)
  return parent?.title || null
})


// 提交表單
const handleSubmit = async () => {
  // 驗證必填欄位
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

  const success = await submitForm(
    async () => {
      if (isEditMode.value && props.item) {
        await updateItem(props.item.id, {
          title: formData.title.trim(),
          description: formData.description?.trim() || undefined,
          price: formData.price!,
          parent_id: formData.parent_id,
          workflow_id: formData.workflow_id
        })
        return '項目已更新'
      } else {
        await createItem({
          title: formData.title.trim(),
          description: formData.description?.trim() || undefined,
          price: formData.price!,
          parent_id: formData.parent_id,
          workflow_id: formData.workflow_id
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
    :title="isEditMode ? '編輯合作項目' : isAddingChild ? '新增子項目' : '新增合作項目'"
    size="md"
  >
    <template #body>
      <form class="space-y-4" @submit.prevent="handleSubmit">
        <!-- 父項目提示（新增子項目時） -->
        <div v-if="isAddingChild && parentItemName" class="p-3 bg-primary-50 dark:bg-primary-900/20 rounded-lg border border-primary-200 dark:border-primary-800">
          <div class="flex items-center gap-2">
            <BaseIcon name="i-lucide-info" class="w-4 h-4 text-primary-600 dark:text-primary-400 flex-shrink-0" />
            <p class="text-sm text-primary-800 dark:text-primary-200">
              將新增為 <span class="font-semibold">{{ parentItemName }}</span> 的子項目
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

        <BaseFormField label="價格" name="price" required>
          <BaseInput
            v-model.number="formData.price"
            type="number"
            placeholder="0"
            class="w-full"
          />
        </BaseFormField>

        <BaseFormField label="流程" name="workflow_id">
          <BaseSelect
            v-model="formData.workflow_id"
            :items="workflowOptions"
            value-key="value"
            placeholder="選擇流程（選填）"
            :disabled="!!parentItem && !props.item"
            class="w-full"
          />
          <p v-if="parentItem && !props.item" class="text-xs text-muted mt-1">
            子項目將繼承父項目「{{ parentItem.title }}」的流程
          </p>
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

