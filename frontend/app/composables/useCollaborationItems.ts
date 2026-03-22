/**
 * 合作項目相關 composable
 * 提供便捷的合作項目操作方法
 */
export const useCollaborationItems = () => {
  // 確保在正確的上下文中使用 store
  let itemsStore: ReturnType<typeof useCollaborationItemsStore>
  try {
    itemsStore = useCollaborationItemsStore()
  } catch (error) {
    console.error('Failed to initialize collaboration items store:', error)
    throw error
  }

  return {
    // Store 方法
    fetchItems: itemsStore.fetchItems,
    createItem: itemsStore.createItem,
    updateItem: itemsStore.updateItem,
    deleteItem: itemsStore.deleteItem,
    reorderItems: itemsStore.reorderItems,
    findItemById: itemsStore.findItemById,

    // Computed 值
    items: computed(() => itemsStore.items),
    individualItems: computed(() => itemsStore.individualItems),
    bundleItems: computed(() => itemsStore.bundleItems),
    loading: computed(() => itemsStore.loading),
    error: computed(() => itemsStore.error),
  }
}
