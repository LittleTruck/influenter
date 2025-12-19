<script setup lang="ts">
import { useCollaborationItems } from '~/composables/useCollaborationItems'
import { useErrorHandler } from '~/composables/useErrorHandler'
import CollaborationItemTree from '~/components/settings/collaboration-items/CollaborationItemTree.vue'
import CollaborationItemFormModal from '~/components/settings/collaboration-items/CollaborationItemFormModal.vue'
import LoadingState from '~/components/common/LoadingState.vue'
import EmptyState from '~/components/common/EmptyState.vue'

definePageMeta({
  middleware: 'auth'
})

const { items, loading, fetchItems, createItem, updateItem, deleteItem, reorderItems } = useCollaborationItems()
const { handleError, handleSuccess } = useErrorHandler()

// 表單狀態
const showItemForm = ref(false)
const editingItem = ref<any>(null)
const parentId = ref<string | null>(null)

// 載入項目列表
onMounted(async () => {
  await fetchItems()
})

// 處理新增項目
const handleAddItem = (parentIdValue?: string | null) => {
  editingItem.value = null
  parentId.value = parentIdValue || null
  showItemForm.value = true
}

// 處理編輯項目
const handleEditItem = (item: any) => {
  editingItem.value = item
  parentId.value = null
  showItemForm.value = true
}

// 處理刪除項目
const handleDeleteItem = async (item: any) => {
  try {
    await deleteItem(item.id)
    handleSuccess('項目已刪除')
  } catch (error: unknown) {
    handleError(error, '刪除失敗')
  }
}

// 處理重新排序
const handleReorder = async (itemIds: string[], parentId: string | null) => {
  try {
    await reorderItems(itemIds, parentId)
    handleSuccess('排序已更新')
  } catch (error: unknown) {
    handleError(error, '排序失敗')
  }
}

// 處理表單提交
const handleFormSubmit = () => {
  fetchItems()
  showItemForm.value = false
  editingItem.value = null
  parentId.value = null
}
</script>

<template>
  <UDashboardPanel>
    <template #header>
      <UDashboardNavbar title="合作項目管理">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>

        <template #trailing>
          <UButton
            icon="i-lucide-arrow-left"
            variant="ghost"
            @click="navigateTo('/cases')"
          >
            返回案件列表
          </UButton>
          <UButton
            icon="i-lucide-plus"
            @click="handleAddItem()"
          >
            新增項目
          </UButton>
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <LoadingState v-if="loading" />

      <EmptyState
        v-else-if="items.length === 0"
        icon="i-lucide-package"
        title="還沒有合作項目"
        action-label="建立第一個項目"
        :show-icon-background="false"
        @action="handleAddItem()"
      />

      <CollaborationItemTree
        v-else
        :items="items"
        @add-item="handleAddItem"
        @edit-item="handleEditItem"
        @delete-item="handleDeleteItem"
        @reorder="handleReorder"
      />

      <!-- 項目表單 Modal -->
      <CollaborationItemFormModal
        v-model="showItemForm"
        :item="editingItem"
        :parent-id="parentId"
        @submit="handleFormSubmit"
      />
    </template>
  </UDashboardPanel>
</template>

