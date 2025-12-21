<script setup lang="ts">
/**
 * BaseSwitch - Switch 元件的基礎封裝
 * 封裝 USwitch，提供統一的開關介面
 */

defineOptions({ inheritAttrs: false })

interface Props {
  modelValue?: boolean
  disabled?: boolean
  size?: 'xs' | 'sm' | 'md' | 'lg'
  color?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  disabled: false,
  size: 'sm',
  color: 'primary'
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const handleUpdate = (value: boolean) => {
  emit('update:modelValue', value)
}
</script>

<template>
  <USwitch
    v-bind="$attrs"
    :model-value="props.modelValue"
    :disabled="props.disabled"
    :size="props.size"
    :color="props.color"
    @update:model-value="handleUpdate"
  >
    <slot />
  </USwitch>
</template>



