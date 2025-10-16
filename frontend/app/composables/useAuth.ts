import type { User, LoginResponse, GoogleAuthResponse, ApiError } from '~/types/auth'

export const useAuth = () => {
  const config = useRuntimeConfig()
  
  // 確保在正確的上下文中使用 store
  let authStore: ReturnType<typeof useAuthStore>
  try {
    authStore = useAuthStore()
  } catch (error) {
    console.error('Failed to initialize auth store:', error)
    throw error
  }

  /**
   * 使用 Google credential 登入
   */
  const loginWithGoogle = async (googleResponse: GoogleAuthResponse): Promise<LoginResponse> => {
    try {
      authStore.setLoading(true)

      const response = await $fetch<LoginResponse>(`${config.public.apiBase}/api/v1/auth/google`, {
        method: 'POST',
        body: {
          credential: googleResponse.credential,
          clientId: googleResponse.clientId,
        },
      })

      // 儲存使用者資料和 token
      await authStore.login(response.user, response.token)

      return response
    } catch (error: any) {
      console.error('Google login error:', error)
      throw {
        message: error?.data?.message || '登入失敗，請稍後再試',
        code: error?.statusCode,
        details: error?.data,
      } as ApiError
    } finally {
      authStore.setLoading(false)
    }
  }

  /**
   * 取得當前使用者資料
   */
  const fetchCurrentUser = async (): Promise<User> => {
    try {
      const token = authStore.token
      if (!token) {
        throw new Error('No token found')
      }

      const user = await $fetch<User>(`${config.public.apiBase}/api/v1/auth/me`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })

      authStore.setUser(user)
      return user
    } catch (error: any) {
      console.error('Fetch current user error:', error)
      throw {
        message: error?.data?.message || '無法取得使用者資料',
        code: error?.statusCode,
        details: error?.data,
      } as ApiError
    }
  }

  /**
   * 登出
   */
  const logout = async () => {
    try {
      const token = authStore.token
      if (token) {
        // 呼叫後端登出 API（選用）
        await $fetch(`${config.public.apiBase}/api/v1/auth/logout`, {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }).catch(() => {
          // 忽略錯誤，繼續清除本地狀態
        })
      }
    } finally {
      // 清除本地狀態
      authStore.logout()
    }
  }

  /**
   * 檢查是否已登入
   */
  const checkAuth = () => {
    return authStore.isLoggedIn
  }

  return {
    loginWithGoogle,
    fetchCurrentUser,
    logout,
    checkAuth,
    loading: computed(() => authStore.loading),
    user: computed(() => authStore.user),
    isAuthenticated: computed(() => authStore.isAuthenticated),
  }
}

