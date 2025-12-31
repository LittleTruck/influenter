<script setup lang="ts">
import type { CollaborationItem } from '~/types/collaborationItems'
import { useCollaborationItems } from '~/composables/useCollaborationItems'
import { formatAmount } from '~/utils/formatters'
import { BaseButton, BaseModal, BaseFormField, BaseInput, BaseTextarea } from '~/components/base'
import CollaborationItemOption from './CollaborationItemOption.vue'

interface CollaborationItemInput {
  id?: string
  title: string
  description?: string
  price: number
  isCustom?: boolean
}

interface Props {
  /** 當前選中的項目 */
  items: Array<CollaborationItem & { isCustom?: boolean }>
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'save': [items: CollaborationItemInput[]]
  'cancel': []
}>()

const { items: presetItems, flatItems, fetchItems } = useCollaborationItems()

// 載入預設項目列表
onMounted(async () => {
  if (presetItems.value.length === 0) {
    await fetchItems()
  }
})

// 選中的項目（包括預設和自訂）
const selectedItems = ref<CollaborationItemInput[]>([])

// 初始化選中的項目
watch(() => props.items, (newItems) => {
  selectedItems.value = newItems.map(item => ({
    id: item.id,
    title: item.title,
    description: item.description,
    price: item.price,
    isCustom: item.isCustom
  }))
}, { immediate: true })

// 切換預設項目選中狀態
const togglePresetItem = (itemId: string) => {
  const index = selectedItems.value.findIndex(item => item.id === itemId && !item.isCustom)
  if (index > -1) {
    // 取消選中
    selectedItems.value = selectedItems.value.filter((_, i) => i !== index)
  } else {
    // 選中預設項目
    const item = flatItems.value.find(i => i.id === itemId)
    if (item) {
      selectedItems.value.push({
        id: item.id,
        title: item.title,
        description: item.description,
        price: item.price,
        isCustom: false
      })
    }
  }
}

// 新增自訂項目
const showCustomForm = ref(false)
const customItem = ref<CollaborationItemInput>({
  title: '',
  description: '',
  price: 0
})

const handleAddCustom = () => {
  if (!customItem.value.title || customItem.value.price < 0) {
    return
  }
  
  selectedItems.value.push({
    ...customItem.value,
    isCustom: true
  })
  
  // 重置表單
  customItem.value = {
    title: '',
    description: '',
    price: 0
  }
  showCustomForm.value = false
}

// 移除項目
const removeItem = (index: number) => {
  selectedItems.value = selectedItems.value.filter((_, i) => i !== index)
}

// 編輯自訂項目
const editingIndex = ref<number | null>(null)
const editCustomItem = (index: number) => {
  const item = selectedItems.value[index]
  if (item.isCustom) {
    editingIndex.value = index
    customItem.value = { ...item }
    showCustomForm.value = true
  }
}

const handleSaveCustom = () => {
  if (!customItem.value.title || customItem.value.price < 0) {
    return
  }
  
  if (editingIndex.value !== null) {
    // 更新現有項目
    selectedItems.value[editingIndex.value] = {
      ...customItem.value,
      isCustom: true
    }
    editingIndex.value = null
  } else {
    // 新增項目
    handleAddCustom()
  }
  
  customItem.value = {
    title: '',
    description: '',
    price: 0
  }
  showCustomForm.value = false
}

// 計算總價
const totalPrice = computed(() => {
  return selectedItems.value.reduce((sum, item) => sum + (item.price || 0), 0)
})

// 保存
const handleSave = () => {
  emit('save', selectedItems.value)
}

// 取消
const handleCancel = () => {
  emit('cancel')
}
</script>

