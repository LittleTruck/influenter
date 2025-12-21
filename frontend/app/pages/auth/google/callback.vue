<script setup lang="ts">
import { BaseCard, BaseIcon } from '~/components/base'

definePageMeta({
  middleware: [],
  layout: false
})

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const toast = useToast()

const loading = ref(true)
const error = ref<string | null>(null)

onMounted(async () => {
  try {
    // 取得 authorization code
    const code = route.query.code as string
    const errorParam = route.query.error as string
    
    if (errorParam) {
      throw new Error(errorParam === 'access_denied' ? '用戶取消授權' : '授權失敗')
    }
    
    if (!code) {
      throw new Error('未收到授權碼')
    }
    
    // 呼叫後端 callback API
    const config = useRuntimeConfig()
    const response = await $fetch<{
      user: any
      token: string
    }>(`${config.public.apiBase}/api/v1/auth/google/callback`, {
      method: 'POST',
      body: {
        code,
        redirect_uri: `${window.location.origin}/auth/google/callback`
      }
    })
    
    // 儲存 token 和用戶資訊
    authStore.setToken(response.token)
    authStore.setUser(response.user)
    
    // 顯示成功訊息
    toast.add({
      title: '登入成功',
      description: `歡迎回來，${response.user.name}！`,
      color: 'success'
    })
    
    // 取得返回 URL
    const returnUrl = localStorage.getItem('oauth_return_url') || '/'
    localStorage.removeItem('oauth_return_url')
    
    // 導航到目標頁面
    await router.push(returnUrl)
  } catch (e: any) {
    console.error('OAuth callback error:', e)
    error.value = e.message || '登入失敗'
    
    toast.add({
      title: '登入失敗',
      description: error.value,
      color: 'error'
    })
    
    // 3 秒後返回登入頁
    setTimeout(() => {
      router.push('/auth/login')
    }, 3000)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-elevated">
    <BaseCard class="max-w-md w-full">
      <div class="text-center py-8">
        <!-- Loading -->
        <div v-if="loading" class="space-y-4">
          <BaseIcon 
            name="i-lucide-loader-2" 
            class="w-12 h-12 mx-auto text-primary animate-spin"
          />
          <h2 class="text-lg font-semibold text-highlighted">正在處理登入...</h2>
          <p class="text-sm text-muted">請稍候</p>
        </div>
        
        <!-- Error -->
        <div v-else-if="error" class="space-y-4">
          <BaseIcon 
            name="i-lucide-x-circle" 
            class="w-12 h-12 mx-auto text-error"
          />
          <h2 class="text-lg font-semibold text-error">登入失敗</h2>
          <p class="text-sm text-muted">{{ error }}</p>
          <p class="text-xs text-muted">將在 3 秒後返回登入頁面...</p>
        </div>
        
        <!-- Success -->
        <div v-else class="space-y-4">
          <BaseIcon 
            name="i-lucide-check-circle" 
            class="w-12 h-12 mx-auto text-success"
          />
          <h2 class="text-lg font-semibold text-success">登入成功！</h2>
          <p class="text-sm text-muted">正在跳轉...</p>
        </div>
      </div>
    </BaseCard>
  </div>
</template>

