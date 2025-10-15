// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',
  
  future: {
    compatibilityVersion: 4
  },

  devtools: { enabled: true },
  
  // Modules
  modules: [
    '@nuxt/ui',
    '@nuxt/fonts',
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
  css: ['~/assets/css/main.css'],

  // TypeScript
  typescript: {
    strict: true,
    typeCheck: false
  },

  // Nuxt UI Configuration
  ui: {
    colorMode: {
      preference: 'light'
    }
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


