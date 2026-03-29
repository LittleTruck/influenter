<script setup lang="ts">
import { BaseCard, BaseAlert, BaseIcon } from '~/components/base'

definePageMeta({
  layout: false,
  middleware: ['guest'], // 已登入的使用者會被導向首頁
})

useSeoMeta({
  title: '登入 - Influenter',
  description: '使用 Google 帳號登入 Influenter 網紅案件管理系統'
})

const config = useRuntimeConfig()
const googleClientId = config.public.googleClientId
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
    <BaseCard class="max-w-md w-full mx-4">
      <template #header>
        <div class="text-center">
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
            Influenter
          </h1>
          <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
            AI 驅動的網紅案件管理系統
          </p>
        </div>
      </template>

      <div class="space-y-6">
        <!-- 登入說明 -->
        <div class="text-center">
          <p class="text-gray-700 dark:text-gray-300">
            使用 Google 帳號快速登入
          </p>
        </div>

        <!-- Google 登入按鈕 -->
        <div class="flex flex-col items-center space-y-4">
          <div v-if="googleClientId" class="w-full max-w-md">
            <GoogleLoginButton />
          </div>
          <BaseAlert
            v-else
            color="warning"
            variant="soft"
            title="Google Client ID 未設定"
            description="請在環境變數中設定 NUXT_PUBLIC_GOOGLE_CLIENT_ID"
          />
        </div>

        <!-- 功能特色 -->
        <div class="pt-6 border-t border-gray-200 dark:border-gray-700">
          <div class="space-y-3">
            <div class="flex items-start gap-3">
              <BaseIcon name="i-lucide-mail" class="w-5 h-5 text-primary mt-0.5" />
              <div>
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">智能郵件分析</h3>
                <p class="text-xs text-gray-600 dark:text-gray-400">自動分析合作邀約郵件</p>
              </div>
            </div>
            <div class="flex items-start gap-3">
              <BaseIcon name="i-lucide-briefcase" class="w-5 h-5 text-primary mt-0.5" />
              <div>
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">案件追蹤管理</h3>
                <p class="text-xs text-gray-600 dark:text-gray-400">輕鬆管理所有合作案件</p>
              </div>
            </div>
            <div class="flex items-start gap-3">
              <BaseIcon name="i-lucide-sparkles" class="w-5 h-5 text-primary mt-0.5" />
              <div>
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">AI 回覆建議</h3>
                <p class="text-xs text-gray-600 dark:text-gray-400">自動生成專業回覆內容</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </BaseCard>
  </div>
</template>

