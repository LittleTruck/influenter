import { defineStore } from 'pinia'
import { defaultDayCalcConfig, type DayCalcConfig } from '~/utils/workdays'

export interface Holiday {
  date: string // "yyyy-MM-dd"
  name: string
  is_workday: boolean // true=補班日
  source: string
}

/**
 * 國定假日資料 store。
 * 全 app 載入一次，提供給工作日計算（工期預覽、期限倒數）使用。
 */
export const useHolidaysStore = defineStore('holidays', () => {
  const holidays = ref<Holiday[]>([])
  const loaded = ref(false)
  const loading = ref(false)

  const fetchHolidays = async (force = false) => {
    if ((loaded.value || loading.value) && !force) return
    loading.value = true
    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()
      const res = await $fetch<{ data: Holiday[] }>(`${config.public.apiBase}/api/v1/holidays`, {
        headers: { Authorization: `Bearer ${authStore.token}` }
      })
      holidays.value = res.data || []
      loaded.value = true
    } catch {
      // 載入失敗：退化為「僅排除週末」（dayCalcConfig 仍可用），下次仍可重試
      loaded.value = false
    } finally {
      loading.value = false
    }
  }

  // 由假日清單組出工作日計算設定（reactive：載入後使用端的 computed 會自動重算）
  const dayCalcConfig = computed<DayCalcConfig>(() => {
    const cfg = defaultDayCalcConfig()
    for (const h of holidays.value) {
      if (h.is_workday) cfg.makeupWorkdays.add(h.date)
      else cfg.holidays.add(h.date)
    }
    return cfg
  })

  return { holidays, loaded, loading, fetchHolidays, dayCalcConfig }
})
