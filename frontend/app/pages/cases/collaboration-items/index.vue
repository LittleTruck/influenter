<script setup lang="ts">
import { nextTick } from 'vue'
import { useCollaborationItems } from '~/composables/useCollaborationItems'
import { useErrorHandler } from '~/composables/useErrorHandler'
import { BaseDashboardPanel, BaseDashboardNavbar, BaseDashboardSidebarCollapse, BaseButton, BaseTabs } from '~/components/base'
import CollaborationItemTree from '~/components/settings/collaboration-items/CollaborationItemTree.vue'
import CollaborationItemFormModal from '~/components/settings/collaboration-items/CollaborationItemFormModal.vue'
import WorkflowManagement from '~/components/settings/collaboration-items/WorkflowManagement.vue'
import LoadingState from '~/components/common/LoadingState.vue'
import EmptyState from '~/components/common/EmptyState.vue'

definePageMeta({
  middleware: 'auth'
})

const { items, loading, error, fetchItems, createItem, updateItem, deleteItem, reorderItems } = useCollaborationItems()
const { handleError, handleSuccess } = useErrorHandler()

// Tab 狀態
const activeTab = ref('items')

// WorkflowManagement 組件引用
const workflowManagementRef = ref<InstanceType<typeof WorkflowManagement> | null>(null)

// 表單狀態
const showItemForm = ref(false)
const editingItem = ref<any>(null)
const parentId = ref<string | null>(null)

// 頁面層面的 loading 保護：如果超過 3 秒還在載入，強制顯示內容
const pageLoadingTimeout = ref(false)

// 計算實際應該顯示的 loading 狀態（store loading 且未超時）
const isActuallyLoading = computed(() => loading.value && !pageLoadingTimeout.value)

// 監聽 loading 狀態變化，當它變為 false 時，確保頁面顯示內容
watch(loading, (newValue) => {
  if (!newValue) {
    pageLoadingTimeout.value = false
  }
})

// 載入項目列表（忽略 404 錯誤，因為後端還沒實作）
onMounted(async () => {
  // 使用 nextTick 確保組件完全掛載後再執行
  await nextTick()
  
  // 設置超時保護：如果 3 秒後還在載入，強制顯示內容
  const timeoutId = setTimeout(() => {
    pageLoadingTimeout.value = true
    console.debug('Page loading timeout, showing content anyway')
  }, 3000)
  
  try {
    await fetchItems()
  } catch (err: any) {
    // 忽略所有錯誤，因為 store 已經處理了
    console.debug('fetchItems completed with error (expected if API not available):', err)
  } finally {
    clearTimeout(timeoutId)
    // 確保頁面不會一直載入（即使 fetchItems 失敗，也顯示內容）
    pageLoadingTimeout.value = true
  }
})

// 處理新增項目
const handleAddItem = (parentIdValue?: string | null) => {
  editingItem.value = null
  parentId.value = parentIdValue || null
  showItemForm.value = true
}

// 處理新增流程
const handleAddWorkflow = () => {
  workflowManagementRef.value?.handleAddWorkflow()
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

// Tab 選項
const tabItems = [
  { label: '項目管理', value: 'items', icon: 'i-lucide-package' },
  { label: '流程管理', value: 'workflows', icon: 'i-lucide-list-checks' }
]
</script>

<template>
  <BaseDashboardPanel>
    <template #header>
      <BaseDashboardNavbar title="合作項目管理">
        <template #leading>
          <BaseDashboardSidebarCollapse />
        </template>

        <template #trailing>
          <BaseButton
            icon="i-lucide-arrow-left"
            variant="ghost"
            @click="navigateTo('/cases')"
          >
            返回案件列表
          </BaseButton>
        </template>
      </BaseDashboardNavbar>
    </template>

    <template #body>
      <!-- Tab 導航和新增按鈕 -->
      <div class="flex items-center justify-between mb-0">
        <div class="flex items-center bg-gray-100 dark:bg-gray-800 p-1 rounded-lg">
          <BaseButton
            :color="activeTab === 'items' ? 'primary' : 'neutral'"
            :variant="activeTab === 'items' ? 'solid' : 'ghost'"
            size="sm"
            icon="i-lucide-package"
            @click="activeTab = 'items'"
          >
            項目管理
          </BaseButton>
          <BaseButton
            :color="activeTab === 'workflows' ? 'primary' : 'neutral'"
            :variant="activeTab === 'workflows' ? 'solid' : 'ghost'"
            size="sm"
            icon="i-lucide-list-checks"
            @click="activeTab = 'workflows'"
          >
            流程管理
          </BaseButton>
        </div>
        <BaseButton
          v-if="activeTab === 'items'"
          icon="i-lucide-plus"
          size="sm"
          @click="handleAddItem()"
        >
          新增項目
        </BaseButton>
        <BaseButton
          v-else
          icon="i-lucide-plus"
          size="sm"
          @click="handleAddWorkflow"
        >
          新增流程
        </BaseButton>
      </div>

      <!-- 項目管理 Tab -->
      <div v-if="activeTab === 'items'">
        <LoadingState v-if="isActuallyLoading" />

        <template v-else>
          <EmptyState
            v-if="items.length === 0"
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
        </template>

        <!-- 項目表單 Modal -->
        <CollaborationItemFormModal
          v-model="showItemForm"
          :item="editingItem"
          :parent-id="parentId"
          @submit="handleFormSubmit"
        />
      </div>

      <!-- 流程管理 Tab -->
      <div v-if="activeTab === 'workflows'">
        <WorkflowManagement ref="workflowManagementRef" />
      </div>
    </template>
  </BaseDashboardPanel>
</template>

