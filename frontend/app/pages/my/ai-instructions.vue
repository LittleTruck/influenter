<script setup lang="ts">
import { BaseButton } from '~/components/base'
import SectionPageHeader from '~/components/ui/SectionPageHeader.vue'

definePageMeta({
  middleware: 'auth'
})

const authStore = useAuthStore()
const toast = useToast()
const config = useRuntimeConfig()

// AI 注意事項
const aiInstructions = ref(authStore.user?.ai_instructions || '')
const savingAiInstructions = ref(false)

const saveAiInstructions = async () => {
  savingAiInstructions.value = true
  try {
    await $fetch(`${config.public.apiBase}/api/v1/auth/ai-instructions`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${authStore.token}` },
      body: { ai_instructions: aiInstructions.value || null },
    })
    // 同步到 store
    if (authStore.user) {
      authStore.user.ai_instructions = aiInstructions.value || undefined
    }
    toast.add({ title: '已儲存', description: 'AI 注意事項已更新' })
  } catch (e: any) {
    toast.add({ title: '儲存失敗', description: e?.message || '請稍後再試', color: 'error' })
  } finally {
    savingAiInstructions.value = false
  }
}
</script>

<template>
  <div>
    <SectionPageHeader
      icon="i-lucide-brain"
      title="AI 注意事項"
      description="AI 分析郵件和撰寫回覆時會參考這些注意事項"
    >
      <template #actions>
        <BaseButton
          color="primary"
          :loading="savingAiInstructions"
          @click="saveAiInstructions"
        >
          儲存
        </BaseButton>
      </template>
    </SectionPageHeader>

    <div class="space-y-4">
      <p class="text-sm text-muted">
        在此輸入您的合作注意事項（例如：修改規則、授權範圍、報價標準等），AI 擬信時會自動參考這些說明。
      </p>
      <textarea
        v-model="aiInstructions"
        class="w-full min-h-[400px] p-3 rounded-lg border border-default bg-elevated/50 text-highlighted placeholder-muted text-sm resize-y focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary"
        placeholder="例如：&#10;- 影片修改次數上限為 2 次&#10;- 不接受買斷授權&#10;- 報價以粉絲數 × 0.5 為基準"
      />
    </div>
  </div>
</template>
