// Google Identity Services 初始化
export default defineNuxtPlugin((nuxtApp) => {
  const config = useRuntimeConfig()
  const clientId = config.public.googleClientId as string
  
  if (!clientId) {
    console.warn('Google Client ID not found. Please set NUXT_PUBLIC_GOOGLE_CLIENT_ID in your environment.')
    return
  }

  // 載入 Google Identity Services 腳本
  if (import.meta.client) {
    const script = document.createElement('script')
    script.src = 'https://accounts.google.com/gsi/client'
    script.async = true
    script.defer = true
    document.head.appendChild(script)
    
    // 提供 Google Client ID
    nuxtApp.provide('googleClientId', clientId)
  }
})

