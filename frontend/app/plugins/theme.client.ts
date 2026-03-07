export default defineNuxtPlugin(() => {
  const { initThemeColor } = useThemeColor()
  initThemeColor()
})
