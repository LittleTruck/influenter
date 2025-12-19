<script setup lang="ts">
import type { CollaborationItem, CreateCollaborationItemRequest, UpdateCollaborationItemRequest } from '~/types/collaborationItems'
import { useCollaborationItems } from '~/composables/useCollaborationItems'
import { useFormModal } from '~/composables/useFormModal'
import { useErrorHandler } from '~/composables/useErrorHandler'

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
}>({
  title: '',
  description: '',
  price: null,
  parent_id: null
})

// 初始化表單資料
const initForm = () => {
  if (props.item) {
    formData.title = props.item.title
    formData.description = props.item.description || ''
    formData.price = props.item.price
    formData.parent_id = props.item.parent_id || null
  } else {
    formData.title = ''
    formData.description = ''
    formData.price = null
    // 如果是新增子項目，預設父項目 ID
    formData.parent_id = props.parentId || null
  }
}

// 監聽 props 變化
watch([() => props.item, () => props.parentId, () => props.modelValue], () => {
  if (props.modelValue) {
    initForm()
  }
}, { immediate: true })

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
          parent_id: formData.parent_id
        })
        return '項目已更新'
      } else {
        await createItem({
          title: formData.title.trim(),
          description: formData.description?.trim() || undefined,
          price: formData.price!,
          parent_id: formData.parent_id
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
  <UModal
    v-model:open="isOpen"
    :title="isEditMode ? '編輯合作項目' : isAddingChild ? '新增子項目' : '新增合作項目'"
    :ui="{ width: 'max-w-lg' }"
  >
    <template #body>
      <UForm :state="formData" class="space-y-4" @submit.prevent="handleSubmit">
        <!-- 父項目提示（新增子項目時） -->
        <div v-if="isAddingChild && parentItemName" class="p-3 bg-primary-50 dark:bg-primary-900/20 rounded-lg border border-primary-200 dark:border-primary-800">
          <div class="flex items-center gap-2">
            <UIcon name="i-lucide-info" class="w-4 h-4 text-primary-600 dark:text-primary-400 flex-shrink-0" />
            <p class="text-sm text-primary-800 dark:text-primary-200">
              將新增為 <span class="font-semibold">{{ parentItemName }}</span> 的子項目
            </p>
          </div>
        </div>

        <UFormField label="項目名稱" name="title" required>
          <UInput
            v-model="formData.title"
            placeholder="請輸入項目名稱"
            class="w-full"
          />
        </UFormField>

        <UFormField label="描述" name="description">
          <UTextarea
            v-model="formData.description"
            placeholder="請輸入項目描述（選填）"
            :rows="3"
            class="w-full"
          />
        </UFormField>

        <UFormField label="價格" name="price" required>
          <UInput
            v-model.number="formData.price"
            type="number"
            min="0"
            step="0.01"
            placeholder="0"
            class="w-full"
          />
        </UFormField>

        <div class="flex items-center justify-end gap-2 pt-2">
          <UButton
            color="neutral"
            variant="ghost"
            @click="handleCancel"
          >
            取消
          </UButton>
          <UButton
            type="submit"
            :loading="isSubmitting"
          >
            {{ isEditMode ? '更新' : '建立' }}
          </UButton>
        </div>
      </UForm>
    </template>
  </UModal>
</template>

