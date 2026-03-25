// 認證 middleware - 保護需要登入的頁面
// 使用 cookie 讓 SSR 和 CSR 都能檢查認證狀態

export default defineNuxtRouteMiddleware((to, from) => {
  const token = useCookie('auth_token')

  // 在 client 端也檢查 localStorage（向下相容）
  if (import.meta.client && !token.value) {
    const localToken = localStorage.getItem('auth_token')
    if (localToken) {
      // 將 localStorage 的 token 同步到 cookie
      token.value = localToken
      return
    }
  }

  if (!token.value) {
    return navigateTo({
      path: '/auth/login',
      query: {
        redirect: to.fullPath,
      },
    })
  }
})

