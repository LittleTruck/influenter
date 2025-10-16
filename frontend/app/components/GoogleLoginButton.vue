<template>
  <div ref="googleButtonRef"></div>
</template>

<script setup lang="ts">
const props = defineProps<{
  callback: (response: any) => void
}>()

const googleButtonRef = ref<HTMLElement>()
const config = useRuntimeConfig()
const clientId = config.public.googleClientId as string

onMounted(() => {
  if (!googleButtonRef.value || !clientId) {
    console.error('Google button container or client ID not found')
    return
  }

  // 等待 Google Identity Services 腳本載入
  const checkGoogleLoaded = setInterval(() => {
    if (window.google) {
      clearInterval(checkGoogleLoaded)
      initializeGoogleButton()
    }
  }, 100)

  // 10 秒後停止檢查
  setTimeout(() => clearInterval(checkGoogleLoaded), 10000)
})

const initializeGoogleButton = () => {
  if (!googleButtonRef.value) return

  try {
    window.google.accounts.id.initialize({
      client_id: clientId,
      callback: props.callback,
    })

    window.google.accounts.id.renderButton(
      googleButtonRef.value,
      {
        theme: 'outline',
        size: 'large',
        width: 400,
        text: 'signin_with',
        shape: 'rectangular',
      }
    )

    // 可選：顯示 One Tap UI
    // window.google.accounts.id.prompt()
  } catch (error) {
    console.error('Failed to initialize Google Sign-In:', error)
  }
}
</script>

