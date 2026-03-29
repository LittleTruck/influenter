// Google Identity Services — 不再需要載入外部腳本
// GoogleLoginButton 使用手動 OAuth redirect，不依賴 GIS SDK
export default defineNuxtPlugin(() => {
  // no-op: 保留檔案避免其他地方 import 出錯
})

