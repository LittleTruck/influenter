// Guest middleware - 已登入的使用者訪問登入頁面會被導向首頁

export default defineNuxtRouteMiddleware((to, from) => {
  const token = useCookie('auth_token')

  if (token.value) {
    return navigateTo('/')
  }
})

