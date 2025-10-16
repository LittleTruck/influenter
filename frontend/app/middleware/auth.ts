// 認證 middleware - 保護需要登入的頁面

export default defineNuxtRouteMiddleware((to, from) => {
  // 在客戶端使用 localStorage 檢查 token
  if (import.meta.client) {
    const token = localStorage.getItem('auth_token')
    
    // 如果沒有 token，導向登入頁面
    if (!token) {
      return navigateTo({
        path: '/auth/login',
        query: {
          redirect: to.fullPath, // 記錄原本要去的頁面，登入後可以導回
        },
      })
    }
  }
  
  // 在服務器端，允許通過（因為 SSR 時無法檢查 localStorage）
  // 實際的認證會在客戶端 hydration 時處理
})

