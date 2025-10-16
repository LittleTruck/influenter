// Guest middleware - 已登入的使用者訪問登入頁面會被導向首頁

export default defineNuxtRouteMiddleware((to, from) => {
  // 在客戶端使用 localStorage 檢查 token
  if (import.meta.client) {
    const token = localStorage.getItem('auth_token')
    
    // 如果已經登入（有 token），導向首頁
    if (token) {
      return navigateTo('/')
    }
  }
  
  // 在服務器端，允許通過
})

