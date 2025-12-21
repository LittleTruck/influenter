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
  /** 選中的項目 ID 列表或完整項目列表 */
  modelValue?: string[] | CollaborationItemInput[]
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: () => []
})

const emit = defineEmits<{
  'update:modelValue': [value: string[] | CollaborationItemInput[]]
}>()

const { items, flatItems, loading, fetchItems } = useCollaborationItems()

// 載入項目列表
onMounted(async () => {
  await fetchItems()
})

// 判斷是否為完整項目列表（包含自訂項目）
const isFullItemList = computed(() => {
  return Array.isArray(props.modelValue) && props.modelValue.length > 0 && typeof props.modelValue[0] === 'object' && props.modelValue[0] !== null && 'title' in props.modelValue[0]
})

// 選中的項目（統一為完整項目格式）
const selectedItems = computed({
  get: () => {
    // 確保 modelValue 存在且為數組
    if (!props.modelValue || !Array.isArray(props.modelValue)) {
      return []
    }
    
    if (isFullItemList.value) {
      return props.modelValue as CollaborationItemInput[]
    }
    
    // 將 ID 列表轉換為完整項目
    const ids = props.modelValue as string[]
    // 確保 flatItems.value 是數組
    if (!flatItems.value || !Array.isArray(flatItems.value)) {
      return []
    }
    
    return ids.map(id => {
      const item = flatItems.value.find(i => i.id === id)
      if (item) {
        return {
          id: item.id,
          title: item.title,
          description: item.description,
          price: item.price,
          isCustom: false
        }
      }
      return null
    }).filter(Boolean) as CollaborationItemInput[]
  },
  set: (value) => {
    // 如果原本是 ID 列表，則轉換為 ID 列表；否則保持完整項目列表
    if (!isFullItemList.value) {
      emit('update:modelValue', value.map(item => item.id!).filter(Boolean) as string[])
    } else {
      emit('update:modelValue', value)
    }
  }
})

// 切換預設項目選中狀態
const toggleItem = (itemId: string) => {
  if (!flatItems.value || !Array.isArray(flatItems.value)) {
    return
  }
  
  const index = selectedItems.value.findIndex(item => item.id === itemId && !item.isCustom)
  if (index > -1) {
    // 取消選中
    selectedItems.value = selectedItems.value.filter((_, i) => i !== index)
  } else {
    // 選中預設項目
    const item = flatItems.value.find(i => i.id === itemId)
    if (item) {
      selectedItems.value = [...selectedItems.value, {
        id: item.id,
        title: item.title,
        description: item.description,
        price: item.price,
        isCustom: false
      }]
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
  
  selectedItems.value = [...selectedItems.value, {
    ...customItem.value,
    id: `custom_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
    isCustom: true
  }]
  
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
  if (!selectedItems.value || !Array.isArray(selectedItems.value) || index < 0) {
    return
  }
  selectedItems.value = selectedItems.value.filter((_, i) => i !== index)
}

// 計算總價
const totalPrice = computed(() => {
  if (!selectedItems.value || !Array.isArray(selectedItems.value)) {
    return 0
  }
  return selectedItems.value.reduce((sum, item) => sum + (item.price || 0), 0)
})
</script>

<template>
  <div class="collaboration-items-selector space-y-4">
    <!-- 預設項目選擇 -->
    <div>
      <div class="flex items-center justify-between mb-2">
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
          從常用項目選擇
        </label>
      </div>
      <div class="max-h-64 overflow-y-auto border border-gray-200 dark:border-gray-700 rounded-lg p-2 space-y-1">
        <div v-if="loading" class="text-sm text-gray-500 dark:text-gray-400 p-4 text-center">
          載入中...
        </div>
        <div v-else-if="!items || items.length === 0" class="text-sm text-gray-500 dark:text-gray-400 p-4 text-center">
          <p class="mb-2">還沒有合作項目</p>
          <BaseButton
            size="xs"
            variant="outline"
            to="/cases/collaboration-items"
          >
            前往建立
          </BaseButton>
        </div>
        <template v-else>
          <CollaborationItemOption
            v-for="item in items"
            :key="item.id"
            :item="item"
            :level="0"
            :selected-ids="(selectedItems || []).filter(i => !i.isCustom).map(i => i.id!).filter(Boolean)"
            @toggle="toggleItem"
          />
        </template>
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
      
      <div v-if="!selectedItems || selectedItems.filter(i => i.isCustom).length === 0" class="text-sm text-gray-500 dark:text-gray-400 p-4 text-center border border-gray-200 dark:border-gray-700 rounded-lg">
        尚未添加自訂項目
      </div>
      
      <div v-else class="space-y-2">
        <div
          v-for="(item, index) in (selectedItems || []).filter(i => i.isCustom)"
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
          <BaseButton
            icon="i-lucide-trash"
            size="xs"
            variant="ghost"
            color="error"
            @click="removeItem((selectedItems || []).findIndex(i => i === item))"
          />
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
        <div class="flex items-center justify-end gap-2">
          <BaseButton variant="ghost" @click="showCustomForm = false">取消</BaseButton>
          <BaseButton @click="handleAddCustom">確定</BaseButton>
        </div>
      </template>
    </BaseModal>

    <!-- 總價顯示 -->
    <div v-if="selectedItems && selectedItems.length > 0" class="mt-3 p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
      <div class="flex items-center justify-between">
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
          已選 {{ selectedItems.length }} 個項目
        </span>
        <span class="text-lg font-semibold text-primary-600 dark:text-primary-400">
          {{ formatAmount(totalPrice) }}
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.collaboration-item-option {
  transition: background-color 0.2s;
}
</style>
