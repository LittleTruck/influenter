<script setup lang="ts">
import type { CaseDetail, CaseEmail } from '~/types/cases'
import { BaseSlideover, BaseButton, BaseFormField, BaseSelect, BaseTextarea } from '~/components/base'

interface Props {
  modelValue: boolean
  caseId: string
  case: CaseDetail | null
  emails: CaseEmail[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'sent': []
}>()

const config = useRuntimeConfig()
const authStore = useAuthStore()
const toast = useToast()

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

interface ReplyTemplate {
  id: string
  title: string
  prompt: string
}

const selectedEmailId = ref<string>('')
const instruction = ref('')
const replyBody = ref('')
const generating = ref(false)
const sending = ref(false)
const selectedTemplateId = ref<string | undefined>(undefined)
const templates = ref<ReplyTemplate[]>([])

const emailOptions = computed(() => {
  return props.emails.map((e) => ({
    label: `${e.from_name || e.from_email}${e.subject ? ` - ${e.subject}` : ''}`,
    value: e.id
  }))
})

const templateOptions = computed(() =>
  templates.value.map((t) => ({
    label: t.title,
    value: t.id
  }))
)

const fetchTemplates = async () => {
  try {
    const res = await $fetch<{ data: ReplyTemplate[] }>(`${config.public.apiBase}/api/v1/reply-templates`, {
      headers: { Authorization: `Bearer ${authStore.token}` },
    })
    templates.value = res.data || []
  } catch {
    // silent
  }
}

const defaultEmailId = computed(() => {
  const list = props.emails
  if (list.length === 0) return ''
  const last = list[list.length - 1]
  return last ? last.id : ''
})

const activeEmailId = computed(() => selectedEmailId.value || defaultEmailId.value)

const handleGenerate = async () => {
  if (!activeEmailId.value) {
    toast.add({ title: '請選擇要回覆的郵件', color: 'warning' })
    return
  }

  generating.value = true
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
          email_id: activeEmailId.value,
          instruction: instruction.value || undefined,
          template_id: selectedTemplateId.value || undefined
        })
      }
    )
    replyBody.value = res.draft ?? ''
    toast.add({ title: '草稿已產生', color: 'success' })
  } catch (e: any) {
    const msg = e?.data?.message || e?.message || '產生草稿失敗'
    toast.add({ title: msg, color: 'error' })
  } finally {
    generating.value = false
  }
}

const handleSend = async () => {
  const body = replyBody.value?.trim()
  if (!body) {
    toast.add({ title: '請輸入回信內容', color: 'warning' })
    return
  }
  if (!activeEmailId.value) {
    toast.add({ title: '請選擇要回覆的郵件', color: 'warning' })
    return
  }

  sending.value = true
  try {
    await $fetch(
      `${config.public.apiBase}/api/v1/emails/${activeEmailId.value}/send-reply`,
      {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ body })
      }
    )
    toast.add({ title: '回信已寄出', color: 'success' })
    replyBody.value = ''
    instruction.value = ''
    isOpen.value = false
    emit('sent')
  } catch (e: any) {
    const msg = e?.data?.message || e?.message || '寄出失敗'
    toast.add({ title: msg, color: 'error' })
  } finally {
    sending.value = false
  }
}

const handleClose = () => {
  isOpen.value = false
}

watch(isOpen, (open) => {
  if (open) {
    selectedEmailId.value = defaultEmailId.value
    instruction.value = ''
    replyBody.value = ''
    selectedTemplateId.value = undefined
    fetchTemplates()
  }
})
</script>

<template>
  <BaseSlideover
    v-model="isOpen"
    title="回覆郵件"
    description="撰寫回信內容，或使用 AI 產生草稿"
    side="right"
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
            :disabled="generating || sending"
          />
        </BaseFormField>

        <BaseFormField v-if="templates.length > 0" label="回覆範本（選填）">
          <BaseSelect
            v-model="selectedTemplateId"
            :options="templateOptions"
            placeholder="選擇範本"
            class="w-full"
            :disabled="generating || sending"
          />
        </BaseFormField>

        <BaseFormField label="AI 補充說明（選填）">
          <BaseTextarea
            v-model="instruction"
            placeholder="例如：希望婉拒報價、或強調可配合的檔期…"
            :rows="2"
            :disabled="generating || sending"
          />
        </BaseFormField>

        <BaseButton
          icon="i-lucide-sparkles"
          variant="outline"
          :loading="generating"
          :disabled="emails.length === 0 || sending"
          @click="handleGenerate"
        >
          AI 產生草稿
        </BaseButton>

        <BaseFormField label="回信內容">
          <BaseTextarea
            v-model="replyBody"
            placeholder="可手動輸入或點「AI 產生草稿」填入"
            :rows="10"
            :disabled="generating || sending"
          />
        </BaseFormField>
      </div>
    </template>

    <template #footer>
      <div class="flex justify-end gap-2">
        <BaseButton
          color="neutral"
          variant="outline"
          :disabled="generating || sending"
          @click="handleClose"
        >
          取消
        </BaseButton>
        <BaseButton
          color="primary"
          icon="i-lucide-send"
          :loading="sending"
          :disabled="!replyBody?.trim() || generating"
          @click="handleSend"
        >
          寄出
        </BaseButton>
      </div>
    </template>
  </BaseSlideover>
</template>
