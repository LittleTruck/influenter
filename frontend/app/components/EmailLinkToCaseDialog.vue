<script setup lang="ts">
import { BaseModal, BaseButton, BaseInput, BaseTextarea, BaseSelect, BaseIcon } from '~/components/base'

const props = defineProps<{
  emailId: string
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  linked: [caseId: string]
}>()

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

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
    isOpen.value = false
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

// 監聽 modelValue 變化
watch(() => props.modelValue, (open) => {
  if (open) {
    // TODO: 載入案件列表
    // fetchCases()
  }
})
</script>

<template>
  <BaseModal v-model="isOpen" title="關聯郵件到案件" size="md">
    <template #body>
      <div class="space-y-6">
        <!-- 關聯到現有案件 -->
        <div v-if="!showCreateForm">
          <label class="block text-sm font-medium text-highlighted mb-2">
            選擇現有案件
          </label>
          
          <BaseSelect
            v-model="selectedCaseId"
            :items="cases"
            value-key="id"
            placeholder="選擇案件..."
            class="w-full mb-4"
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
          </BaseSelect>

          <div class="flex justify-end gap-2">
            <BaseButton
              color="neutral"
              variant="outline"
              @click="showCreateForm = true"
            >
              或建立新案件
            </BaseButton>
            <BaseButton
              color="primary"
              :loading="loading"
              :disabled="!selectedCaseId"
              @click="linkToExistingCase"
            >
              關聯
            </BaseButton>
          </div>
        </div>

        <!-- 建立新案件 -->
        <div v-else>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-highlighted mb-2">
                案件名稱 *
              </label>
              <BaseInput
                v-model="newCase.title"
                placeholder="例如：Nike 球鞋業配"
                class="w-full"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-highlighted mb-2">
                品牌名稱 *
              </label>
              <BaseInput
                v-model="newCase.brand_name"
                placeholder="例如：Nike"
                class="w-full"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-highlighted mb-2">
                案件描述
              </label>
              <BaseTextarea
                v-model="newCase.description"
                placeholder="案件的詳細描述..."
                :rows="3"
                class="w-full"
              />
            </div>

            <div class="flex justify-end gap-2">
              <BaseButton
                color="neutral"
                variant="outline"
                @click="showCreateForm = false"
              >
                返回
              </BaseButton>
              <BaseButton
                color="primary"
                :loading="creating"
                @click="createAndLink"
              >
                建立並關聯
              </BaseButton>
            </div>
          </div>
        </div>

        <!-- 空狀態 -->
        <div v-if="!showCreateForm && cases.length === 0" class="text-center py-8">
          <BaseIcon name="i-lucide-inbox" class="w-12 h-12 mx-auto mb-3 text-muted" />
          <p class="text-muted mb-4">目前沒有案件</p>
          <BaseButton
            color="primary"
            variant="outline"
            @click="showCreateForm = true"
          >
            建立第一個案件
          </BaseButton>
        </div>
      </div>
    </template>
  </BaseModal>
</template>
