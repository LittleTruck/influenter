<script setup lang="ts">
import type { CaseDetail } from '~/types/cases'
import type { CaseField } from '~/types/fields'
import { BaseButton, BaseIcon } from '~/components/base'
import FieldInput from '~/components/cases/fields/FieldInput.vue'
import AppSectionWithHeader from '~/components/ui/AppSectionWithHeader.vue'

interface Props {
  /** 案件詳情 */
  case: CaseDetail
  /** 屬性列表 */
  fields: CaseField[]
  /** 是否可編輯 */
  editable?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  editable: true
})

const emit = defineEmits<{
  'field-update': [fieldName: string, value: unknown]
  'field-delete': [fieldId: string]
}>()

// 編輯模式
const isEditing = ref(false)


// 取得屬性值
const getFieldValue = (field: CaseField): unknown => {
  // 系統屬性從 case 物件取得
  if (field.is_system) {
    const caseData = props.case as Record<string, unknown>
    return caseData[field.name]
  }
  // 自定義屬性從 custom_fields 取得
  const caseData = props.case as { custom_fields?: Record<string, unknown> }
  return caseData.custom_fields?.[field.name]
}

// 處理屬性更新
const handleFieldUpdate = (fieldName: string, value: unknown): void => {
  emit('field-update', fieldName, value)
}

// 處理屬性刪除
const handleDeleteField = (fieldId: string) => {
  emit('field-delete', fieldId)
}

// 排序後的屬性列表（系統屬性在前，自定義屬性在後）
const sortedFields = computed(() => {
  return [...props.fields].sort((a, b) => {
    // 系統屬性在前
    if (a.is_system && !b.is_system) return -1
    if (!a.is_system && b.is_system) return 1
    // 同類型按 order 排序
    return a.order - b.order
  })
})
</script>

<template>
  <AppSectionWithHeader
    title="屬性"
    description="案件的詳細屬性資訊"
  >
    <template #actions>
      <BaseButton
        v-if="editable"
        :icon="isEditing ? 'i-lucide-x' : 'i-lucide-edit'"
        variant="ghost"
        size="sm"
        @click="isEditing = !isEditing"
      >
        {{ isEditing ? '取消' : '編輯' }}
      </BaseButton>
    </template>

    <div class="properties-panel space-y-1">
      <div
        v-for="field in sortedFields"
        :key="field.id"
        class="property-row group flex items-center gap-4 py-2 px-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
      >
        <!-- 左側：屬性名稱 -->
        <div class="property-label flex items-center gap-2 min-w-[150px] flex-shrink-0">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ field.label }}
          </span>
        </div>

        <!-- 右側：屬性值（可編輯） -->
        <div class="property-value flex-1 min-w-0">
          <FieldInput
            :field="field"
            :model-value="getFieldValue(field)"
            :editable="isEditing && !field.is_system"
            @update:model-value="handleFieldUpdate(field.name, $event)"
          />
        </div>

        <!-- 刪除按鈕（僅自定義屬性，編輯模式下顯示） -->
        <BaseButton
          v-if="!field.is_system && isEditing"
          icon="i-lucide-trash-2"
          variant="ghost"
          size="xs"
          color="error"
          class="flex-shrink-0"
          @click="handleDeleteField(field.id)"
        />
      </div>
    </div>
  </AppSectionWithHeader>
</template>

<style scoped>
.property-row {
  position: relative;
}

.property-value :deep(input:focus),
.property-value :deep(textarea:focus),
.property-value :deep(select:focus) {
  outline: none;
  box-shadow: 0 0 0 2px rgb(var(--color-primary-500) / 0.5);
  border-radius: 4px;
  animation: focus-in 0.2s ease-out;
}

@keyframes focus-in {
  from {
    transform: scale(0.98);
    opacity: 0.8;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}
</style>
