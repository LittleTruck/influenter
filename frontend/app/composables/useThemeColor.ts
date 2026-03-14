export type ThemeColor = 'green' | 'blue' | 'violet' | 'rose' | 'amber' | 'cyan' | 'indigo'

export const THEME_COLORS: { value: ThemeColor; label: string }[] = [
  { value: 'green', label: '綠色' },
  { value: 'blue', label: '藍色' },
  { value: 'violet', label: '紫色' },
  { value: 'rose', label: '玫瑰' },
  { value: 'amber', label: '琥珀' },
  { value: 'cyan', label: '青色' },
  { value: 'indigo', label: '靛藍' },
]

const STORAGE_KEY = 'influenter-primary-color'

export function useThemeColor() {
  const appConfig = useAppConfig()

  const primaryColor = computed<ThemeColor>({
    get: () => (appConfig.ui.colors?.primary as ThemeColor) || 'green',
    set: (color: ThemeColor) => {
      setPrimaryColor(color)
    },
  })

  function setPrimaryColor(color: ThemeColor) {
    if (!appConfig.ui.colors) {
      appConfig.ui.colors = { primary: color, neutral: 'slate' }
    } else {
      appConfig.ui.colors.primary = color
    }
    if (import.meta.client) {
      localStorage.setItem(STORAGE_KEY, color)
    }
  }

  function initThemeColor() {
    if (!import.meta.client) return
    const saved = localStorage.getItem(STORAGE_KEY) as ThemeColor | null
    if (saved && THEME_COLORS.some(c => c.value === saved)) {
      setPrimaryColor(saved)
    }
  }

  return {
    primaryColor,
    setPrimaryColor,
    initThemeColor,
  }
}
