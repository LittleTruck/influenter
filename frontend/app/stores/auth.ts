import { defineStore } from 'pinia'
import type { User, AuthState } from '~/types/auth'

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    user: null,
    token: null,
    isAuthenticated: false,
    loading: false,
  }),

  getters: {
    currentUser: (state) => state.user,
    isLoggedIn: (state) => state.isAuthenticated && !!state.token,
  },

  actions: {
    setUser(user: User | null) {
      this.user = user
      this.isAuthenticated = !!user
    },

    setToken(token: string | null) {
      this.token = token
      if (token) {
        // 儲存 token 到 localStorage
        if (import.meta.client) {
          localStorage.setItem('auth_token', token)
        }
      } else {
        // 清除 token
        if (import.meta.client) {
          localStorage.removeItem('auth_token')
        }
      }
    },

    setLoading(loading: boolean) {
      this.loading = loading
    },

    async initAuth() {
      // 從 localStorage 載入 token
      if (import.meta.client) {
        const token = localStorage.getItem('auth_token')
        if (token) {
          this.token = token
          // 這裡可以呼叫 API 驗證 token 並取得使用者資料
          try {
            const config = useRuntimeConfig()
            const user = await $fetch<User>(`${config.public.apiBase}/api/v1/auth/me`, {
              method: 'GET',
              headers: {
                Authorization: `Bearer ${token}`,
              },
            })
            this.setUser(user)
          } catch (error) {
            console.error('Failed to fetch current user:', error)
            this.logout()
          }
        }
      }
    },

    async login(user: User, token: string) {
      this.setUser(user)
      this.setToken(token)
    },

    logout() {
      this.setUser(null)
      this.setToken(null)
      // 導向登入頁面
      if (import.meta.client) {
        navigateTo('/auth/login')
      }
    },
  },

  persist: {
    storage: import.meta.client ? localStorage : undefined,
    paths: ['token'],
  },
})

