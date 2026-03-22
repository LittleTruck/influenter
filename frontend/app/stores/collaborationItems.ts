import { defineStore } from 'pinia'
import { logError } from '~/utils/errorUtils'
import { collaborationItemsStorage, generateTempId } from '~/utils/localStorage'
import type {
  CollaborationItem,
  CreateCollaborationItemRequest,
  UpdateCollaborationItemRequest,
  ReorderItemsRequest,
  CollaborationItemListResponse
} from '~/types/collaborationItems'

export const useCollaborationItemsStore = defineStore('collaborationItems', () => {
  // State
  const items: Ref<CollaborationItem[]> = ref([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  /**
   * 單項列表
   */
  const individualItems = computed(() =>
    items.value.filter(item => item.type === 'individual')
  )

  /**
   * 組合列表
   */
  const bundleItems = computed(() =>
    items.value.filter(item => item.type === 'bundle')
  )

  /**
   * 取得所有項目（扁平列表）
   */
  const fetchItems = async () => {
    // 如果已經在載入，直接返回（避免重複請求）
    if (loading.value) {
      console.debug('fetchItems called while already loading, skipping')
      return
    }

    loading.value = true
    error.value = null

    // 設置超時保護：確保 loading 不會永遠為 true
    const timeoutId = setTimeout(() => {
      if (loading.value) {
        console.warn('fetchItems timeout after 10s, forcing loading to false')
        loading.value = false
      }
    }, 10000)

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const data = await $fetch<CollaborationItemListResponse>(
        `${config.public.apiBase}/api/v1/collaboration-items`,
        {
          method: 'GET',
          headers: {
            Authorization: `Bearer ${authStore.token}`
          },
          // 設置請求超時（5秒）
          timeout: 5000
        }
      )

      // 後端已回傳扁平結構，直接使用
      items.value = data.data.sort((a, b) => a.order - b.order)

      // 同步到 localStorage
      collaborationItemsStorage.setItems(data.data)
    } catch (e: any) {
      // 如果是 404 或網絡錯誤，這是正常的（後端還沒實作），不記錄錯誤
      const is404 = e?.statusCode === 404 || e?.status === 404 || e?.response?.status === 404
      const isNetworkError = e?.name === 'FetchError' || e?.message?.includes('fetch')

      if (!is404 && !isNetworkError) {
        error.value = logError(e, '取得合作項目列表失敗', { component: 'collaborationItemsStore', action: 'fetchItems' })
      }

      // Fallback 到 localStorage
      try {
        const localItems = collaborationItemsStorage.getItems()
        if (localItems.length > 0) {
          items.value = localItems.sort((a: CollaborationItem, b: CollaborationItem) => a.order - b.order)
        } else {
          items.value = []
        }
      } catch (storageError) {
        console.error('Failed to load from localStorage:', storageError)
        items.value = []
      }
    } finally {
      clearTimeout(timeoutId)
      // 確保 loading 狀態一定會被設置為 false
      loading.value = false
    }
  }

  /**
   * 創建項目
   */
  const createItem = async (data: CreateCollaborationItemRequest) => {
    loading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const newItem = await $fetch<CollaborationItem>(
        `${config.public.apiBase}/api/v1/collaboration-items`,
        {
          method: 'POST',
          body: data,
          headers: {
            Authorization: `Bearer ${authStore.token}`
          }
        }
      )

      // 重新載入列表
      await fetchItems()

      return newItem
    } catch (e: unknown) {
      error.value = logError(e, '建立合作項目失敗（已儲存到本地）', { component: 'collaborationItemsStore', action: 'createItem' })
      // Fallback 到 localStorage
      const tempId = generateTempId()
      const now = new Date().toISOString()
      const maxOrder = items.value.length > 0
        ? Math.max(...items.value.map(item => item.order))
        : -1

      const newItem: CollaborationItem = {
        id: tempId,
        title: data.title,
        description: data.description,
        price: data.price,
        type: data.type || 'individual',
        order: maxOrder + 1,
        created_at: now,
        updated_at: now
      }

      items.value.push(newItem)
      collaborationItemsStorage.addItem(newItem)

      return newItem
    } finally {
      loading.value = false
    }
  }

  /**
   * 更新項目
   */
  const updateItem = async (id: string, data: UpdateCollaborationItemRequest) => {
    loading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const updatedItem = await $fetch<CollaborationItem>(
        `${config.public.apiBase}/api/v1/collaboration-items/${id}`,
        {
          method: 'PATCH',
          body: data,
          headers: {
            Authorization: `Bearer ${authStore.token}`
          }
        }
      )

      // 重新載入列表
      await fetchItems()

      return updatedItem
    } catch (e: unknown) {
      error.value = logError(e, '更新合作項目失敗（已儲存到本地）', { component: 'collaborationItemsStore', action: 'updateItem' })
      // Fallback 到 localStorage
      const index = items.value.findIndex(item => item.id === id)
      if (index !== -1) {
        items.value[index] = { ...items.value[index], ...data, updated_at: new Date().toISOString() }
        collaborationItemsStorage.updateItem(id, items.value[index])
      }
      return items.value.find(item => item.id === id)!
    } finally {
      loading.value = false
    }
  }

  /**
   * 刪除項目
   */
  const deleteItem = async (id: string) => {
    loading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      await $fetch(`${config.public.apiBase}/api/v1/collaboration-items/${id}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      // 重新載入列表
      await fetchItems()
    } catch (e: unknown) {
      error.value = logError(e, '刪除合作項目失敗（已從本地移除）', { component: 'collaborationItemsStore', action: 'deleteItem' })
      // Fallback: 直接從列表移除
      items.value = items.value.filter(item => item.id !== id)
      collaborationItemsStorage.setItems(items.value)
    } finally {
      loading.value = false
    }
  }

  /**
   * 重新排序項目
   */
  const reorderItems = async (itemIds: string[]) => {
    loading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const data: ReorderItemsRequest = {
        item_ids: itemIds
      }

      await $fetch(`${config.public.apiBase}/api/v1/collaboration-items/reorder`, {
        method: 'PATCH',
        body: data,
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      // 重新載入列表
      await fetchItems()
    } catch (e: unknown) {
      error.value = logError(e, '重新排序失敗（已儲存到本地）', { component: 'collaborationItemsStore', action: 'reorderItems' })
      // Fallback 到 localStorage
      itemIds.forEach((id, index) => {
        const item = items.value.find(i => i.id === id)
        if (item) {
          item.order = index
        }
      })
      items.value.sort((a, b) => a.order - b.order)
      collaborationItemsStorage.setItems(items.value)
    } finally {
      loading.value = false
    }
  }

  /**
   * 根據 ID 查找項目
   */
  const findItemById = (id: string): CollaborationItem | null => {
    return items.value.find(item => item.id === id) || null
  }

  // 重置狀態
  const reset = () => {
    items.value = []
    loading.value = false
    error.value = null
  }

  return {
    // State
    items,
    loading,
    error,

    // Computed
    individualItems,
    bundleItems,

    // Actions
    fetchItems,
    createItem,
    updateItem,
    deleteItem,
    reorderItems,
    findItemById,
    reset
  }
})
