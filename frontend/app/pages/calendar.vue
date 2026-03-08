<script setup lang="ts">
import { BaseDashboardPanel, BaseDashboardNavbar, BaseDashboardSidebarCollapse } from '~/components/base'
import CalendarView from '~/components/calendar/CalendarView.vue'
import LoadingState from '~/components/common/LoadingState.vue'
import ErrorState from '~/components/common/ErrorState.vue'

definePageMeta({
  middleware: 'auth'
})

const { loading, error, cases } = useCases()
const { fetchAllCaseDetails } = useCalendar()

// 載入案件數據（含階段詳情）
onMounted(async () => {
  try {
    await fetchAllCaseDetails()
  } catch (e) {
    // error 由 store 處理
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
  <BaseDashboardPanel grow>
    <template #header>
      <BaseDashboardNavbar title="日曆">
        <template #leading>
          <BaseDashboardSidebarCollapse />
        </template>
      </BaseDashboardNavbar>
    </template>

    <template #body>
      <div class="w-full flex flex-col flex-1 min-h-0 p-4">
        <LoadingState v-if="loading" />
        <ErrorState v-else-if="error && cases.length === 0" :message="errorMessage" />
        <CalendarView v-else class="flex-1 min-h-0" />
      </div>
    </template>
  </BaseDashboardPanel>
</template>




