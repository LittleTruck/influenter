import { logError } from '~/utils/errorUtils'
import type { CollaborationItemAnalyticsResponse } from '~/types/analytics'

/**
 * 數據分析 composable
 * 取得「依月份聚合、依合作項目堆疊」的金額與專案數
 */
export const useAnalytics = () => {
  const loading = ref(false)
  const error = ref<string | null>(null)

  const fetchCollaborationAnalytics = async (): Promise<CollaborationItemAnalyticsResponse> => {
    loading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const data = await $fetch<CollaborationItemAnalyticsResponse>(
        `${config.public.apiBase}/api/v1/analytics/collaboration-items`,
        {
          method: 'GET',
          headers: {
            Authorization: `Bearer ${authStore.token}`
          }
        }
      )

      return data
    } catch (e: unknown) {
      error.value = logError(e, '取得數據分析失敗', { component: 'useAnalytics', action: 'fetchCollaborationAnalytics' })
      return { series: [], months: [] }
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    fetchCollaborationAnalytics
  }
}
