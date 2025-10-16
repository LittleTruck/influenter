// 認證初始化 plugin (僅在客戶端執行)

export default defineNuxtPlugin(async () => {
  // 此 plugin 已經標記為 .client.ts，所以只會在客戶端執行
  const authStore = useAuthStore()

  try {
    await authStore.initAuth()
  } catch (error) {
    console.error('Auth initialization failed:', error)
  }
})

