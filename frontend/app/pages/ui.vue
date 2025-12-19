<script setup lang="ts">
/**
 * UI 元件展示頁面
 * 
 * 展示所有 Nuxt UI 和自定義 Base 元件的功能
 * 用於開發時的 UI 參考和 debug
 */

definePageMeta({
  layout: false, // 不使用任何 layout
})

useSeoMeta({
  title: 'UI 元件展示',
  description: 'CAPSULE-CRM UI 元件展示頁面',
})

// Modal 和 Slideover 狀態
const isModalOpen = ref(false)
const isSlideoverOpen = ref(false)

// Toast 通知
const toast = useToast()

// 測試數據
const testData = ref([
  { id: 1, name: '測試項目 1', status: 'active', email: 'test1@example.com' },
  { id: 2, name: '測試項目 2', status: 'inactive', email: 'test2@example.com' },
  { id: 3, name: '測試項目 3', status: 'pending', email: 'test3@example.com' }
])

const selectedItem = ref(null)

// 測試函數
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

const selectItem = (item: any) => {
  selectedItem.value = item
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    <!-- 導航列 -->
    <nav class="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex items-center">
            <h1 class="text-xl font-semibold text-gray-900 dark:text-white">
              UI 元件展示
            </h1>
          </div>
          <div class="flex items-center space-x-4">
            <UButton @click="openModal" color="primary">
              測試 Modal
            </UButton>
            <UButton @click="openSlideover" color="secondary">
              測試 Slideover
            </UButton>
          </div>
        </div>
      </div>
    </nav>

    <!-- 主要內容 -->
    <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
      <div class="px-4 py-6 sm:px-0">
        <div class="space-y-8">
          
          <!-- 1. 按鈕測試 -->
          <div class="bg-white dark:bg-gray-800 rounded-lg p-6 shadow">
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">按鈕元件</h2>
            <div class="space-y-4">
              <div>
                <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Nuxt UI 按鈕</h3>
                <div class="flex flex-wrap gap-4">
                  <UButton>預設</UButton>
                  <UButton color="primary">主要</UButton>
                  <UButton color="secondary">次要</UButton>
                  <UButton color="success">成功</UButton>
                  <UButton color="warning">警告</UButton>
                  <UButton color="error">錯誤</UButton>
                  <UButton color="neutral">中性</UButton>
                </div>
                <div class="flex flex-wrap gap-4 mt-2">
                  <UButton variant="outline">外框</UButton>
                  <UButton variant="ghost">幽靈</UButton>
                  <UButton variant="soft">柔和</UButton>
                </div>
              </div>
              <div>
                <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">帶圖示按鈕</h3>
                <div class="flex flex-wrap gap-4">
                  <UButton icon="i-heroicons-plus">帶圖示</UButton>
                  <UButton icon="i-heroicons-heart" color="primary">愛心</UButton>
                  <UButton icon="i-heroicons-star" color="warning">星星</UButton>
                  <UButton icon="i-heroicons-check" color="success">確認</UButton>
                </div>
              </div>
            </div>
          </div>

          <!-- 2. 輸入框測試 -->
          <div class="bg-white dark:bg-gray-800 rounded-lg p-6 shadow">
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">輸入框元件</h2>
            <div class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">UInput 變體</label>
                <div class="space-y-2">
                  <UInput placeholder="預設輸入框" />
                  <UInput placeholder="錯誤狀態" color="error" />
                  <UInput placeholder="成功狀態" color="success" />
                  <UInput placeholder="帶圖示" icon="i-heroicons-magnifying-glass" />
                </div>
              </div>
            </div>
          </div>

          <!-- 3. 選擇器測試 -->
          <div class="bg-white dark:bg-gray-800 rounded-lg p-6 shadow">
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">選擇器元件</h2>
            <div class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">USelect 變體</label>
                <div class="space-y-2">
                  <USelect 
                    :items="[
                      { label: '選項 1', value: '1' },
                      { label: '選項 2', value: '2' },
                      { label: '選項 3', value: '3' }
                    ]"
                    placeholder="預設選擇器"
                  />
                  <USelect 
                    :items="[
                      { label: '錯誤選項 1', value: '1' },
                      { label: '錯誤選項 2', value: '2' }
                    ]"
                    placeholder="錯誤狀態"
                    color="error"
                  />
                </div>
              </div>
            </div>
          </div>
                    <!-- 4. 卡片測試 -->
                    <div class="bg-white dark:bg-gray-800 rounded-lg p-6 shadow">
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">卡片元件</h2>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UCard>
                <template #header>
                  <h3 class="text-lg font-semibold">UCard 標題</h3>
                </template>
                <p class="text-gray-600 dark:text-gray-400">這是 UCard 的內容</p>
                <template #footer>
                  <UButton size="sm">操作</UButton>
                </template>
              </UCard>
              
              <UCard>
                <template #header>
                  <h3 class="text-lg font-semibold">UCard 變體</h3>
                </template>
                <p class="text-gray-600 dark:text-gray-400">這是另一個 UCard 的內容</p>
                <template #footer>
                  <div class="flex gap-2">
                    <UButton size="sm" variant="outline">取消</UButton>
                    <UButton size="sm" color="primary">確認</UButton>
                  </div>
                </template>
              </UCard>
            </div>
          </div>

          <!-- 5. 表格測試 -->
          <div class="bg-white dark:bg-gray-800 rounded-lg p-6 shadow">
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">表格元件</h2>
            <div class="space-y-4">
              <div>
                <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">UTable</h3>
                <UTable :rows="testData" />
              </div>
            </div>
          </div>

          <!-- 6. 徽章測試 -->
          <div class="bg-white dark:bg-gray-800 rounded-lg p-6 shadow">
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">徽章元件</h2>
            <div class="flex flex-wrap gap-4">
              <UBadge>預設徽章</UBadge>
              <UBadge color="primary">主要徽章</UBadge>
              <UBadge color="success">成功徽章</UBadge>
              <UBadge color="warning">警告徽章</UBadge>
              <UBadge color="error">錯誤徽章</UBadge>
              <UBadge variant="outline">外框徽章</UBadge>
              <UBadge variant="soft">柔和徽章</UBadge>
            </div>
          </div>

          <!-- 7. 頭像測試 -->
          <div class="bg-white dark:bg-gray-800 rounded-lg p-6 shadow">
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">頭像元件</h2>
            <div class="flex flex-wrap gap-4 items-center">
              <UAvatar src="https://avatars.githubusercontent.com/u/739984?v=4" alt="Avatar" />
              <UAvatar alt="Initials">AB</UAvatar>
              <UAvatar alt="Icon" icon="i-heroicons-user" />
              <UAvatar alt="Fallback" src="invalid-url" />
            </div>
          </div>

          <!-- 8. Toast 測試 -->
          <div class="bg-white dark:bg-gray-800 rounded-lg p-6 shadow">
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">Toast 通知</h2>
            <div class="flex flex-wrap gap-4">
              <UButton @click="toast.add({ title: '成功訊息', color: 'success' })">成功 Toast</UButton>
              <UButton @click="toast.add({ title: '警告訊息', color: 'warning' })">警告 Toast</UButton>
              <UButton @click="toast.add({ title: '錯誤訊息', color: 'error' })">錯誤 Toast</UButton>
              <UButton @click="toast.add({ title: '資訊訊息', color: 'info' })">資訊 Toast</UButton>
            </div>
          </div>

        </div>
      </div>
    </main>
  </div>

  <!-- Modal 測試 -->
  <UModal v-model:open="isModalOpen" title="Modal 測試">
    <template #body>
      <div class="space-y-4">
        <p class="text-sm text-gray-600 dark:text-gray-400">
          這是使用正確 Nuxt UI v4 語法的 Modal。
        </p>
        <div class="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-lg">
          <p class="text-sm text-blue-800 dark:text-blue-200">
            <strong>語法重點：</strong>
          </p>
          <ul class="text-sm text-blue-800 dark:text-blue-200 mt-2 space-y-1">
            <li>• 使用 v-model:open 而不是 v-model</li>
            <li>• 使用 #body 和 #footer slots</li>
            <li>• 不需要 UCard 包裝</li>
            <li>• title 直接作為 prop</li>
          </ul>
        </div>
      </div>
    </template>

    <template #footer>
      <div class="flex justify-end gap-3">
        <UButton color="neutral" variant="outline" @click="closeModal">
          關閉
        </UButton>
        <UButton @click="closeModal">
          確認
        </UButton>
      </div>
    </template>
  </UModal>

  <!-- Slideover 測試 -->
  <USlideover v-model:open="isSlideoverOpen" title="Slideover 測試">
    <template #body>
      <div class="space-y-4">
        <p class="text-sm text-gray-600 dark:text-gray-400">
          這是使用正確 Nuxt UI v4 語法的 Slideover。
        </p>
        <div class="bg-green-50 dark:bg-green-900/20 p-4 rounded-lg">
          <p class="text-sm text-green-800 dark:text-green-200">
            <strong>語法重點：</strong>
          </p>
          <ul class="text-sm text-green-800 dark:text-green-200 mt-2 space-y-1">
            <li>• 使用 v-model:open 而不是 v-model</li>
            <li>• 使用 #body 和 #footer slots</li>
            <li>• 不需要 UCard 包裝</li>
            <li>• title 直接作為 prop</li>
          </ul>
        </div>
      </div>
    </template>

    <template #footer>
      <div class="flex justify-end gap-3">
        <UButton color="neutral" variant="outline" @click="closeSlideover">
          關閉
        </UButton>
        <UButton @click="closeSlideover">
          確認
        </UButton>
      </div>
    </template>
  </USlideover>
</template>