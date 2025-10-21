import { defineStore } from 'pinia'
import type { Ref } from 'vue'

// Email 類型定義
export interface Email {
  id: string
  from_email: string
  from_name?: string
  subject?: string
  snippet?: string
  received_at: string
  is_read: boolean
  has_attachments: boolean
  labels?: string[]
  case_id?: string
  ai_analyzed: boolean
}

export interface EmailDetail extends Email {
  oauth_account_id: string
  provider_message_id: string
  thread_id?: string
  to_email?: string
  body_text?: string
  body_html?: string
  ai_analysis_id?: string
  created_at: string
  updated_at: string
}

export interface EmailQueryParams {
  oauth_account_id?: string
  is_read?: boolean
  case_id?: string
  from_email?: string
  subject?: string
  start_date?: string
  end_date?: string
  page?: number
  page_size?: number
  sort_by?: string
  sort_order?: string
}

export interface EmailListResponse {
  emails: Email[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export interface GmailStatus {
  connected: boolean
  email?: string
  last_sync_at?: string
  sync_status?: string
  sync_error?: string
  token_expired?: boolean
  can_sync?: boolean
  stats?: {
    total_messages: number
    unread_messages: number
    starred_messages: number
    important_messages: number
    category_counts: Record<string, number>
  }
}

export const useEmailsStore = defineStore('emails', () => {
  // State
  const emails: Ref<Email[]> = ref([])
  const currentEmail: Ref<EmailDetail | null> = ref(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  
  // Pagination
  const pagination = ref({
    page: 1,
    page_size: 20,
    total: 0,
    total_pages: 0
  })

  // Filters
  const filters = ref<EmailQueryParams>({
    page: 1,
    page_size: 20,
    sort_by: 'received_at',
    sort_order: 'desc'
  })

  // Gmail status
  const gmailStatus: Ref<GmailStatus | null> = ref(null)
  const syncing = ref(false)

  // Actions
  const fetchEmails = async (params?: Partial<EmailQueryParams>) => {
    loading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()
      
      // 合併篩選參數，並過濾掉 undefined 值
      const mergedParams = { ...filters.value, ...params }
      
      // 移除 undefined 值，這樣可以清除之前的篩選
      const queryParams: any = {}
      Object.keys(mergedParams).forEach(key => {
        const value = mergedParams[key as keyof EmailQueryParams]
        if (value !== undefined) {
          queryParams[key] = value
        }
      })
      
      const data = await $fetch<EmailListResponse>(`${config.public.apiBase}/api/v1/emails`, {
        method: 'GET',
        params: queryParams,
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      emails.value = data.emails
      pagination.value = data.pagination
      filters.value = { ...queryParams }
    } catch (e: any) {
      error.value = e.message
      console.error('Failed to fetch emails:', e)
    } finally {
      loading.value = false
    }
  }

  const fetchEmail = async (id: string) => {
    loading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()
      
      const data = await $fetch<EmailDetail>(`${config.public.apiBase}/api/v1/emails/${id}`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      currentEmail.value = data
    } catch (e: any) {
      error.value = e.message
      console.error('Failed to fetch email:', e)
    } finally {
      loading.value = false
    }
  }

  const markAsRead = async (id: string, isRead = true) => {
    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()
      
      const data = await $fetch<EmailDetail>(`${config.public.apiBase}/api/v1/emails/${id}`, {
        method: 'PATCH',
        body: { is_read: isRead },
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      // 更新本地狀態
      const index = emails.value.findIndex(e => e.id === id)
      if (index !== -1) {
        emails.value[index].is_read = isRead
      }

      if (currentEmail.value?.id === id) {
        currentEmail.value.is_read = isRead
      }

      // 重新載入 Gmail status 以更新統計資料
      await fetchGmailStatus()

      return data
    } catch (e: any) {
      error.value = e.message
      console.error('Failed to mark email as read:', e)
      throw e
    }
  }

  const linkToCase = async (emailId: string, caseId: string) => {
    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()
      
      const data = await $fetch<EmailDetail>(`${config.public.apiBase}/api/v1/emails/${emailId}`, {
        method: 'PATCH',
        body: { case_id: caseId },
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      // 更新本地狀態
      const index = emails.value.findIndex(e => e.id === emailId)
      if (index !== -1) {
        emails.value[index].case_id = caseId
      }

      if (currentEmail.value?.id === emailId) {
        currentEmail.value.case_id = caseId
      }

      return data
    } catch (e: any) {
      error.value = e.message
      console.error('Failed to link email to case:', e)
      throw e
    }
  }

  const fetchGmailStatus = async () => {
    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()
      
      const data = await $fetch<GmailStatus>(`${config.public.apiBase}/api/v1/gmail/status`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      gmailStatus.value = data
    } catch (e: any) {
      error.value = e.message
      console.error('Failed to fetch Gmail status:', e)
      // 設置預設值，確保 gmailStatus 不會是 null
      gmailStatus.value = {
        connected: false,
        stats: {
          total_messages: 0,
          unread_messages: 0,
          starred_messages: 0,
          important_messages: 0,
          category_counts: {}
        }
      }
    }
  }

  const triggerSync = async () => {
    syncing.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()
      
      const data = await $fetch(`${config.public.apiBase}/api/v1/gmail/sync`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      // 維持同步中狀態直到後續資料刷新完成
      await new Promise(resolve => setTimeout(resolve, 2000))
      await fetchGmailStatus()
      await fetchEmails() // 自動刷新郵件列表

      return data
    } catch (e: any) {
      error.value = e.message
      console.error('Failed to trigger sync:', e)
      throw e
    } finally {
      syncing.value = false
    }
  }

  const disconnectGmail = async () => {
    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()
      
      await $fetch(`${config.public.apiBase}/api/v1/gmail/disconnect`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      // 清空本地資料
      emails.value = []
      currentEmail.value = null
      gmailStatus.value = null
    } catch (e: any) {
      error.value = e.message
      console.error('Failed to disconnect Gmail:', e)
      throw e
    }
  }

  // Getters
  const unreadCount = computed(() => {
    // 只使用 Gmail status 中的統計資料（全部未讀數量）
    // 後端使用 PascalCase: UnreadMessages
    if (gmailStatus.value?.stats?.UnreadMessages !== undefined) {
      return gmailStatus.value.stats.UnreadMessages
    }
    // 如果還沒有載入，返回 0
    return 0
  })

  const hasEmails = computed(() => {
    return emails.value.length > 0
  })

  const isConnected = computed(() => {
    return gmailStatus.value?.connected === true
  })

  const canSync = computed(() => {
    return gmailStatus.value?.can_sync === true && !syncing.value
  })

  // 重置狀態
  const reset = () => {
    emails.value = []
    currentEmail.value = null
    loading.value = false
    error.value = null
    pagination.value = {
      page: 1,
      page_size: 20,
      total: 0,
      total_pages: 0
    }
  }

  return {
    // State
    emails,
    currentEmail,
    loading,
    error,
    pagination,
    filters,
    gmailStatus,
    syncing,
    
    // Getters
    unreadCount,
    hasEmails,
    isConnected,
    canSync,
    
    // Actions
    fetchEmails,
    fetchEmail,
    markAsRead,
    linkToCase,
    fetchGmailStatus,
    triggerSync,
    disconnectGmail,
    reset
  }
})

