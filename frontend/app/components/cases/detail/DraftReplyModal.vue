<script setup lang="ts">
import type { CaseDetail, CaseEmail } from '~/types/cases'
import { BaseModal, BaseButton, BaseFormField, BaseSelect, BaseTextarea } from '~/components/base'

interface Props {
  modelValue: boolean
  caseId: string
  case: CaseDetail | null
  emails: CaseEmail[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const config = useRuntimeConfig()
const authStore = useAuthStore()
const toast = useToast()
const router = useRouter()

const DRAFT_STORAGE_KEY_PREFIX = 'email-reply-draft-'

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const selectedEmailId = ref<string>('')
const instruction = ref('')
const loading = ref(false)
const draft = ref('')

const emailOptions = computed(() => {
  return props.emails.map((e) => ({
    label: `${e.from_name || e.from_email}${e.subject ? ` - ${e.subject}` : ''}`,
    value: e.id
  }))
})

const defaultEmailId = computed(() => {
  const list = props.emails
  if (list.length === 0) return ''
  return list[list.length - 1].id
})

const hasDraft = computed(() => draft.value.length > 0)

const handleGenerate = async () => {
  const emailId = selectedEmailId.value || defaultEmailId.value
  if (!emailId) {
    toast.add({ title: '請選擇要回覆的郵件', color: 'warning' })
    return
  }

  loading.value = true
  draft.value = ''
  try {
    const res = await $fetch<{ draft: string }>(
      `${config.public.apiBase}/api/v1/cases/${props.caseId}/draft-reply`,
      {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          email_id: emailId,
          instruction: instruction.value || undefined
        })
      }
    )
    const draftText = res.draft ?? ''
    draft.value = draftText
    toast.add({ title: '草稿已產生', color: 'success' })

    // 存入 sessionStorage 並導向該信件的回覆頁
    try {
      sessionStorage.setItem(`${DRAFT_STORAGE_KEY_PREFIX}${emailId}`, draftText)
      isOpen.value = false
      router.push(`/emails/${emailId}?reply=draft`)
    } catch {
      // 導向失敗時草稿仍顯示在 modal 內
    }
  } catch (e: any) {
    const msg = e?.data?.message || e?.message || '產生草稿失敗'
    toast.add({ title: msg, color: 'error' })
  } finally {
    loading.value = false
  }
}

const handleCopy = async () => {
  try {
    await navigator.clipboard.writeText(draft.value)
    toast.add({ title: '已複製到剪貼簿', color: 'success' })
  } catch {
    toast.add({ title: '複製失敗', color: 'error' })
  }
}

const handleCancel = () => {
  selectedEmailId.value = defaultEmailId.value
  instruction.value = ''
  draft.value = ''
  isOpen.value = false
}

watch(isOpen, (open) => {
  if (open) {
    selectedEmailId.value = defaultEmailId.value
    instruction.value = ''
    draft.value = ''
  }
})
</script>

<template>
  <BaseModal
    v-model="isOpen"
    title="AI 擬信"
    description="選擇要回覆的郵件，AI 將根據案件與來信內容產生回信草稿"
    size="lg"
  >
    <template #body>
      <div class="space-y-4">
        <BaseFormField v-if="emails.length > 1" label="要回覆的郵件">
          <BaseSelect
            v-model="selectedEmailId"
            :options="emailOptions"
            placeholder="請選擇郵件"
            class="w-full"
            :disabled="loading"
          />
        </BaseFormField>

        <BaseFormField label="補充說明（選填）">
          <BaseTextarea
            v-model="instruction"
            placeholder="例如：希望婉拒報價、或強調可配合的檔期…"
            :rows="2"
            :disabled="loading"
          />
        </BaseFormField>

        <div v-if="hasDraft" class="space-y-2">
          <BaseFormField label="回信草稿">
            <BaseTextarea
              v-model="draft"
              :rows="10"
              placeholder="草稿將顯示於此，可編輯後複製使用"
              class="font-mono text-sm"
            />
          </BaseFormField>
          <div class="flex justify-end">
            <BaseButton
              icon="i-lucide-copy"
              variant="outline"
              size="sm"
              @click="handleCopy"
            >
              複製
            </BaseButton>
          </div>
        </div>
      </div>
    </template>

    <template #footer>
      <div class="flex justify-end gap-2">
        <BaseButton
          color="neutral"
          variant="outline"
          :disabled="loading"
          @click="handleCancel"
        >
          {{ hasDraft ? '關閉' : '取消' }}
        </BaseButton>
        <BaseButton
          v-if="!hasDraft"
          color="primary"
          icon="i-lucide-sparkles"
          :loading="loading"
          :disabled="emails.length === 0"
          @click="handleGenerate"
        >
          產生草稿
        </BaseButton>
        <BaseButton
          v-else
          color="primary"
          icon="i-lucide-sparkles"
          :loading="loading"
          @click="handleGenerate"
        >
          重新產生
        </BaseButton>
      </div>
    </template>
  </BaseModal>
</template>
