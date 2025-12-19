<script setup lang="ts">
const props = defineProps<{
  emailId: string
  open: boolean
}>()

const emit = defineEmits<{
  close: []
  linked: [caseId: string]
}>()

const emailsStore = useEmailsStore()
const toast = useToast()

// Cases store (未來實作)
const cases = ref<Array<{ id: string; title: string; brand_name: string }>>([])
const selectedCaseId = ref<string>('')
const loading = ref(false)
const creating = ref(false)

// 新建案件的表單
const newCase = ref({
  title: '',
  brand_name: '',
  description: ''
})

const showCreateForm = ref(false)

// 關聯現有案件
const linkToExistingCase = async () => {
  if (!selectedCaseId.value) {
    toast.add({
      title: '請選擇案件',
      color: 'warning'
    })
    return
  }

  loading.value = true
  try {
    await emailsStore.linkToCase(props.emailId, selectedCaseId.value)
    toast.add({
      title: '關聯成功',
      color: 'success'
    })
    emit('linked', selectedCaseId.value)
    emit('close')
  } catch (e) {
    toast.add({
      title: '關聯失敗',
      color: 'error'
    })
  } finally {
    loading.value = false
  }
}

// 創建新案件並關聯
const createAndLink = async () => {
  if (!newCase.value.title || !newCase.value.brand_name) {
    toast.add({
      title: '請填寫必要欄位',
      color: 'warning'
    })
    return
  }

  creating.value = true
  try {
    // TODO: 呼叫案件 API 創建案件
    toast.add({
      title: '功能開發中',
      description: '案件管理功能即將推出',
      color: 'info'
    })
  } catch (e) {
    toast.add({
      title: '創建失敗',
      color: 'error'
    })
  } finally {
    creating.value = false
  }
}

// 監聽 open 變化
watch(() => props.open, (isOpen) => {
  if (isOpen) {
    // TODO: 載入案件列表
    // fetchCases()
  }
})
</script>

<template>
  <UModal :open="open" @close="emit('close')">
    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="font-semibold text-lg text-highlighted">關聯郵件到案件</h3>
          <UButton
            icon="i-lucide-x"
            color="neutral"
            variant="ghost"
            size="sm"
            @click="emit('close')"
            aria-label="關閉"
          />
        </div>
      </template>

      <div class="space-y-6">
        <!-- 關聯到現有案件 -->
        <div v-if="!showCreateForm">
          <label class="block text-sm font-medium text-highlighted mb-2">
            選擇現有案件
          </label>
          
          <USelectMenu
            v-model="selectedCaseId"
            :items="cases"
            value-attribute="id"
            option-attribute="title"
            placeholder="選擇案件..."
            class="mb-4"
          >
            <template #label>
              <template v-if="selectedCaseId">
                {{ cases.find(c => c.id === selectedCaseId)?.title }}
              </template>
              <template v-else>
                選擇案件...
              </template>
            </template>

            <template #option="{ option }">
              <div class="flex flex-col">
                <span class="font-medium">{{ option.title }}</span>
                <span class="text-sm text-muted">{{ option.brand_name }}</span>
              </div>
            </template>
          </USelectMenu>

          <div class="flex gap-2">
            <UButton
              color="primary"
              :loading="loading"
              :disabled="!selectedCaseId"
              @click="linkToExistingCase"
              block
            >
              關聯
            </UButton>
            
            <UButton
              color="neutral"
              variant="outline"
              @click="showCreateForm = true"
              block
            >
              或建立新案件
            </UButton>
          </div>
        </div>

        <!-- 建立新案件 -->
        <div v-else>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-highlighted mb-2">
                案件名稱 *
              </label>
              <UInput
                v-model="newCase.title"
                placeholder="例如：Nike 球鞋業配"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-highlighted mb-2">
                品牌名稱 *
              </label>
              <UInput
                v-model="newCase.brand_name"
                placeholder="例如：Nike"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-highlighted mb-2">
                案件描述
              </label>
              <UTextarea
                v-model="newCase.description"
                placeholder="案件的詳細描述..."
                :rows="3"
              />
            </div>

            <div class="flex gap-2">
              <UButton
                color="primary"
                :loading="creating"
                @click="createAndLink"
                block
              >
                建立並關聯
              </UButton>
              
              <UButton
                color="neutral"
                variant="outline"
                @click="showCreateForm = false"
                block
              >
                返回
              </UButton>
            </div>
          </div>
        </div>

        <!-- 空狀態 -->
        <div v-if="!showCreateForm && cases.length === 0" class="text-center py-8">
          <UIcon name="i-lucide-inbox" class="w-12 h-12 mx-auto mb-3 text-muted" />
          <p class="text-muted mb-4">目前沒有案件</p>
          <UButton
            color="primary"
            variant="outline"
            @click="showCreateForm = true"
          >
            建立第一個案件
          </UButton>
        </div>
      </div>
    </UCard>
  </UModal>
</template>

