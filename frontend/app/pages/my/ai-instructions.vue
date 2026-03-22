<script setup lang="ts">
import { BaseButton, BaseInput, BaseModal, BaseRichTextEditor } from '~/components/base'
import SectionPageHeader from '~/components/ui/SectionPageHeader.vue'

definePageMeta({
  middleware: 'auth'
})

interface ReplyTemplate {
  id: string
  title: string
  prompt: string
  order: number
}

const authStore = useAuthStore()
const toast = useToast()
const config = useRuntimeConfig()

// AI 助理設定
const aiReplyHeader = ref(authStore.user?.ai_reply_header || '')
const aiReplyFooter = ref(authStore.user?.ai_reply_footer || '')
const aiInstructions = ref(authStore.user?.ai_instructions || '')
const saving = ref(false)

const saveSettings = async () => {
  saving.value = true
  try {
    await $fetch(`${config.public.apiBase}/api/v1/auth/ai-instructions`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${authStore.token}` },
      body: {
        ai_instructions: aiInstructions.value || null,
        ai_reply_header: aiReplyHeader.value || null,
        ai_reply_footer: aiReplyFooter.value || null,
      },
    })
    if (authStore.user) {
      authStore.user.ai_instructions = aiInstructions.value || undefined
      authStore.user.ai_reply_header = aiReplyHeader.value || undefined
      authStore.user.ai_reply_footer = aiReplyFooter.value || undefined
    }
    toast.add({ title: '已儲存', description: 'AI 助理設定已更新' })
  } catch (e: any) {
    toast.add({ title: '儲存失敗', description: e?.message || '請稍後再試', color: 'error' })
  } finally {
    saving.value = false
  }
}

// 回覆範本
const templates = ref<ReplyTemplate[]>([])
const loadingTemplates = ref(false)
const showTemplateModal = ref(false)
const editingTemplate = ref<ReplyTemplate | null>(null)
const templateForm = ref({ title: '', prompt: '' })
const savingTemplate = ref(false)

const fetchTemplates = async () => {
  loadingTemplates.value = true
  try {
    const res = await $fetch<{ data: ReplyTemplate[] }>(`${config.public.apiBase}/api/v1/reply-templates`, {
      headers: { Authorization: `Bearer ${authStore.token}` },
    })
    templates.value = res.data || []
  } catch {
    // silent
  } finally {
    loadingTemplates.value = false
  }
}

const openCreateModal = () => {
  editingTemplate.value = null
  templateForm.value = { title: '', prompt: '' }
  showTemplateModal.value = true
}

const openEditModal = (tpl: ReplyTemplate) => {
  editingTemplate.value = tpl
  templateForm.value = { title: tpl.title, prompt: tpl.prompt }
  showTemplateModal.value = true
}

const saveTemplate = async () => {
  if (!templateForm.value.title.trim() || isHtmlEmpty(templateForm.value.prompt)) {
    toast.add({ title: '請填寫標題和提示詞', color: 'error' })
    return
  }
  savingTemplate.value = true
  try {
    if (editingTemplate.value) {
      await $fetch(`${config.public.apiBase}/api/v1/reply-templates/${editingTemplate.value.id}`, {
        method: 'PATCH',
        headers: { Authorization: `Bearer ${authStore.token}` },
        body: templateForm.value,
      })
      toast.add({ title: '已更新', description: '範本已更新' })
    } else {
      await $fetch(`${config.public.apiBase}/api/v1/reply-templates`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${authStore.token}` },
        body: templateForm.value,
      })
      toast.add({ title: '已新增', description: '範本已建立' })
    }
    showTemplateModal.value = false
    await fetchTemplates()
  } catch (e: any) {
    toast.add({ title: '儲存失敗', description: e?.message || '請稍後再試', color: 'error' })
  } finally {
    savingTemplate.value = false
  }
}

const deleteTemplate = async (tpl: ReplyTemplate) => {
  if (!confirm(`確定要刪除「${tpl.title}」範本嗎？`)) return
  try {
    await $fetch(`${config.public.apiBase}/api/v1/reply-templates/${tpl.id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${authStore.token}` },
    })
    toast.add({ title: '已刪除', description: '範本已刪除' })
    await fetchTemplates()
  } catch (e: any) {
    toast.add({ title: '刪除失敗', description: e?.message || '請稍後再試', color: 'error' })
  }
}

/** 檢查 HTML 內容是否實質為空 */
const isHtmlEmpty = (html: string) => {
  const text = html.replace(/<[^>]*>/g, '').trim()
  return !text
}

onMounted(() => {
  fetchTemplates()
})
</script>

