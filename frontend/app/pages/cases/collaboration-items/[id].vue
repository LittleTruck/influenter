<script setup lang="ts">
import type { CollaborationItem, CollaborationItemPhase, CreateCollaborationItemPhaseRequest, UpdateCollaborationItemPhaseRequest } from '~/types/collaborationItems'
import { useCollaborationItems } from '~/composables/useCollaborationItems'
import { useErrorHandler } from '~/composables/useErrorHandler'
import { BaseDashboardPanel, BaseDashboardNavbar, BaseDashboardSidebarCollapse, BaseButton } from '~/components/base'
import PhaseList from '~/components/collaboration-items/PhaseList.vue'
import PhaseFormModal from '~/components/collaboration-items/PhaseFormModal.vue'
import LoadingState from '~/components/common/LoadingState.vue'
import ErrorState from '~/components/common/ErrorState.vue'
import AppSection from '~/components/ui/AppSection.vue'

definePageMeta({
  middleware: 'auth'
})

const route = useRoute()
const router = useRouter()
const { items, fetchItems } = useCollaborationItems()
const { handleError, handleSuccess } = useErrorHandler()

const itemId = computed(() => route.params.id as string)

// 階段列表
const phases = ref<CollaborationItemPhase[]>([])
const loadingPhases = ref(false)

// 當前合作項目
const currentItem = computed(() => {
  const flatten = (items: CollaborationItem[]): CollaborationItem[] => {
    const result: CollaborationItem[] = []
    items.forEach(item => {
      if (item.id === itemId.value) {
        result.push(item)
      }
      if (item.children && item.children.length > 0) {
        result.push(...flatten(item.children))
      }
    })
    return result
  }
  return flatten(items.value)[0]
})

// 表單狀態
const showPhaseForm = ref(false)
const editingPhase = ref<CollaborationItemPhase | null>(null)

// 載入數據
onMounted(async () => {
  await Promise.all([
    fetchItems(),
    fetchPhases()
  ])
})

// 載入階段列表
const fetchPhases = async () => {
  loadingPhases.value = true
  try {
    // TODO: 呼叫 API 載入階段列表
    // const data = await $fetch(`/api/v1/collaboration-items/${itemId.value}/phases`)
    // phases.value = data
    phases.value = []
  } catch (error: unknown) {
    handleError(error, '載入階段列表失敗')
  } finally {
    loadingPhases.value = false
  }
}

// 處理新增階段
const handleAddPhase = () => {
  editingPhase.value = null
  showPhaseForm.value = true
}

// 處理編輯階段
const handleEditPhase = (phase: CollaborationItemPhase) => {
  editingPhase.value = phase
  showPhaseForm.value = true
}

// 處理刪除階段
const handleDeletePhase = async (phase: CollaborationItemPhase) => {
  try {
    // TODO: 呼叫 API 刪除階段
    // await $fetch(`/api/v1/collaboration-items/${itemId.value}/phases/${phase.id}`, { method: 'DELETE' })
    handleSuccess('階段已刪除')
    await fetchPhases()
  } catch (error: unknown) {
    handleError(error, '刪除失敗')
  }
}

// 處理重新排序
const handleReorder = async (updatedPhases: CollaborationItemPhase[]) => {
  try {
    // TODO: 呼叫 API 重新排序
    // await $fetch(`/api/v1/collaboration-items/${itemId.value}/phases/reorder`, {
    //   method: 'PATCH',
    //   body: { phase_ids: updatedPhases.map(p => p.id) }
    // })
    handleSuccess('排序已更新')
    await fetchPhases()
  } catch (error: unknown) {
    handleError(error, '排序失敗')
  }
}

// 處理表單提交
const handlePhaseSubmit = async (data: CreateCollaborationItemPhaseRequest | UpdateCollaborationItemPhaseRequest) => {
  try {
    if (editingPhase.value) {
      // 更新階段
      // TODO: 呼叫 API 更新階段
      // await $fetch(`/api/v1/collaboration-items/${itemId.value}/phases/${editingPhase.value.id}`, {
      //   method: 'PATCH',
      //   body: data
      // })
      handleSuccess('階段已更新')
    } else {
      // 新增階段
      // TODO: 呼叫 API 新增階段
      // await $fetch(`/api/v1/collaboration-items/${itemId.value}/phases`, {
      //   method: 'POST',
      //   body: data
      // })
      handleSuccess('階段已新增')
    }
    showPhaseForm.value = false
    editingPhase.value = null
    await fetchPhases()
  } catch (error: unknown) {
    handleError(error, '操作失敗')
  }
}
</script>

<template>
  <BaseDashboardPanel>
    <template #header>
      <BaseDashboardNavbar :title="currentItem?.title || '合作項目階段管理'">
        <template #leading>
          <BaseDashboardSidebarCollapse />
        </template>

        <template #trailing>
          <BaseButton
            icon="i-lucide-arrow-left"
            variant="ghost"
            @click="router.push('/cases/collaboration-items')"
          >
            返回列表
          </BaseButton>
        </template>
      </BaseDashboardNavbar>
    </template>

    <template #body>
      <!-- 載入中 -->
      <LoadingState v-if="loadingPhases" />

      <!-- 錯誤狀態 -->
      <ErrorState
        v-else-if="!currentItem"
        title="找不到合作項目"
        message="請返回列表選擇一個合作項目"
      />

      <!-- 內容 -->
      <div v-else class="space-y-6">
        <!-- 合作項目資訊 -->
        <AppSection>
          <template #header>
            <h2 class="text-lg font-semibold">合作項目資訊</h2>
          </template>
          <div>
            <h3 class="text-xl font-bold text-highlighted mb-2">{{ currentItem.title }}</h3>
            <p v-if="currentItem.description" class="text-muted">
              {{ currentItem.description }}
            </p>
          </div>
        </AppSection>

        <!-- 階段流程管理 -->
        <AppSection>
          <template #header>
            <div>
              <h2 class="text-lg font-semibold">階段流程</h2>
              <p class="text-sm text-muted mt-1">
                定義此合作項目的預設階段流程，案件套用流程時會使用這些設定
              </p>
            </div>
          </template>
          <PhaseList
            :phases="phases"
            :loading="loadingPhases"
            :editable="true"
            @add-phase="handleAddPhase"
            @edit-phase="handleEditPhase"
            @delete-phase="handleDeletePhase"
            @reorder="handleReorder"
          />
        </AppSection>
      </div>

      <!-- 階段表單 Modal -->
      <PhaseFormModal
        v-model="showPhaseForm"
        :phase="editingPhase"
        :collaboration-item-id="itemId"
        @submit="handlePhaseSubmit"
      />
    </template>
  </BaseDashboardPanel>
</template>

