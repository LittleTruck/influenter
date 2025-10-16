<template>
  <UDashboardPanel>
    <template #header>
      <UDashboardNavbar title="歡迎使用 Influenter">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>

        <template #trailing>
          <UColorModeButton />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
        <!-- Welcome Section -->
        <div class="mb-8">
          <div class="flex items-center gap-3 mb-4">
            <UIcon name="i-lucide-sparkles" class="w-10 h-10 text-primary" />
            <div>
              <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
                {{ user ? `歡迎回來，${user.name || user.email}` : '歡迎來到 Influenter' }}
              </h1>
              <p class="text-gray-500 dark:text-gray-400 mt-1">
                AI 驅動的網紅案件管理系統
              </p>
            </div>
          </div>
        </div>

        <!-- Stats Cards -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
          <UCard>
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-gray-500 dark:text-gray-400">總案件數</p>
                <p class="text-2xl font-bold text-gray-900 dark:text-white">0</p>
              </div>
              <UIcon name="i-lucide-briefcase" class="w-8 h-8 text-blue-500" />
            </div>
          </UCard>

          <UCard>
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-gray-500 dark:text-gray-400">未讀郵件</p>
                <p class="text-2xl font-bold text-gray-900 dark:text-white">0</p>
              </div>
              <UIcon name="i-lucide-mail" class="w-8 h-8 text-green-500" />
            </div>
          </UCard>

          <UCard>
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-gray-500 dark:text-gray-400">本月收入</p>
                <p class="text-2xl font-bold text-gray-900 dark:text-white">NT$ 0</p>
              </div>
              <UIcon name="i-lucide-dollar-sign" class="w-8 h-8 text-yellow-500" />
            </div>
          </UCard>
        </div>

        <!-- Getting Started -->
        <UCard>
          <template #header>
            <h2 class="text-xl font-semibold">快速開始</h2>
          </template>

          <div class="space-y-4">
            <UAlert
              icon="i-lucide-lightbulb"
              color="primary"
              variant="subtle"
              title="系統初始化成功！"
              description="您的 Influenter 系統已經準備就緒。接下來可以開始設定您的 Gmail 帳號和 OpenAI API。"
            />

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UButton
                icon="i-lucide-mail"
                label="連接 Gmail"
                color="primary"
                variant="soft"
                size="lg"
                block
                to="/settings/gmail"
              />

              <UButton
                icon="i-lucide-sparkles"
                label="設定 AI"
                color="primary"
                variant="soft"
                size="lg"
                block
                to="/settings/ai"
              />
            </div>
          </div>
        </UCard>

        <!-- Features Preview -->
        <div class="mt-8 grid grid-cols-1 md:grid-cols-2 gap-6">
          <UCard>
            <template #header>
              <div class="flex items-center gap-2">
                <UIcon name="i-lucide-mail" class="w-5 h-5 text-primary" />
                <h3 class="font-semibold">智能郵件分析</h3>
              </div>
            </template>
            <p class="text-sm text-gray-600 dark:text-gray-300">
              AI 自動分析來信，識別合作邀約並提取關鍵資訊
            </p>
          </UCard>

          <UCard>
            <template #header>
              <div class="flex items-center gap-2">
                <UIcon name="i-lucide-briefcase" class="w-5 h-5 text-primary" />
                <h3 class="font-semibold">案件管理</h3>
              </div>
            </template>
            <p class="text-sm text-gray-600 dark:text-gray-300">
              追蹤每個合作案件的進度、截止日期和任務清單
            </p>
          </UCard>

          <UCard>
            <template #header>
              <div class="flex items-center gap-2">
                <UIcon name="i-lucide-calendar" class="w-5 h-5 text-primary" />
                <h3 class="font-semibold">日曆視圖</h3>
              </div>
            </template>
            <p class="text-sm text-gray-600 dark:text-gray-300">
              清楚掌握所有案件的時程安排和重要日期
            </p>
          </UCard>

          <UCard>
            <template #header>
              <div class="flex items-center gap-2">
                <UIcon name="i-lucide-message-square" class="w-5 h-5 text-primary" />
                <h3 class="font-semibold">AI 回覆生成</h3>
              </div>
            </template>
            <p class="text-sm text-gray-600 dark:text-gray-300">
              根據案件狀態和您的偏好，自動生成專業的回覆郵件
            </p>
          </UCard>
        </div>
      </template>
  </UDashboardPanel>
</template>

<script setup lang="ts">
// 首頁 - Dashboard

definePageMeta({
  middleware: ['auth'], // 需要登入才能存取
})

const authStore = useAuthStore()
const user = computed(() => authStore.user)
</script>


