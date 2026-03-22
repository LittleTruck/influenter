<script setup lang="ts">
import { nextTick } from 'vue'
import type { CollaborationItemType } from '~/types/collaborationItems'
import { useCollaborationItems } from '~/composables/useCollaborationItems'
import { useErrorHandler } from '~/composables/useErrorHandler'
import { BaseButton } from '~/components/base'
import SectionPageHeader from '~/components/ui/SectionPageHeader.vue'
import CollaborationItemTree from '~/components/settings/collaboration-items/CollaborationItemTree.vue'
import CollaborationItemFormModal from '~/components/settings/collaboration-items/CollaborationItemFormModal.vue'
import LoadingState from '~/components/common/LoadingState.vue'
import EmptyState from '~/components/common/EmptyState.vue'

definePageMeta({
  middleware: 'auth'
})

const { items, individualItems, bundleItems, loading, fetchItems, deleteItem, reorderItems } = useCollaborationItems()
const { handleError, handleSuccess } = useErrorHandler()

// 表單狀態
const showItemForm = ref(false)
const editingItem = ref<any>(null)
const defaultType = ref<CollaborationItemType>('individual')

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

// 載入項目列表
onMounted(async () => {
  await nextTick()

  const timeoutId = setTimeout(() => {
    pageLoadingTimeout.value = true
    console.debug('Page loading timeout, showing content anyway')
  }, 3000)

  try {
    await fetchItems()
  } catch (err: any) {
    console.debug('fetchItems completed with error (expected if API not available):', err)
  } finally {
    clearTimeout(timeoutId)
    pageLoadingTimeout.value = true
  }
})

// 處理新增項目
const handleAddItem = (type?: CollaborationItemType) => {
  editingItem.value = null
  defaultType.value = type || 'individual'
  showItemForm.value = true
}

// 處理編輯項目
const handleEditItem = (item: any) => {
  editingItem.value = item
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
const handleReorder = async (itemIds: string[]) => {
  try {
    await reorderItems(itemIds)
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
}
</script>

<template>
  <div>
    <SectionPageHeader
      icon="i-lucide-package"
      title="合作項目"
      description="定義您提供的合作方案與報價，建立案件時可直接套用"
    >
      <template #actions>
        <BaseButton
          icon="i-lucide-plus"
          size="sm"
          @click="handleAddItem('individual')"
        >
          新增單項
        </BaseButton>
        <BaseButton
          icon="i-lucide-plus"
          size="sm"
          variant="outline"
          @click="handleAddItem('bundle')"
        >
          新增組合
        </BaseButton>
      </template>
    </SectionPageHeader>

    <LoadingState v-if="isActuallyLoading" />

    <template v-else>
      <EmptyState
        v-if="items.length === 0"
        icon="i-lucide-package"
        title="還沒有合作項目"
        action-label="建立第一個項目"
        :show-icon-background="false"
        @action="handleAddItem('individual')"
      />

      <CollaborationItemTree
        v-else
        :individual-items="individualItems"
        :bundle-items="bundleItems"
        @edit-item="handleEditItem"
        @delete-item="handleDeleteItem"
        @reorder="handleReorder"
      />
    </template>

    <!-- 項目表單 Modal -->
    <CollaborationItemFormModal
      v-model="showItemForm"
      :item="editingItem"
      :default-type="defaultType"
      @submit="handleFormSubmit"
    />
  </div>
</template>
