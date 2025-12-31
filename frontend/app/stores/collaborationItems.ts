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
   * 將扁平數據轉換為樹狀結構
   */
  const buildTree = (flatItems: CollaborationItem[]): CollaborationItem[] => {
    // 創建一個映射表，方便查找
    const itemMap = new Map<string, CollaborationItem>()
    const rootItems: CollaborationItem[] = []

    // 先創建所有項目的副本，並初始化 children 陣列
    flatItems.forEach(item => {
      itemMap.set(item.id, { ...item, children: [] })
    })

    // 構建樹狀結構
    flatItems.forEach(item => {
      const node = itemMap.get(item.id)!
      if (!item.parent_id) {
        // 頂層項目
        rootItems.push(node)
      } else {
        // 子項目，添加到父項目的 children 中
        const parent = itemMap.get(item.parent_id)
        if (parent) {
          if (!parent.children) {
            parent.children = []
          }
          parent.children.push(node)
        }
      }
    })

    // 對每個層級進行排序
    const sortItems = (items: CollaborationItem[]) => {
      items.sort((a, b) => a.order - b.order)
      items.forEach(item => {
        if (item.children && item.children.length > 0) {
          sortItems(item.children)
        }
      })
    }

    sortItems(rootItems)
    return rootItems
  }

  /**
   * 將樹狀結構轉換為扁平數據
   */
  const flattenTree = (tree: CollaborationItem[], parentId: string | null = null): CollaborationItem[] => {
    const result: CollaborationItem[] = []
    tree.forEach((item, index) => {
      const flatItem: CollaborationItem = {
        ...item,
        parent_id: parentId,
        order: index,
        children: undefined // 移除 children 欄位
      }
      result.push(flatItem)
      if (item.children && item.children.length > 0) {
        result.push(...flattenTree(item.children, item.id))
      }
    })
    return result
  }

  /**
   * 取得所有項目（樹狀結構）
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

      // 後端返回扁平結構，轉換為樹狀
      items.value = buildTree(data.data)
      
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
          items.value = buildTree(localItems)
        } else {
          // 確保 items 為空陣列
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

      // 計算新項目的 order（同一層級內的最後一個）
      const siblings = items.value.filter(item => 
        item.parent_id === (data.parent_id || null)
      )
      const maxOrder = siblings.length > 0 
        ? Math.max(...siblings.map(item => item.order))
        : -1

      const newItem = await $fetch<CollaborationItem>(
        `${config.public.apiBase}/api/v1/collaboration-items`,
        {
          method: 'POST',
          body: {
            ...data,
            order: maxOrder + 1
          },
          headers: {
            Authorization: `Bearer ${authStore.token}`
          }
        }
      )

      // 重新載入列表以獲取樹狀結構
      await fetchItems()

      return newItem
    } catch (e: unknown) {
      error.value = logError(e, '建立合作項目失敗（已儲存到本地）', { component: 'collaborationItemsStore', action: 'createItem' })
      // Fallback 到 localStorage
      const tempId = generateTempId()
      const now = new Date().toISOString()
      const siblings = items.value.filter(item => 
        item.parent_id === (data.parent_id || null)
      )
      const maxOrder = siblings.length > 0 
        ? Math.max(...siblings.map(item => item.order))
        : -1

      const newItem: CollaborationItem = {
        id: tempId,
        ...data,
        parent_id: data.parent_id || null,
        order: maxOrder + 1,
        created_at: now,
        updated_at: now
      }

      // 添加到列表
      const flatItems = flattenTree(items.value)
      flatItems.push(newItem)
      items.value = buildTree(flatItems)
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
      const flatItems = flattenTree(items.value)
      const index = flatItems.findIndex(item => item.id === id)
      if (index !== -1) {
        flatItems[index] = { ...flatItems[index], ...data, updated_at: new Date().toISOString() }
        items.value = buildTree(flatItems)
        collaborationItemsStorage.updateItem(id, flatItems[index])
      }
      return flatItems.find(item => item.id === id)!
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
      // Fallback 到 localStorage：遞迴刪除項目及其子項目
      const flatItems = flattenTree(items.value)
      const itemToDelete = flatItems.find(item => item.id === id)
      if (itemToDelete) {
        // 找出所有子項目
        const getAllChildren = (parentId: string): string[] => {
          const children = flatItems.filter(item => item.parent_id === parentId)
          const childIds = children.map(child => child.id)
          children.forEach(child => {
            childIds.push(...getAllChildren(child.id))
          })
          return childIds
        }
        const allIdsToDelete = [id, ...getAllChildren(id)]
        const filtered = flatItems.filter(item => !allIdsToDelete.includes(item.id))
        items.value = buildTree(filtered)
        collaborationItemsStorage.setItems(filtered)
      }
    } finally {
      loading.value = false
    }
  }

  /**
   * 重新排序同一層級的項目
   */
  const reorderItems = async (itemIds: string[], parentId: string | null = null) => {
    loading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const data: ReorderItemsRequest = {
        item_ids: itemIds,
        parent_id: parentId
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
      const flatItems = flattenTree(items.value)
      itemIds.forEach((id, index) => {
        const item = flatItems.find(i => i.id === id)
        if (item) {
          item.order = index
          item.parent_id = parentId || null
        }
      })
      items.value = buildTree(flatItems)
      collaborationItemsStorage.setItems(flatItems)
    } finally {
      loading.value = false
    }
  }

  /**
   * 移動項目到不同父項目
   */
  const moveItem = async (itemId: string, newParentId: string | null) => {
    return updateItem(itemId, { parent_id: newParentId })
  }

  /**
   * 取得扁平列表（用於某些場景）
   */
  const flatItems = computed(() => {
    return flattenTree(items.value)
  })

  /**
   * 根據 ID 查找項目（遞迴查找）
   */
  const findItemById = (id: string, itemsList: CollaborationItem[] = items.value): CollaborationItem | null => {
    for (const item of itemsList) {
      if (item.id === id) {
        return item
      }
      if (item.children && item.children.length > 0) {
        const found = findItemById(id, item.children)
        if (found) {
          return found
        }
      }
    }
    return null
  }

  /**
   * 取得所有子項目 ID（遞迴）
   */
  const getAllChildrenIds = (parentId: string): string[] => {
    const flat = flatItems.value
    const result: string[] = []
    const getChildren = (pid: string | null) => {
      const children = flat.filter(item => item.parent_id === pid)
      children.forEach(child => {
        result.push(child.id)
        getChildren(child.id)
      })
    }
    getChildren(parentId)
    return result
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
    flatItems,

    // Actions
    fetchItems,
    createItem,
    updateItem,
    deleteItem,
    reorderItems,
    moveItem,
    buildTree,
    flattenTree,
    findItemById,
    getAllChildrenIds,
    reset
  }
})

