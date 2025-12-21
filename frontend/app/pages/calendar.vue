<script setup lang="ts">
import { BaseDashboardPanel, BaseDashboardNavbar, BaseDashboardSidebarCollapse } from '~/components/base'
import CalendarView from '~/components/calendar/CalendarView.vue'
import { useCases } from '~/composables/useCases'
import LoadingState from '~/components/common/LoadingState.vue'
import ErrorState from '~/components/common/ErrorState.vue'

definePageMeta({
  middleware: 'auth'
})

const { fetchCases, loading, error, cases } = useCases()

// 載入案件數據
onMounted(async () => {
  try {
    await fetchCases()
  } catch (e) {
    console.error('載入案件失敗:', e)
  }
})

// 計算錯誤訊息
const errorMessage = computed(() => {
  if (!error.value) return ''
  if (typeof error.value === 'string') return error.value
  if (error.value instanceof Error) return error.value.message
  return '載入案件時發生錯誤'
})

// 即使有錯誤，如果有數據也顯示日曆
const shouldShowCalendar = computed(() => {
  return !loading.value && (cases.value.length > 0 || !error.value)
})
</script>

<template>
  <BaseDashboardPanel>
    <template #header>
      <BaseDashboardNavbar title="日曆">
        <template #leading>
          <BaseDashboardSidebarCollapse />
        </template>
      </BaseDashboardNavbar>
    </template>

    <template #body>
      <div class="w-full min-h-[600px] flex flex-col p-4 sm:p-6 lg:p-8">
        <LoadingState v-if="loading" />
        <ErrorState v-else-if="error && cases.length === 0" :message="errorMessage" />
        <div v-else class="w-full min-h-[600px] flex flex-col">
          <CalendarView />
        </div>
      </div>
    </template>
  </BaseDashboardPanel>
</template>




