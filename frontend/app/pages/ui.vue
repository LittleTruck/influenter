<script setup lang="ts">
/**
 * UI 元件展示頁面
 * 
 * 統整所有 Base 組件和 UI 組件的使用方式
 * 展示各種變體、狀態和可能性
 * 訪問路徑: /ui
 */

import { 
  BaseButton, BaseInput, BaseTextarea, BaseSelect, BaseCard, 
  BaseBadge, BaseAvatar, BaseProgress, BaseTabs, BaseTable, 
  BasePagination, BaseAlert, BaseCollapsible, BaseIcon,
  BaseCheckbox, BaseSwitch, BaseModal, BaseSlideover, BaseFormField
} from '~/components/base'
import draggable from 'vuedraggable'

definePageMeta({
  layout: false, // 不使用 layout，確保在 UApp 內直接渲染
})

useSeoMeta({
  title: 'UI 元件展示',
  description: 'Influenter UI 組件完整展示頁面',
})

// 狀態管理
const toast = useToast()
const isModalOpen = ref(false)
const isSlideoverOpen = ref(false)

// Collapsible 狀態
const collapsibleStates = reactive({
  button: true,
  input: true,
  textarea: true,
  select: true,
  checkboxSwitch: true,
  card: true,
  badge: true,
  avatar: true,
  progress: true,
  tabs: true,
  table: true,
  pagination: true,
  alert: true,
  collapsible: true,
  collapsibleInner: false, // 內部嵌套的 Collapsible
  toast: true,
  board: true,
  draggable: true,
})

// 表單數據
const formData = reactive({
  input: '',
  email: '',
  password: '',
  number: 0,
  textarea: '',
  select: '',
  selectWithSearch: '',
  checkbox: false,
  switch: false,
  date: '',
})

// 選擇器選項
const selectOptions = [
  { label: '選項 1', value: '1' },
  { label: '選項 2', value: '2' },
  { label: '選項 3', value: '3' },
  { label: '選項 4', value: '4' },
  { label: '選項 5', value: '5' },
]

// 表格數據
const tableData = ref([
  { id: 1, name: '項目 A', status: 'active', price: 1000, date: '2024-01-15', title: '項目 A', brand_name: '品牌 A' },
  { id: 2, name: '項目 B', status: 'pending', price: 2000, date: '2024-01-16', title: '項目 B', brand_name: '品牌 B' },
  { id: 3, name: '項目 C', status: 'inactive', price: 1500, date: '2024-01-17', title: '項目 C', brand_name: '品牌 C' },
  { id: 4, name: '項目 D', status: 'active', price: 3000, date: '2024-01-18', title: '項目 D', brand_name: '品牌 D' },
  { id: 5, name: '項目 E', status: 'pending', price: 2500, date: '2024-01-19', title: '項目 E', brand_name: '品牌 E' },
  { id: 6, name: '項目 F', status: 'active', price: 1800, date: '2024-01-20', title: '項目 F', brand_name: '品牌 F' },
])

// 看板卡片數據
const boardCards = reactive({
  to_confirm: [
    { id: '1', title: '待確認案件 1', brand_name: '品牌 A', quoted_amount: 10000, currency: 'TWD', deadline_date: '2024-12-25', status: 'to_confirm' },
    { id: '2', title: '待確認案件 2', brand_name: '品牌 B', quoted_amount: 20000, currency: 'TWD', deadline_date: '2024-12-30', status: 'to_confirm' },
  ],
  in_progress: [
    { id: '3', title: '進行中案件 1', brand_name: '品牌 C', quoted_amount: 15000, currency: 'TWD', deadline_date: '2025-01-05', status: 'in_progress' },
  ],
  completed: [
    { id: '4', title: '已完成案件 1', brand_name: '品牌 D', quoted_amount: 30000, currency: 'TWD', deadline_date: '2024-12-20', status: 'completed' },
  ],
  cancelled: [],
})

// 拖曳條目數據
const draggableItems = ref([
  { id: '1', name: '拖曳項目 1', description: '這是第一個可拖曳的項目' },
  { id: '2', name: '拖曳項目 2', description: '這是第二個可拖曳的項目' },
  { id: '3', name: '拖曳項目 3', description: '這是第三個可拖曳的項目' },
  { id: '4', name: '拖曳項目 4', description: '這是第四個可拖曳的項目' },
])

// 分頁狀態
const currentPage = ref(1)
const pageSize = ref(10)
const totalPages = ref(5)

// Tabs 狀態
const activeTab = ref('tab1')

// 進度條狀態
const progress = ref(45)

// Alert 顯示狀態
const alertStates = reactive({
  success: false,
  warning: false,
  error: false,
  info: false,
})

// 測試函數
const showToast = (type: 'success' | 'error' | 'warning' | 'info', message: string) => {
  toast.add({
    title: message,
    color: type,
    description: type === 'success' ? '操作成功完成' : type === 'error' ? '發生錯誤' : type === 'warning' ? '請注意' : '提示訊息',
  })
}

const showAlert = (type: 'success' | 'warning' | 'error' | 'info') => {
  alertStates[type] = true
}

const openModal = () => {
  isModalOpen.value = true
}

const closeModal = () => {
  isModalOpen.value = false
}

const openSlideover = () => {
  isSlideoverOpen.value = true
}

const closeSlideover = () => {
  isSlideoverOpen.value = false
}
</script>