<template>
  <div class="collaboration-items-editor space-y-4">
    <!-- 預設項目選擇 -->
    <div>
      <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
        從常用項目選擇
      </label>
      <div class="max-h-48 overflow-y-auto border border-gray-200 dark:border-gray-700 rounded-lg p-2 space-y-1">
        <div v-if="presetItems.length === 0" class="text-sm text-gray-500 dark:text-gray-400 p-4 text-center">
          還沒有常用項目
        </div>
        <CollaborationItemOption
          v-for="item in presetItems"
          :key="item.id"
          :item="item"
          :level="0"
          :selected-ids="selectedItems.filter(i => !i.isCustom).map(i => i.id!).filter(Boolean)"
          @toggle="togglePresetItem"
        />
      </div>
    </div>

    <!-- 自訂項目列表 -->
    <div>
      <div class="flex items-center justify-between mb-2">
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
          自訂項目
        </label>
        <BaseButton
          icon="i-lucide-plus"
          size="xs"
          variant="outline"
          @click="showCustomForm = true"
        >
          新增自訂項目
        </BaseButton>
      </div>
      
      <div v-if="selectedItems.filter(i => i.isCustom).length === 0" class="text-sm text-gray-500 dark:text-gray-400 p-4 text-center border border-gray-200 dark:border-gray-700 rounded-lg">
        尚未添加自訂項目
      </div>
      
      <div v-else class="space-y-2">
        <div
          v-for="(item, index) in selectedItems.filter(i => i.isCustom)"
          :key="index"
          class="flex items-center justify-between p-3 border border-gray-200 dark:border-gray-700 rounded-lg"
        >
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="font-medium text-gray-900 dark:text-white">{{ item.title }}</span>
              <span class="text-xs px-2 py-0.5 bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300 rounded">
                自訂
              </span>
            </div>
            <p v-if="item.description" class="text-sm text-gray-500 dark:text-gray-400 mt-1">
              {{ item.description }}
            </p>
            <span class="text-sm font-semibold text-primary-600 dark:text-primary-400">
              {{ formatAmount(item.price) }}
            </span>
          </div>
          <div class="flex items-center gap-2 ml-4">
            <BaseButton
              icon="i-lucide-edit"
              size="xs"
              variant="ghost"
              @click="editCustomItem(selectedItems.findIndex(i => i === item))"
            />
            <BaseButton
              icon="i-lucide-trash"
              size="xs"
              variant="ghost"
              color="error"
              @click="removeItem(selectedItems.findIndex(i => i === item))"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 自訂項目表單 Modal -->
    <BaseModal v-model="showCustomForm" title="新增自訂項目" size="md">
      <template #body>
        <div class="space-y-4">
          <BaseFormField label="項目名稱" name="title" required>
            <BaseInput
              v-model="customItem.title"
              placeholder="請輸入項目名稱"
              class="w-full"
            />
          </BaseFormField>

          <BaseFormField label="描述" name="description">
            <BaseTextarea
              v-model="customItem.description"
              placeholder="請輸入項目描述（選填）"
              :rows="3"
              class="w-full"
            />
          </BaseFormField>

          <BaseFormField label="價格" name="price" required>
            <BaseInput
              v-model.number="customItem.price"
              type="number"
              placeholder="0"
              class="w-full"
            />
          </BaseFormField>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2">
          <BaseButton variant="ghost" @click="showCustomForm = false">取消</BaseButton>
          <BaseButton @click="handleSaveCustom">確定</BaseButton>
        </div>
      </template>
    </BaseModal>

    <!-- 總價顯示 -->
    <div v-if="selectedItems.length > 0" class="pt-3 border-t border-gray-200 dark:border-gray-700">
      <div class="flex items-center justify-between">
        <span class="text-base font-semibold text-gray-900 dark:text-white">
          總價
        </span>
        <span class="text-xl font-bold text-primary-600 dark:text-primary-400">
          {{ formatAmount(totalPrice) }}
        </span>
      </div>
    </div>

    <!-- 操作按鈕 -->
    <div class="flex items-center justify-end gap-2 pt-2">
      <BaseButton variant="ghost" @click="handleCancel">取消</BaseButton>
      <BaseButton @click="handleSave">儲存</BaseButton>
    </div>
  </div>
</template>




