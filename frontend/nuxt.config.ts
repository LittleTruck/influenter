// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-04-03',
  devtools: { enabled: true },
  
  // Modules
  modules: [
    '@pinia/nuxt',
  ],

  // App config
  app: {
    head: {
      title: 'Influenter - 網紅案件管理系統',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'AI 驅動的網紅案件管理系統' },
      ],
      link: [
        { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' }
      ]
    }
  },

  // Runtime config
  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080',
    }
  },

  // CSS
  css: [],

  // TypeScript
  typescript: {
    strict: true,
    typeCheck: false  // 關閉 vue-tsc 類型檢查以避免開發環境中的錯誤
  },

  // Nitro
  nitro: {
    compressPublicAssets: true,
  },

  // Dev server
  devServer: {
    port: 3000,
    host: '0.0.0.0'
  }
})


