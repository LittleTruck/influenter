<script setup lang="ts">
import { BaseDashboardPanel, BaseDashboardNavbar, BaseDashboardSidebarCollapse, BaseButton, BaseIcon } from '~/components/base'
import AppSection from '~/components/ui/AppSection.vue'

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
  <BaseDashboardPanel>
    <template #header>
      <BaseDashboardNavbar title="AI 注意事項">
        <template #leading>
          <BaseDashboardSidebarCollapse />
        </template>
      </BaseDashboardNavbar>
    </template>

    <template #body>
      <div class="max-w-4xl mx-auto">
        <AppSection>
          <template #header>
            <div class="flex items-center gap-3">
              <BaseIcon name="i-lucide-brain" class="w-5 h-5 text-primary" />
              <h2 class="text-lg font-semibold text-highlighted">AI 注意事項</h2>
            </div>
          </template>

          <div class="space-y-4">
            <p class="text-sm text-muted">
              在此輸入您的合作注意事項（例如：修改規則、授權範圍、報價標準等），AI 擬信時會自動參考這些說明。
            </p>
            <textarea
              v-model="aiInstructions"
              class="w-full min-h-[120px] p-3 rounded-lg border border-default bg-elevated/50 text-highlighted placeholder-muted text-sm resize-y focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary"
              placeholder="例如：&#10;- 影片修改次數上限為 2 次&#10;- 不接受買斷授權&#10;- 報價以粉絲數 × 0.5 為基準"
            />
            <div class="flex justify-end">
              <BaseButton
                color="primary"
                :loading="savingAiInstructions"
                @click="saveAiInstructions"
              >
                儲存
              </BaseButton>
            </div>
          </div>
        </AppSection>
      </div>
    </template>
  </BaseDashboardPanel>
</template>