<template>
  <div class="min-h-screen w-full bg-gray-50 dark:bg-gray-900 overflow-y-auto">
    <!-- 頂部導航 -->
    <nav class="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700 sticky top-0 z-50">
      <div class="w-full px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between items-center h-16">
          <div class="flex items-center gap-4">
            <BaseIcon name="i-lucide-palette" class="w-6 h-6 text-primary-600 dark:text-primary-400" />
            <h1 class="text-xl font-bold text-gray-900 dark:text-white">
              UI 元件展示
            </h1>
            <BaseBadge color="primary" variant="soft">開發用</BaseBadge>
          </div>
          <div class="flex items-center gap-2">
            <BaseButton
              @click="openModal"
              color="primary"
            >
              測試 Modal
            </BaseButton>
            <BaseButton
              @click="openSlideover"
              color="secondary"
            >
              測試 Slideover
            </BaseButton>
          </div>
        </div>
      </div>
    </nav>

    <!-- 主要內容 -->
    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="space-y-6">
        
        <!-- BaseButton 按鈕組件 -->
        <BaseCard>
          <BaseCollapsible v-model:open="collapsibleStates.button">
            <div class="flex items-center gap-2 cursor-pointer p-4">
              <BaseIcon name="i-lucide-mouse-pointer-click" class="w-5 h-5" />
              <h2 class="text-2xl font-bold">BaseButton 按鈕</h2>
              <BaseIcon 
                :name="collapsibleStates.button ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" 
                class="w-5 h-5 ml-auto"
              />
            </div>
            <template #content>
              <div class="px-4 pb-4 space-y-6">
                  <!-- 顏色變體 -->
                  <div>
                    <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">顏色變體</h3>
                    <div class="flex flex-wrap gap-3">
                      <BaseButton color="primary">Primary</BaseButton>
                      <BaseButton color="secondary">Secondary</BaseButton>
                      <BaseButton color="success">Success</BaseButton>
                      <BaseButton color="warning">Warning</BaseButton>
                      <BaseButton color="error">Error</BaseButton>
                      <BaseButton color="neutral">Neutral</BaseButton>
                    </div>
                  </div>

                  <!-- 樣式變體 -->
                  <div>
                    <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">樣式變體</h3>
                    <div class="flex flex-wrap gap-3">
                      <BaseButton variant="solid" color="primary">Solid</BaseButton>
                      <BaseButton variant="outline" color="primary">Outline</BaseButton>
                      <BaseButton variant="soft" color="primary">Soft</BaseButton>
                      <BaseButton variant="ghost" color="primary">Ghost</BaseButton>
                    </div>
                  </div>

                  <!-- 尺寸 -->
                  <div>
                    <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">尺寸</h3>
                    <div class="flex flex-wrap items-center gap-3">
                      <BaseButton size="xs">Extra Small</BaseButton>
                      <BaseButton size="sm">Small</BaseButton>
                      <BaseButton size="md">Medium</BaseButton>
                      <BaseButton size="lg">Large</BaseButton>
                      <BaseButton size="xl">Extra Large</BaseButton>
                    </div>
                  </div>

                  <!-- 圖示按鈕 -->
                  <div>
                    <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">圖示按鈕</h3>
                    <div class="flex flex-wrap gap-3">
                      <BaseButton icon="i-lucide-plus">新增</BaseButton>
                      <BaseButton icon="i-lucide-edit" variant="outline">編輯</BaseButton>
                      <BaseButton icon="i-lucide-trash" color="error" variant="ghost">刪除</BaseButton>
                      <BaseButton icon="i-lucide-download" variant="soft">下載</BaseButton>
                      <BaseButton icon="i-lucide-search" square />
                      <BaseButton icon="i-lucide-settings" square variant="ghost" />
                    </div>
                  </div>

                  <!-- 狀態 -->
                  <div>
                    <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">狀態</h3>
                    <div class="flex flex-wrap gap-3">
                      <BaseButton loading>載入中</BaseButton>
                      <BaseButton disabled>禁用</BaseButton>
                      <BaseButton :loading="true" color="success">載入成功</BaseButton>
                    </div>
                  </div>

                  <!-- 區塊按鈕 -->
                  <div>
                    <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">區塊按鈕</h3>
                    <BaseButton block color="primary">完整寬度按鈕</BaseButton>
                  </div>
              </div>
            </template>
          </BaseCollapsible>
        </BaseCard>

        <!-- BaseInput 輸入框 -->
        <BaseCard>
          <BaseCollapsible v-model:open="collapsibleStates.input">
            <div class="flex items-center gap-2 cursor-pointer p-4">
              <BaseIcon name="i-lucide-edit" class="w-5 h-5" />
              <h2 class="text-2xl font-bold">BaseInput 輸入框</h2>
              <BaseIcon 
                :name="collapsibleStates.input ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" 
                class="w-5 h-5 ml-auto"
              />
            </div>
            <template #content>
              <div class="px-4 pb-4 space-y-6">
                  <!-- 基本輸入 -->
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <BaseFormField label="文字輸入">
                      <BaseInput v-model="formData.input" placeholder="請輸入文字" class="w-full" />
                    </BaseFormField>
                    <BaseFormField label="郵箱輸入">
                      <BaseInput v-model="formData.email" type="email" placeholder="請輸入郵箱" class="w-full" />
                    </BaseFormField>
                    <BaseFormField label="密碼輸入">
                      <BaseInput v-model="formData.password" type="password" placeholder="請輸入密碼" class="w-full" />
                    </BaseFormField>
                    <BaseFormField label="數字輸入">
                      <BaseInput v-model="formData.number" type="number" placeholder="請輸入數字" class="w-full" />
                    </BaseFormField>
                  </div>

                  <!-- 帶圖示 -->
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <BaseFormField label="搜尋輸入">
                      <BaseInput icon="i-lucide-search" placeholder="搜尋..." class="w-full" />
                    </BaseFormField>
                    <BaseFormField label="使用者名稱">
                      <BaseInput icon="i-lucide-user" placeholder="使用者名稱" class="w-full" />
                    </BaseFormField>
                  </div>

                  <!-- 狀態 -->
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <BaseFormField label="正常狀態">
                      <BaseInput placeholder="正常狀態" class="w-full" />
                    </BaseFormField>
                    <BaseFormField label="禁用狀態">
                      <BaseInput placeholder="禁用狀態" disabled class="w-full" />
                    </BaseFormField>
                    <BaseFormField label="必填欄位" required>
                      <BaseInput placeholder="必填欄位" required class="w-full" />
                    </BaseFormField>
                  </div>

                  <!-- 尺寸 -->
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <BaseFormField label="Extra Small">
                      <BaseInput size="xs" placeholder="Extra Small" class="w-full" />
                    </BaseFormField>
                    <BaseFormField label="Small">
                      <BaseInput size="sm" placeholder="Small" class="w-full" />
                    </BaseFormField>
                    <BaseFormField label="Medium">
                      <BaseInput size="md" placeholder="Medium" class="w-full" />
                    </BaseFormField>
                    <BaseFormField label="Large">
                      <BaseInput size="lg" placeholder="Large" class="w-full" />
                    </BaseFormField>
                  </div>
              </div>
            </template>
          </BaseCollapsible>
        </BaseCard>

        <!-- BaseTextarea 多行輸入 -->
        <BaseCard>
          <BaseCollapsible v-model:open="collapsibleStates.textarea">
            <div class="flex items-center gap-2 cursor-pointer p-4">
              <BaseIcon name="i-lucide-file-text" class="w-5 h-5" />
              <h2 class="text-2xl font-bold">BaseTextarea 多行輸入</h2>
              <BaseIcon 
                :name="collapsibleStates.textarea ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" 
                class="w-5 h-5 ml-auto"
              />
            </div>
            <template #content>
              <div class="px-4 pb-4 space-y-4">
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <BaseFormField label="多行文字">
                      <BaseTextarea v-model="formData.textarea" placeholder="請輸入多行文字..." :rows="4" class="w-full" />
                    </BaseFormField>
                    <BaseFormField label="禁用狀態">
                      <BaseTextarea placeholder="禁用狀態" disabled :rows="4" class="w-full" />
                    </BaseFormField>
                    <BaseFormField label="必填欄位" required>
                      <BaseTextarea placeholder="必填欄位" required :rows="4" class="w-full" />
                    </BaseFormField>
                </div>
              </div>
            </template>
          </BaseCollapsible>
        </BaseCard>

        <!-- BaseSelect 選擇器 -->
        <BaseCard>
          <BaseCollapsible v-model:open="collapsibleStates.select">
            <div class="flex items-center gap-2 cursor-pointer p-4">
              <BaseIcon name="i-lucide-list" class="w-5 h-5" />
              <h2 class="text-2xl font-bold">BaseSelect 選擇器</h2>
              <BaseIcon 
                :name="collapsibleStates.select ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" 
                class="w-5 h-5 ml-auto"
              />
            </div>
            <template #content>
              <div class="px-4 pb-4 space-y-4">
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <BaseFormField label="基本選擇器（無搜尋）">
                      <BaseSelect v-model="formData.select" :items="selectOptions" placeholder="請選擇選項" class="w-full" />
                    </BaseFormField>
                    <BaseFormField label="可搜尋選擇器">
                      <BaseSelect v-model="formData.selectWithSearch" :items="selectOptions" placeholder="請選擇選項（可搜尋）" :searchable="true" class="w-full" />
                    </BaseFormField>
                    <BaseFormField label="禁用狀態">
                      <BaseSelect :items="selectOptions" placeholder="禁用狀態" disabled class="w-full" />
                    </BaseFormField>
                    <BaseFormField label="錯誤狀態" error="請選擇一個有效的選項">
                      <BaseSelect :items="selectOptions" placeholder="錯誤狀態" class="w-full" />
                    </BaseFormField>
                  </div>
              </div>
            </template>
          </BaseCollapsible>
        </BaseCard>

        <!-- BaseCheckbox & BaseSwitch -->
        <BaseCard>
          <BaseCollapsible v-model:open="collapsibleStates.checkboxSwitch">
            <div class="flex items-center gap-2 cursor-pointer p-4">
              <BaseIcon name="i-lucide-toggle-left" class="w-5 h-5" />
              <h2 class="text-2xl font-bold">BaseCheckbox & BaseSwitch</h2>
              <BaseIcon 
                :name="collapsibleStates.checkboxSwitch ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" 
                class="w-5 h-5 ml-auto"
              />
            </div>
            <template #content>
              <div class="px-4 pb-4 grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div class="space-y-4">
                    <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300">BaseCheckbox</h3>
                    <div class="space-y-2">
                      <BaseCheckbox v-model="formData.checkbox" label="選項 1" />
                      <BaseCheckbox label="選項 2（禁用）" disabled />
                      <BaseCheckbox label="選項 3（已選）" :model-value="true" />
                    </div>
                  </div>
                  <div class="space-y-4">
                    <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300">BaseSwitch</h3>
                    <div class="space-y-2">
                      <BaseSwitch v-model="formData.switch" label="開關 1" />
                      <BaseSwitch label="開關 2（禁用）" disabled />
                      <BaseSwitch label="開關 3（開啟）" :model-value="true" />
                    </div>
                  </div>
              </div>
            </template>
          </BaseCollapsible>
        </BaseCard>

        <!-- BaseCard 卡片 -->
        <BaseCard>
          <BaseCollapsible v-model:open="collapsibleStates.card">
            <div class="flex items-center gap-2 cursor-pointer p-4">
              <BaseIcon name="i-lucide-square" class="w-5 h-5" />
              <h2 class="text-2xl font-bold">BaseCard 卡片</h2>
              <BaseIcon 
                :name="collapsibleStates.card ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" 
                class="w-5 h-5 ml-auto"
              />
            </div>
            <template #content>
              <div class="px-4 pb-4 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  <BaseCard>
                    <template #header>
                      <h3 class="font-semibold">基本卡片</h3>
                    </template>
                    <p class="text-sm text-gray-600 dark:text-gray-400">這是基本卡片的內容區域。</p>
                  </BaseCard>

                  <BaseCard>
                    <template #header>
                      <div class="flex items-center justify-between">
                        <h3 class="font-semibold">帶操作</h3>
                        <BaseButton size="xs" icon="i-lucide-more-horizontal" variant="ghost" />
                      </div>
                    </template>
                    <p class="text-sm text-gray-600 dark:text-gray-400 mb-3">卡片內容可以包含任何內容。</p>
                    <template #footer>
                      <div class="flex gap-2">
                        <BaseButton size="sm" variant="outline">取消</BaseButton>
                        <BaseButton size="sm">確認</BaseButton>
                      </div>
                    </template>
                  </BaseCard>

                  <BaseCard>
                    <div class="flex items-center gap-3">
                      <BaseAvatar src="https://i.pravatar.cc/150?img=1" alt="Avatar" />
                      <div>
                        <p class="font-semibold">使用者名稱</p>
                        <p class="text-sm text-gray-500 dark:text-gray-400">使用者描述</p>
                      </div>
                    </div>
                  </BaseCard>
              </div>
            </template>
          </BaseCollapsible>
        </BaseCard>

        <!-- BaseBadge 徽章 -->
        <BaseCard>
          <BaseCollapsible v-model:open="collapsibleStates.badge">
            <div class="flex items-center gap-2 cursor-pointer p-4">
              <BaseIcon name="i-lucide-tag" class="w-5 h-5" />
              <h2 class="text-2xl font-bold">BaseBadge 徽章</h2>
              <BaseIcon 
                :name="collapsibleStates.badge ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" 
                class="w-5 h-5 ml-auto"
              />
            </div>
            <template #content>
              <div class="px-4 pb-4 space-y-4">
                  <div>
                    <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">顏色</h3>
                    <div class="flex flex-wrap gap-3">
                      <BaseBadge color="primary">Primary</BaseBadge>
                      <BaseBadge color="secondary">Secondary</BaseBadge>
                      <BaseBadge color="success">Success</BaseBadge>
                      <BaseBadge color="warning">Warning</BaseBadge>
                      <BaseBadge color="error">Error</BaseBadge>
                      <BaseBadge color="neutral">Neutral</BaseBadge>
                    </div>
                  </div>

                  <div>
                    <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">樣式</h3>
                    <div class="flex flex-wrap gap-3">
                      <BaseBadge variant="solid" color="primary">Solid</BaseBadge>
                      <BaseBadge variant="outline" color="primary">Outline</BaseBadge>
                      <BaseBadge variant="soft" color="primary">Soft</BaseBadge>
                    </div>
                  </div>
              </div>
            </template>
          </BaseCollapsible>
        </BaseCard>

        <!-- BaseAvatar 頭像 -->
        <BaseCard>
          <BaseCollapsible v-model:open="collapsibleStates.avatar">
            <div class="flex items-center gap-2 cursor-pointer p-4">
              <BaseIcon name="i-lucide-user" class="w-5 h-5" />
              <h2 class="text-2xl font-bold">BaseAvatar 頭像</h2>
              <BaseIcon 
                :name="collapsibleStates.avatar ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" 
                class="w-5 h-5 ml-auto"
              />
            </div>
            <template #content>
              <div class="px-4 pb-4 flex flex-wrap items-center gap-6">
                  <div class="flex flex-col items-center gap-2">
                    <BaseAvatar src="https://i.pravatar.cc/150?img=1" alt="Avatar 1" />
                    <span class="text-xs text-gray-500">圖片</span>
                  </div>
                  <div class="flex flex-col items-center gap-2">
                    <BaseAvatar alt="Initials">AB</BaseAvatar>
                    <span class="text-xs text-gray-500">文字</span>
                  </div>
                  <div class="flex flex-col items-center gap-2">
                    <BaseAvatar icon="i-lucide-user" />
                    <span class="text-xs text-gray-500">圖示</span>
                  </div>
                  <div class="flex flex-col items-center gap-2">
                    <BaseAvatar src="https://invalid-url-that-does-not-exist.com/avatar.jpg" alt="Fallback" />
                    <span class="text-xs text-gray-500">Fallback</span>
                  </div>
              </div>
            </template>
          </BaseCollapsible>
        </BaseCard>

        <!-- BaseProgress 進度條 -->
        <BaseCard>
          <BaseCollapsible v-model:open="collapsibleStates.progress">
            <div class="flex items-center gap-2 cursor-pointer p-4">
              <BaseIcon name="i-lucide-activity" class="w-5 h-5" />
              <h2 class="text-2xl font-bold">BaseProgress 進度條</h2>
              <BaseIcon 
                :name="collapsibleStates.progress ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" 
                class="w-5 h-5 ml-auto"
              />
            </div>
            <template #content>
              <div class="px-4 pb-4 space-y-4 max-w-2xl">
                  <div>
                    <div class="flex justify-between mb-1">
                      <span class="text-sm font-medium">進度 {{ progress }}%</span>
                      <span class="text-sm text-gray-500">{{ progress }}%</span>
                    </div>
                    <BaseProgress :value="progress" />
                  </div>
                  <div>
                    <BaseProgress :value="75" color="success" />
                  </div>
                  <div>
                    <BaseProgress :value="50" color="warning" />
                  </div>
                  <div>
                    <BaseProgress :value="25" color="error" />
                  </div>
              </div>
            </template>
          </BaseCollapsible>
        </BaseCard>

        <!-- BaseTabs 標籤頁 -->
        <BaseCard>
          <BaseCollapsible v-model:open="collapsibleStates.tabs">
            <div class="flex items-center gap-2 cursor-pointer p-4">
              <BaseIcon name="i-lucide-folder" class="w-5 h-5" />
              <h2 class="text-2xl font-bold">BaseTabs 標籤頁</h2>
              <BaseIcon 
                :name="collapsibleStates.tabs ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" 
                class="w-5 h-5 ml-auto"
              />
            </div>
            <template #content>
              <div class="px-4 pb-4 max-w-2xl">
                  <BaseTabs v-model="activeTab" :items="[
                    { label: '標籤 1', value: 'tab1' },
                    { label: '標籤 2', value: 'tab2' },
                    { label: '標籤 3', value: 'tab3' },
                  ]">
                    <template #tab1>
                      <div class="p-4">
                        <p>這是標籤 1 的內容</p>
                      </div>
                    </template>
                    <template #tab2>
                      <div class="p-4">
                        <p>這是標籤 2 的內容</p>
                      </div>
                    </template>
                    <template #tab3>
                      <div class="p-4">
                        <p>這是標籤 3 的內容</p>
                      </div>
                    </template>
                  </BaseTabs>
              </div>
            </template>
          </BaseCollapsible>
        </BaseCard>

        <!-- BaseTable 表格 -->
        <BaseCard>
          <BaseCollapsible v-model:open="collapsibleStates.table">
            <div class="flex items-center gap-2 cursor-pointer p-4">
              <BaseIcon name="i-lucide-table" class="w-5 h-5" />
              <h2 class="text-2xl font-bold">BaseTable 表格</h2>
              <BaseIcon 
                :name="collapsibleStates.table ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" 
                class="w-5 h-5 ml-auto"
              />
            </div>
            <template #content>
              <div class="px-4 pb-4">
                <div class="base-card border border-default/70 bg-default/90 bg-elevated/25 rounded-lg overflow-hidden shadow-sm">
                  <BaseTable 
                    :columns="[
                      { accessorKey: 'title', header: '標題' },
                      { accessorKey: 'status', header: '狀態' },
                      { accessorKey: 'price', header: '價格' },
                      { accessorKey: 'date', header: '日期' },
                    ]" 
                    :data="tableData"
                    :ui="{
                      base: 'min-w-full',
                      thead: '[&>tr]:bg-gray-50/50 dark:[&>tr]:bg-gray-800/30 [&>tr]:after:content-none',
                      tbody: '[&>tr]:last:[&>td]:border-b-0 [&>tr]:hover:bg-gray-50 dark:[&>tr]:hover:bg-gray-800/20 [&>tr]:transition-colors [&>tr]:cursor-pointer',
                      th: 'px-4 py-2 text-left text-xs font-medium text-gray-600 dark:text-gray-400 border-b border-gray-200 dark:border-gray-700',
                      td: 'px-4 py-2.5 text-sm text-gray-900 dark:text-gray-100 border-b border-gray-200 dark:border-gray-700 cursor-pointer'
                    }"
                  >
                    <template #title-data="{ row }">
                      <div class="flex flex-col">
                        <p class="font-medium text-gray-900 dark:text-white">{{ row.title }}</p>
                        <p class="text-sm text-gray-500 dark:text-gray-400">{{ row.brand_name }}</p>
                      </div>
                    </template>
                    <template #status-data="{ row }">
                      <BaseBadge 
                        :color="row.status === 'active' ? 'success' : row.status === 'pending' ? 'warning' : 'neutral'"
                        variant="soft"
                        size="sm"
                      >
                        {{ row.status }}
                      </BaseBadge>
                    </template>
                    <template #price-data="{ row }">
                      <span class="font-medium text-gray-900 dark:text-gray-100">NT$ {{ row.price.toLocaleString() }}</span>
                    </template>
                  </BaseTable>
                </div>
              </div>
            </template>
          </BaseCollapsible>
        </BaseCard>

        <!-- BasePagination 分頁 -->
        <BaseCard>
          <BaseCollapsible v-model:open="collapsibleStates.pagination">
            <div class="flex items-center gap-2 cursor-pointer p-4">
              <BaseIcon name="i-lucide-chevrons-left" class="w-5 h-5" />
              <h2 class="text-2xl font-bold">BasePagination 分頁</h2>
              <BaseIcon 
                :name="collapsibleStates.pagination ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" 
                class="w-5 h-5 ml-auto"
              />
            </div>
            <template #content>
              <div class="px-4 pb-4">
                  <BasePagination 
                    v-model="currentPage" 
                    :total="totalPages"
                    :page-size="pageSize"
                  />
                </div>
              </template>
          </BaseCollapsible>
        </BaseCard>

        <!-- 看板（Kanban Board） -->
        <BaseCard>
          <BaseCollapsible v-model:open="collapsibleStates.board">
            <div class="flex items-center gap-2 cursor-pointer p-4">
              <BaseIcon name="i-lucide-layout-grid" class="w-5 h-5" />
              <h2 class="text-2xl font-bold">看板（Kanban Board）</h2>
              <BaseIcon 
                :name="collapsibleStates.board ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" 
                class="w-5 h-5 ml-auto"
              />
            </div>
            <template #content>
              <div class="px-4 pb-4">
                <div class="flex gap-4 overflow-x-auto -mx-4 px-4 pb-4">
                  <!-- 待確認 -->
                  <div class="flex flex-col min-w-[20rem] w-[20rem] flex-shrink-0">
                    <div class="flex flex-col h-full bg-gray-50 dark:bg-gray-800/30 rounded-xl p-4">
                      <div class="flex items-center gap-2 mb-3 pb-3 border-b border-gray-200 dark:border-gray-700">
                        <div class="h-2 w-2 rounded-full bg-yellow-500 flex-shrink-0"></div>
                        <span class="font-semibold text-sm text-gray-700 dark:text-gray-300">待確認</span>
                        <BaseBadge color="neutral" variant="subtle" size="xs" class="ml-auto">
                          {{ boardCards.to_confirm.length }}
                        </BaseBadge>
                      </div>
                      <draggable
                        v-model="boardCards.to_confirm"
                        :group="{ name: 'board-cards', pull: true, put: true }"
                        item-key="id"
                        :animation="200"
                        ghost-class="drag-ghost"
                        chosen-class="drag-chosen"
                        drag-class="drag-dragging"
                        class="space-y-2 min-h-[100px]"
                      >
                        <template #item="{ element }">
                          <div class="board-card-item">
                            <BaseCard class="cursor-move">
                              <div class="p-3">
                                <p class="font-medium text-sm mb-1">{{ element.title }}</p>
                                <p class="text-xs text-gray-500 dark:text-gray-400 mb-2">{{ element.brand_name }}</p>
                                <p class="text-xs font-semibold text-gray-900 dark:text-white">
                                  NT$ {{ element.quoted_amount?.toLocaleString() }}
                                </p>
                              </div>
                            </BaseCard>
                          </div>
                        </template>
                      </draggable>
                    </div>
                  </div>

                  <!-- 進行中 -->
                  <div class="flex flex-col min-w-[20rem] w-[20rem] flex-shrink-0">
                    <div class="flex flex-col h-full bg-gray-50 dark:bg-gray-800/30 rounded-xl p-4">
                      <div class="flex items-center gap-2 mb-3 pb-3 border-b border-gray-200 dark:border-gray-700">
                        <div class="h-2 w-2 rounded-full bg-blue-500 flex-shrink-0"></div>
                        <span class="font-semibold text-sm text-gray-700 dark:text-gray-300">進行中</span>
                        <BaseBadge color="neutral" variant="subtle" size="xs" class="ml-auto">
                          {{ boardCards.in_progress.length }}
                        </BaseBadge>
                      </div>
                      <draggable
                        v-model="boardCards.in_progress"
                        :group="{ name: 'board-cards', pull: true, put: true }"
                        item-key="id"
                        :animation="200"
                        ghost-class="drag-ghost"
                        chosen-class="drag-chosen"
                        drag-class="drag-dragging"
                        class="space-y-2 min-h-[100px]"
                      >
                        <template #item="{ element }">
                          <div class="board-card-item">
                            <BaseCard class="cursor-move">
                              <div class="p-3">
                                <p class="font-medium text-sm mb-1">{{ element.title }}</p>
                                <p class="text-xs text-gray-500 dark:text-gray-400 mb-2">{{ element.brand_name }}</p>
                                <p class="text-xs font-semibold text-gray-900 dark:text-white">
                                  NT$ {{ element.quoted_amount?.toLocaleString() }}
                                </p>
                              </div>
                            </BaseCard>
                          </div>
                        </template>
                      </draggable>
                    </div>
                  </div>

                  <!-- 已完成 -->
                  <div class="flex flex-col min-w-[20rem] w-[20rem] flex-shrink-0">
                    <div class="flex flex-col h-full bg-gray-50 dark:bg-gray-800/30 rounded-xl p-4">
                      <div class="flex items-center gap-2 mb-3 pb-3 border-b border-gray-200 dark:border-gray-700">
                        <div class="h-2 w-2 rounded-full bg-green-500 flex-shrink-0"></div>
                        <span class="font-semibold text-sm text-gray-700 dark:text-gray-300">已完成</span>
                        <BaseBadge color="neutral" variant="subtle" size="xs" class="ml-auto">
                          {{ boardCards.completed.length }}
                        </BaseBadge>
                      </div>
                      <draggable
                        v-model="boardCards.completed"
                        :group="{ name: 'board-cards', pull: true, put: true }"
                        item-key="id"
                        :animation="200"
                        ghost-class="drag-ghost"
                        chosen-class="drag-chosen"
                        drag-class="drag-dragging"
                        class="space-y-2 min-h-[100px]"
                      >
                        <template #item="{ element }">
                          <div class="board-card-item">
                            <BaseCard class="cursor-move">
                              <div class="p-3">
                                <p class="font-medium text-sm mb-1">{{ element.title }}</p>
                                <p class="text-xs text-gray-500 dark:text-gray-400 mb-2">{{ element.brand_name }}</p>
                                <p class="text-xs font-semibold text-gray-900 dark:text-white">
                                  NT$ {{ element.quoted_amount?.toLocaleString() }}
                                </p>
                              </div>
                            </BaseCard>
                          </div>
                        </template>
                      </draggable>
                    </div>
                  </div>
                </div>
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-4">
                  提示：您可以拖曳卡片在不同欄位之間移動
                </p>
              </div>
            </template>
          </BaseCollapsible>
        </BaseCard>

        <!-- 拖曳條目（Draggable） -->
        <BaseCard>
          <BaseCollapsible v-model:open="collapsibleStates.draggable">
            <div class="flex items-center gap-2 cursor-pointer p-4">
              <BaseIcon name="i-lucide-move" class="w-5 h-5" />
              <h2 class="text-2xl font-bold">拖曳條目（Draggable）</h2>
              <BaseIcon 
                :name="collapsibleStates.draggable ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" 
                class="w-5 h-5 ml-auto"
              />
            </div>
            <template #content>
              <div class="px-4 pb-4">
                <p class="text-sm text-gray-600 dark:text-gray-400 mb-4">
                  使用 vuedraggable 實現的可拖曳列表，類似於合作項目管理中的拖曳功能
                </p>
                <div class="space-y-2">
                  <draggable
                    v-model="draggableItems"
                    item-key="id"
                    handle=".drag-handle"
                    :animation="200"
                    ghost-class="drag-ghost"
                    chosen-class="drag-chosen"
                    drag-class="drag-dragging"
                    class="space-y-2"
                  >
                    <template #item="{ element }">
                      <BaseCard class="cursor-move">
                        <div class="flex items-center gap-3 p-4">
                          <BaseIcon name="i-lucide-grip-vertical" class="w-5 h-5 text-gray-400 drag-handle cursor-grab" />
                          <div class="flex-1">
                            <p class="font-medium text-gray-900 dark:text-white">{{ element.name }}</p>
                            <p class="text-sm text-gray-500 dark:text-gray-400">{{ element.description }}</p>
                          </div>
                          <BaseBadge color="neutral" variant="soft" size="sm">{{ element.id }}</BaseBadge>
                        </div>
                      </BaseCard>
                    </template>
                  </draggable>
                </div>
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-4">
                  提示：拖曳左側的圖標來重新排序項目
                </p>
              </div>
            </template>
          </BaseCollapsible>
        </BaseCard>

        <!-- BaseAlert 警告 -->
        <BaseCard>
          <BaseCollapsible v-model:open="collapsibleStates.alert">
            <div class="flex items-center gap-2 cursor-pointer p-4">
              <BaseIcon name="i-lucide-alert-circle" class="w-5 h-5" />
              <h2 class="text-2xl font-bold">BaseAlert 警告</h2>
              <BaseIcon 
                :name="collapsibleStates.alert ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" 
                class="w-5 h-5 ml-auto"
              />
            </div>
            <template #content>
              <div class="px-4 pb-4 space-y-4">
                  <div class="flex flex-wrap gap-3">
                    <BaseButton @click="showAlert('success')" color="success">
                      顯示成功 Alert
                    </BaseButton>
                    <BaseButton @click="showAlert('warning')" color="warning">
                      顯示警告 Alert
                    </BaseButton>
                    <BaseButton @click="showAlert('error')" color="error">
                      顯示錯誤 Alert
                    </BaseButton>
                    <BaseButton @click="showAlert('info')" color="neutral">
                      顯示資訊 Alert
                    </BaseButton>
                  </div>

                  <div class="space-y-3 max-w-2xl">
                    <BaseAlert v-if="alertStates.success" color="success" icon="i-lucide-check-circle" :close="{ onClick: () => alertStates.success = false }">
                      <template #title>成功訊息</template>
                      <template #description>操作已成功完成</template>
                    </BaseAlert>
                    <BaseAlert v-if="alertStates.warning" color="warning" icon="i-lucide-alert-triangle" :close="{ onClick: () => alertStates.warning = false }">
                      <template #title>警告訊息</template>
                      <template #description>請注意這個警告</template>
                    </BaseAlert>
                    <BaseAlert v-if="alertStates.error" color="error" icon="i-lucide-x-circle" :close="{ onClick: () => alertStates.error = false }">
                      <template #title>錯誤訊息</template>
                      <template #description>發生了一個錯誤</template>
                    </BaseAlert>
                    <BaseAlert v-if="alertStates.info" color="info" icon="i-lucide-info" :close="{ onClick: () => alertStates.info = false }">
                      <template #title>資訊訊息</template>
                      <template #description>這是一條資訊訊息</template>
                    </BaseAlert>
                  </div>
              </div>
            </template>
          </BaseCollapsible>
        </BaseCard>

        <!-- BaseCollapsible 摺疊 -->
        <BaseCard>
          <BaseCollapsible v-model:open="collapsibleStates.collapsible">
            <div class="flex items-center gap-2 cursor-pointer p-4">
              <BaseIcon name="i-lucide-chevron-down" class="w-5 h-5" />
              <h2 class="text-2xl font-bold">BaseCollapsible 摺疊</h2>
              <BaseIcon 
                :name="collapsibleStates.collapsible ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" 
                class="w-5 h-5 ml-auto"
              />
            </div>
            <template #content>
              <div class="px-4 pb-4 max-w-2xl">
                  <BaseCollapsible v-model:open="collapsibleStates.collapsibleInner">
                    <div class="flex items-center justify-between p-4 bg-gray-100 dark:bg-gray-800 rounded-lg cursor-pointer">
                      <span class="font-semibold">可摺疊內容</span>
                      <BaseIcon 
                        :name="collapsibleStates.collapsibleInner ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" 
                        class="w-5 h-5" 
                      />
                    </div>
                    <template #content>
                      <div class="p-4 bg-gray-50 dark:bg-gray-900 rounded-b-lg">
                        <p>這是摺疊內容的詳細資訊。可以在這裡放置任何內容。</p>
                      </div>
                    </template>
                  </BaseCollapsible>
              </div>
            </template>
          </BaseCollapsible>
        </BaseCard>

        <!-- Toast 通知測試 -->
        <BaseCard>
          <BaseCollapsible v-model:open="collapsibleStates.toast">
            <div class="flex items-center gap-2 cursor-pointer p-4">
              <BaseIcon name="i-lucide-bell" class="w-5 h-5" />
              <h2 class="text-2xl font-bold">Toast 通知</h2>
              <BaseIcon 
                :name="collapsibleStates.toast ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" 
                class="w-5 h-5 ml-auto"
              />
            </div>
            <template #content>
              <div class="px-4 pb-4 flex flex-wrap gap-3">
                  <BaseButton @click="showToast('success', '成功操作')" color="success">
                    成功 Toast
                  </BaseButton>
                  <BaseButton @click="showToast('error', '發生錯誤')" color="error">
                    錯誤 Toast
                  </BaseButton>
                  <BaseButton @click="showToast('warning', '警告訊息')" color="warning">
                    警告 Toast
                  </BaseButton>
                  <BaseButton @click="showToast('info', '資訊訊息')" color="neutral">
                    資訊 Toast
                  </BaseButton>
              </div>
            </template>
          </BaseCollapsible>
        </BaseCard>

      </div>
    </main>

  </div>

  <!-- Modal 測試 -->
  <BaseModal v-model="isModalOpen" title="Modal 示範">
    <template #body>
      <div class="space-y-4">
        <p class="text-gray-600 dark:text-gray-400">
          這是 BaseModal 組件的示範。Modal 可以用來顯示重要的對話框或表單。
        </p>
        <div class="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-lg">
          <p class="text-sm text-blue-800 dark:text-blue-200 font-semibold mb-2">Modal 特性：</p>
          <ul class="text-sm text-blue-800 dark:text-blue-200 space-y-1">
            <li>• 支持多種尺寸 (sm, md, lg, xl, full)</li>
            <li>• 支持自定義標題和描述</li>
            <li>• 支持 body 和 footer slots</li>
            <li>• 自動處理關閉和背景遮罩</li>
          </ul>
        </div>
      </div>
    </template>
    <template #footer>
      <div class="flex justify-end gap-3">
        <BaseButton color="neutral" variant="outline" @click="closeModal">
          取消
        </BaseButton>
        <BaseButton @click="closeModal">
          確認
        </BaseButton>
      </div>
    </template>
  </BaseModal>

  <!-- Slideover 測試 -->
  <BaseSlideover v-model="isSlideoverOpen" title="Slideover 示範" side="right" size="lg">
    <template #body>
      <div class="space-y-4">
        <p class="text-gray-600 dark:text-gray-400">
          這是 Slideover 組件的示範。Slideover 從側邊滑入，適合顯示輔助資訊或表單。
        </p>
        <BaseCard>
          <p class="text-sm">可以在 Slideover 中放置任何內容</p>
        </BaseCard>
      </div>
    </template>
    <template #footer>
      <div class="flex justify-end gap-3">
        <BaseButton color="neutral" variant="outline" @click="closeSlideover">
          關閉
        </BaseButton>
      </div>
    </template>
  </BaseSlideover>
</template>

<style scoped>
.drag-handle {
  cursor: grab;
}

.drag-handle:active {
  cursor: grabbing;
}

/* 看板卡片拖曳效果 */
.board-card-item {
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  transform-origin: center;
  will-change: transform, box-shadow;
}

.board-card-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px -2px rgba(0, 0, 0, 0.1),
              0 2px 6px -2px rgba(0, 0, 0, 0.05);
}
</style>
