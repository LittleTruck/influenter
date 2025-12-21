<script setup lang="ts">
/**
 * BaseFormGroup - 表單群組元件的基礎封裝
 * 封裝 UFormGroup，提供統一的表單群組介面
 */

defineOptions({ inheritAttrs: false })

interface Props {
  /** 標籤文字 */
  label?: string
  /** 描述文字 */
  description?: string
  /** 錯誤訊息 */
  error?: string
  /** 必填標記 */
  required?: boolean
}

const props = defineProps<Props>()
</script>

<template>
  <UFormGroup
    v-bind="$attrs"
    :label="label"
    :description="description"
    :error="error"
    :required="required"
    class="base-form-group"
  >
    <slot />
    <template v-if="$slots.description" #description>
      <slot name="description" />
    </template>
  </UFormGroup>
</template>

<style scoped>
.base-form-group {
  /* 可以在這裡添加全域表單群組樣式 */
}
</style>



