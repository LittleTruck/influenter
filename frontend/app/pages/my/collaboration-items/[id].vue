<script setup lang="ts">
import type { CollaborationItemPhase, CreateCollaborationItemPhaseRequest, UpdateCollaborationItemPhaseRequest } from '~/types/collaborationItems'
import { useCollaborationItems } from '~/composables/useCollaborationItems'
import { useErrorHandler } from '~/composables/useErrorHandler'
import { BaseButton } from '~/components/base'
import SectionPageHeader from '~/components/ui/SectionPageHeader.vue'
import AppSection from '~/components/ui/AppSection.vue'
import PhaseList from '~/components/collaboration-items/PhaseList.vue'
import PhaseFormModal from '~/components/collaboration-items/PhaseFormModal.vue'
import LoadingState from '~/components/common/LoadingState.vue'
import ErrorState from '~/components/common/ErrorState.vue'

definePageMeta({
  middleware: 'auth'
})

const route = useRoute()
const router = useRouter()
const { fetchItems, findItemById } = useCollaborationItems()
const { handleError, handleSuccess } = useErrorHandler()

const itemId = computed(() => route.params.id as string)

// 階段列表
const phases = ref<CollaborationItemPhase[]>([])
const loadingPhases = ref(false)

// 當前合作項目
const currentItem = computed(() => findItemById(itemId.value))

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
const handleDeletePhase = async (_phase: CollaborationItemPhase) => {
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
const handleReorder = async (_updatedPhases: CollaborationItemPhase[]) => {
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
const handlePhaseSubmit = async (_data: CreateCollaborationItemPhaseRequest | UpdateCollaborationItemPhaseRequest) => {
  try {
    if (editingPhase.value) {
      // 更新階段
      // TODO: 呼叫 API 更新階段
      handleSuccess('階段已更新')
    } else {
      // 新增階段
      // TODO: 呼叫 API 新增階段
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
  <div>
    <!-- 載入中 -->
    <LoadingState v-if="loadingPhases" />

    <!-- 錯誤狀態 -->
    <ErrorState
      v-else-if="!currentItem"
      title="找不到合作項目"
      message="請返回列表選擇一個合作項目"
    />

    <!-- 內容 -->
    <template v-else>
      <SectionPageHeader
        icon="i-lucide-package"
        :title="currentItem.title || '合作項目階段管理'"
        :description="currentItem.description"
      >
        <template #actions>
          <BaseButton
            icon="i-lucide-arrow-left"
            variant="ghost"
            @click="router.push('/my/collaboration-items')"
          >
            返回列表
          </BaseButton>
        </template>
      </SectionPageHeader>

      <div class="space-y-6">
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
  </div>
</template>
