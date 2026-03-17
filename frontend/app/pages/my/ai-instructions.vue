<script setup lang="ts">
import { BaseButton } from '~/components/base'
import SectionPageHeader from '~/components/ui/SectionPageHeader.vue'

definePageMeta({
  middleware: 'auth'
})

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
    // 同步到 store
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
        <textarea
          v-model="aiReplyHeader"
          class="w-full min-h-[100px] p-3 rounded-lg border border-default bg-elevated/50 text-highlighted placeholder-muted text-sm resize-y focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary"
          placeholder="例如：&#10;您好，感謝您的來信！"
        />
      </div>

      <!-- 標尾設定 -->
      <div class="space-y-2">
        <label class="block text-sm font-medium text-highlighted">信件標尾</label>
        <p class="text-sm text-muted">
          AI 擬信時會自動加在回覆結尾的固定文字（例如：簽名檔、結尾問候）
        </p>
        <textarea
          v-model="aiReplyFooter"
          class="w-full min-h-[100px] p-3 rounded-lg border border-default bg-elevated/50 text-highlighted placeholder-muted text-sm resize-y focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary"
          placeholder="例如：&#10;Best regards,&#10;[您的名字]"
        />
      </div>

      <!-- 特殊需求設定 -->
      <div class="space-y-2">
        <label class="block text-sm font-medium text-highlighted">特殊需求</label>
        <p class="text-sm text-muted">
          AI 擬信時會參考的合作注意事項（例如：修改規則、授權範圍、報價標準等）
        </p>
        <textarea
          v-model="aiInstructions"
          class="w-full min-h-[200px] p-3 rounded-lg border border-default bg-elevated/50 text-highlighted placeholder-muted text-sm resize-y focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary"
          placeholder="例如：&#10;- 影片修改次數上限為 2 次&#10;- 不接受買斷授權&#10;- 報價以粉絲數 × 0.5 為基準"
        />
      </div>

      <!-- 範本設定（稍後開發） -->
      <div class="space-y-2 opacity-50">
        <label class="block text-sm font-medium text-highlighted">回覆範本</label>
        <p class="text-sm text-muted">
          預設的回覆範本，AI 可參考範本格式產生草稿（即將推出）
        </p>
        <div class="w-full min-h-[80px] p-4 rounded-lg border border-dashed border-default bg-elevated/30 flex items-center justify-center">
          <span class="text-sm text-muted">即將推出</span>
        </div>
      </div>
    </div>
  </div>
</template>