<template>
  <div>
    <SectionPageHeader
      icon="i-lucide-brain"
      title="AI 助理設定"
      description="AI 擬信時會自動參考以下設定"
    >
      <template #actions>
        <BaseButton
          color="primary"
          :loading="saving"
          @click="saveSettings"
        >
          儲存
        </BaseButton>
      </template>
    </SectionPageHeader>

    <div class="space-y-8">
      <!-- 標頭設定 -->
      <div class="space-y-2">
        <label class="block text-sm font-medium text-highlighted">信件標頭</label>
        <p class="text-sm text-muted">
          AI 擬信時會自動加在回覆開頭的固定文字（例如：問候語、開場白）
        </p>
        <BaseRichTextEditor
          v-model="aiReplyHeader"
          placeholder="例如：您好，感謝您的來信！"
          min-height="100px"
        />
      </div>

      <!-- 標尾設定 -->
      <div class="space-y-2">
        <label class="block text-sm font-medium text-highlighted">信件標尾</label>
        <p class="text-sm text-muted">
          AI 擬信時會自動加在回覆結尾的固定文字（例如：簽名檔、結尾問候）
        </p>
        <BaseRichTextEditor
          v-model="aiReplyFooter"
          placeholder="例如：Best regards, [您的名字]"
          min-height="100px"
        />
      </div>

      <!-- 特殊需求設定 -->
      <div class="space-y-2">
        <label class="block text-sm font-medium text-highlighted">特殊需求</label>
        <p class="text-sm text-muted">
          AI 擬信時會參考的合作注意事項（例如：修改規則、授權範圍、報價標準等）
        </p>
        <BaseRichTextEditor
          v-model="aiInstructions"
          placeholder="例如：影片修改次數上限為 2 次、不接受買斷授權、報價以粉絲數 × 0.5 為基準"
          min-height="200px"
        />
      </div>

      <!-- 回覆範本 -->
      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <div>
            <label class="block text-sm font-medium text-highlighted">回覆範本</label>
            <p class="text-sm text-muted mt-1">
              建立常用的回覆範本，AI 擬信時可選擇範本來產生對應風格的草稿
            </p>
          </div>
          <BaseButton
            icon="i-lucide-plus"
            size="sm"
            variant="outline"
            @click="openCreateModal"
          >
            新增範本
          </BaseButton>
        </div>

        <!-- 範本列表 -->
        <div v-if="loadingTemplates" class="py-8 text-center text-sm text-muted">
          載入中...
        </div>
        <div v-else-if="templates.length === 0" class="py-8 text-center">
          <div class="text-muted text-sm">尚未建立任何範本</div>
          <div class="text-muted text-xs mt-1">點擊「新增範本」開始建立</div>
        </div>
        <div v-else class="space-y-3">
          <div
            v-for="tpl in templates"
            :key="tpl.id"
            class="rounded-lg border border-default bg-elevated/50 p-4"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="flex-1 min-w-0">
                <h4 class="text-sm font-medium text-highlighted">{{ tpl.title }}</h4>
                <p class="text-xs text-muted mt-1 line-clamp-2 whitespace-pre-line">{{ tpl.prompt }}</p>
              </div>
              <div class="flex items-center gap-1 shrink-0">
                <BaseButton
                  icon="i-lucide-pencil"
                  size="xs"
                  variant="ghost"
                  color="neutral"
                  @click="openEditModal(tpl)"
                />
                <BaseButton
                  icon="i-lucide-trash-2"
                  size="xs"
                  variant="ghost"
                  color="error"
                  @click="deleteTemplate(tpl)"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 新增/編輯範本 Modal -->
    <BaseModal
      v-model="showTemplateModal"
      :title="editingTemplate ? '編輯範本' : '新增範本'"
    >
      <div class="space-y-4">
        <div class="space-y-1">
          <label class="block text-sm font-medium text-highlighted">範本標題</label>
          <BaseInput
            v-model="templateForm.title"
            placeholder="例如：初次回覆、報價回覆、婉拒回覆"
          />
        </div>
        <div class="space-y-1">
          <label class="block text-sm font-medium text-highlighted">提示詞</label>
          <p class="text-xs text-muted">
            描述這個範本的回覆風格、內容重點，AI 會根據此提示詞產生對應的草稿
          </p>
          <BaseRichTextEditor
            v-model="templateForm.prompt"
            placeholder="例如：請以正式但友善的語氣回覆，先感謝對方的邀約，接著表達有興趣合作，並詢問更多細節如時程、預算和合作形式。"
            min-height="200px"
          />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-2">
          <BaseButton variant="ghost" @click="showTemplateModal = false">取消</BaseButton>
          <BaseButton color="primary" :loading="savingTemplate" @click="saveTemplate">
            {{ editingTemplate ? '更新' : '建立' }}
          </BaseButton>
        </div>
      </template>
    </BaseModal>
  </div>
</template>
