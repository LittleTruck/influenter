// Guest middleware - 已登入的使用者訪問登入頁面會被導向首頁

export default defineNuxtRouteMiddleware((to, from) => {
  // 在客戶端檢查認證狀態
  if (import.meta.client) {
    const authStore = useAuthStore()
    const token = authStore.token || localStorage.getItem('auth_token')
    
    // 如果已經登入（有 token），導向首頁
    if (token) {
      return navigateTo('/')
    }
  }
  
  // 在服務器端，允許通過
})

