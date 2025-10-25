<template>
  <UApp :toaster="toasterConfig">
    <NuxtLoadingIndicator />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </UApp>
</template>

<script setup lang="ts">
// 主應用程式入口
// 初始化 Toast 功能
const toast = useToast()

// Toaster 配置
const toasterConfig = {
  position: 'bottom-right',
  duration: 5000,
  expand: true,
  progress: true
}
</script>



