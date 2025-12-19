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
    moveItem: itemsStore.moveItem,
    findItemById: itemsStore.findItemById,
    getAllChildrenIds: itemsStore.getAllChildrenIds,

    // Computed 值
    items: computed(() => itemsStore.items),
    flatItems: computed(() => itemsStore.flatItems),
    loading: computed(() => itemsStore.loading),
    error: computed(() => itemsStore.error),

    // 工具方法
    buildTree: itemsStore.buildTree,
    flattenTree: itemsStore.flattenTree
  }
}

