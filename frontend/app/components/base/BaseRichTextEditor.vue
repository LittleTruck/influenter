<script setup lang="ts">
/**
 * BaseRichTextEditor - 富文本編輯器元件
 * 基於 TipTap，提供類似 Gmail 的編輯體驗
 * 支援粗體、斜體、底線、刪除線、連結、對齊、列表、引用等
 */
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Underline from '@tiptap/extension-underline'
import Link from '@tiptap/extension-link'
import TextAlign from '@tiptap/extension-text-align'
import { TextStyle } from '@tiptap/extension-text-style'
import Color from '@tiptap/extension-color'
import Placeholder from '@tiptap/extension-placeholder'
import Image from '@tiptap/extension-image'
import { Table, TableRow, TableCell, TableHeader } from '@tiptap/extension-table'

defineOptions({ inheritAttrs: false })

interface Props {
  /** 輸入值（HTML） */
  modelValue?: string
  /** 佔位符 */
  placeholder?: string
  /** 是否禁用 */
  disabled?: boolean
  /** 編輯器最小高度 */
  minHeight?: string
  /** 是否顯示工具列 */
  toolbar?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  placeholder: '請輸入內容...',
  disabled: false,
  minHeight: '150px',
  toolbar: true,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const editor = useEditor({
  content: props.modelValue,
  editable: !props.disabled,
  extensions: [
    StarterKit.configure({
      heading: { levels: [1, 2, 3] },
    }),
    Underline,
    Link.configure({
      openOnClick: false,
      HTMLAttributes: { target: '_blank', rel: 'noopener noreferrer' },
    }),
    TextAlign.configure({
      types: ['heading', 'paragraph'],
    }),
    TextStyle,
    Color,
    Placeholder.configure({
      placeholder: props.placeholder,
    }),
    Image.configure({
      inline: true,
    }),
    // 表格 - 可拖拉欄寬 + email-safe inline styles
    Table.configure({
      resizable: true,
      HTMLAttributes: {
        style: 'border-collapse: collapse; border: 1px solid #d1d5db;',
      },
    }),
    TableRow.configure({
      HTMLAttributes: {},
    }),
    TableCell.configure({
      HTMLAttributes: {
        style: 'border: 1px solid #d1d5db; padding: 8px 12px; text-align: left; vertical-align: top;',
      },
    }),
    TableHeader.configure({
      HTMLAttributes: {
        style: 'border: 1px solid #d1d5db; padding: 8px 12px; text-align: left; vertical-align: top; background-color: #f3f4f6; font-weight: 600;',
      },
    }),
  ],
  onUpdate: ({ editor: e }) => {
    emit('update:modelValue', e.getHTML())
  },
})

// 監聽外部值變化
watch(() => props.modelValue, (val) => {
  if (editor.value && val !== editor.value.getHTML()) {
    editor.value.commands.setContent(val || '')
  }
})

watch(() => props.disabled, (val) => {
  editor.value?.setEditable(!val)
})

onBeforeUnmount(() => {
  editor.value?.destroy()
})

// 連結操作
const linkUrl = ref('')
const isLinkPopoverOpen = ref(false)

function setLink() {
  if (!linkUrl.value) {
    editor.value?.chain().focus().extendMarkRange('link').unsetLink().run()
  } else {
    const url = linkUrl.value.startsWith('http') ? linkUrl.value : `https://${linkUrl.value}`
    editor.value?.chain().focus().extendMarkRange('link').setLink({ href: url }).run()
  }
  linkUrl.value = ''
  isLinkPopoverOpen.value = false
}

function openLinkPopover() {
  const existing = editor.value?.getAttributes('link').href
  linkUrl.value = existing || ''
  isLinkPopoverOpen.value = true
}

// 文字顏色
const colorOptions = [
  { label: '預設', value: '' },
  { label: '紅色', value: '#dc2626' },
  { label: '橘色', value: '#ea580c' },
  { label: '黃色', value: '#ca8a04' },
  { label: '綠色', value: '#16a34a' },
  { label: '藍色', value: '#2563eb' },
  { label: '紫色', value: '#9333ea' },
  { label: '灰色', value: '#6b7280' },
]

function setColor(color: string) {
  if (!color) {
    editor.value?.chain().focus().unsetColor().run()
  } else {
    editor.value?.chain().focus().setColor(color).run()
  }
}

// 工具列按鈕定義
type ToolbarItem = {
  icon: string
  title: string
  action: () => void
  isActive?: () => boolean
} | { type: 'divider' }

const toolbarGroups = computed<ToolbarItem[][]>(() => {
  if (!editor.value) return []
  const e = editor.value
  return [
    // 文字格式
    [
      { icon: 'i-lucide-bold', title: '粗體', action: () => e.chain().focus().toggleBold().run(), isActive: () => e.isActive('bold') },
      { icon: 'i-lucide-italic', title: '斜體', action: () => e.chain().focus().toggleItalic().run(), isActive: () => e.isActive('italic') },
      { icon: 'i-lucide-underline', title: '底線', action: () => e.chain().focus().toggleUnderline().run(), isActive: () => e.isActive('underline') },
      { icon: 'i-lucide-strikethrough', title: '刪除線', action: () => e.chain().focus().toggleStrike().run(), isActive: () => e.isActive('strike') },
    ],
    // 列表 & 引用
    [
      { icon: 'i-lucide-list', title: '項目符號列表', action: () => e.chain().focus().toggleBulletList().run(), isActive: () => e.isActive('bulletList') },
      { icon: 'i-lucide-list-ordered', title: '編號列表', action: () => e.chain().focus().toggleOrderedList().run(), isActive: () => e.isActive('orderedList') },
      { icon: 'i-lucide-quote', title: '引用', action: () => e.chain().focus().toggleBlockquote().run(), isActive: () => e.isActive('blockquote') },
    ],
    // 對齊
    [
      { icon: 'i-lucide-align-left', title: '靠左', action: () => e.chain().focus().setTextAlign('left').run(), isActive: () => e.isActive({ textAlign: 'left' }) },
      { icon: 'i-lucide-align-center', title: '置中', action: () => e.chain().focus().setTextAlign('center').run(), isActive: () => e.isActive({ textAlign: 'center' }) },
      { icon: 'i-lucide-align-right', title: '靠右', action: () => e.chain().focus().setTextAlign('right').run(), isActive: () => e.isActive({ textAlign: 'right' }) },
    ],
    // 其他
    [
      { icon: 'i-lucide-link', title: '連結', action: () => openLinkPopover(), isActive: () => e.isActive('link') },
      { icon: 'i-lucide-minus', title: '水平線', action: () => e.chain().focus().setHorizontalRule().run() },
      { icon: 'i-lucide-eraser', title: '清除格式', action: () => e.chain().focus().clearNodes().unsetAllMarks().run() },
    ],
  ]
})

// 表格操作選單
const tableMenuItems = computed(() => {
  if (!editor.value) return []
  const e = editor.value
  const inTable = e.isActive('table')
  return [
    [
      { label: '插入表格 (3×3)', icon: 'i-lucide-table', disabled: inTable, click: () => e.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run() },
    ],
    [
      { label: '上方插入列', icon: 'i-lucide-between-vertical-start', disabled: !inTable, click: () => e.chain().focus().addRowBefore().run() },
      { label: '下方插入列', icon: 'i-lucide-between-vertical-end', disabled: !inTable, click: () => e.chain().focus().addRowAfter().run() },
      { label: '左方插入欄', icon: 'i-lucide-between-horizontal-start', disabled: !inTable, click: () => e.chain().focus().addColumnBefore().run() },
      { label: '右方插入欄', icon: 'i-lucide-between-horizontal-end', disabled: !inTable, click: () => e.chain().focus().addColumnAfter().run() },
    ],
    [
      { label: '合併儲存格', icon: 'i-lucide-table-cells-merge', disabled: !inTable, click: () => e.chain().focus().mergeCells().run() },
      { label: '分割儲存格', icon: 'i-lucide-table-cells-split', disabled: !inTable, click: () => e.chain().focus().splitCell().run() },
    ],
    [
      { label: '刪除該列', icon: 'i-lucide-row-delete', disabled: !inTable, click: () => e.chain().focus().deleteRow().run() },
      { label: '刪除該欄', icon: 'i-lucide-column-delete', disabled: !inTable, click: () => e.chain().focus().deleteColumn().run() },
      { label: '刪除整個表格', icon: 'i-lucide-trash-2', color: 'error' as const, disabled: !inTable, click: () => e.chain().focus().deleteTable().run() },
    ],
  ]
})

// 提供方法給外部使用
defineExpose({
  /** 取得 HTML 內容 */
  getHTML: () => editor.value?.getHTML() ?? '',
  /** 取得純文字內容 */
  getText: () => editor.value?.getText() ?? '',
  /** 設定內容 */
  setContent: (content: string) => editor.value?.commands.setContent(content),
  /** 聚焦編輯器 */
  focus: () => editor.value?.commands.focus(),
  /** 清空內容 */
  clear: () => editor.value?.commands.clearContent(),
  /** 取得 editor 實例 */
  editor,
})
</script>

<template>
  <div
    v-bind="$attrs"
    class="base-rich-editor"
    :class="[
      disabled && 'base-rich-editor--disabled',
    ]"
  >
    <!-- 工具列 -->
    <div v-if="toolbar && editor" class="base-rich-editor__toolbar">
      <template v-for="(group, gi) in toolbarGroups" :key="gi">
        <div v-if="gi > 0" class="base-rich-editor__divider" />
        <template v-for="(item, ii) in group" :key="`${gi}-${ii}`">
          <button
            v-if="!('type' in item)"
            type="button"
            class="base-rich-editor__btn"
            :class="{ 'base-rich-editor__btn--active': item.isActive?.() }"
            :title="item.title"
            :disabled="disabled"
            @click="item.action"
          >
            <UIcon :name="item.icon" class="w-4 h-4" />
          </button>
        </template>

        <!-- 文字顏色（緊接在第一組文字格式後面） -->
        <UPopover v-if="gi === 0">
          <button
            type="button"
            class="base-rich-editor__btn"
            title="文字顏色"
            :disabled="disabled"
          >
            <UIcon name="i-lucide-palette" class="w-4 h-4" />
          </button>
          <template #content>
            <div class="p-2 grid grid-cols-4 gap-1">
              <button
                v-for="c in colorOptions"
                :key="c.value"
                type="button"
                class="w-7 h-7 rounded-md border border-gray-200 dark:border-gray-700 hover:scale-110 transition-transform flex items-center justify-center"
                :title="c.label"
                @click="setColor(c.value)"
              >
                <span
                  v-if="c.value"
                  class="w-4 h-4 rounded-full"
                  :style="{ backgroundColor: c.value }"
                />
                <UIcon v-else name="i-lucide-ban" class="w-4 h-4 text-gray-400" />
              </button>
            </div>
          </template>
        </UPopover>
      </template>

      <!-- 表格下拉選單 -->
      <div class="base-rich-editor__divider" />
      <UDropdownMenu :items="tableMenuItems">
        <button
          type="button"
          class="base-rich-editor__btn"
          :class="{ 'base-rich-editor__btn--active': editor?.isActive('table') }"
          title="表格"
          :disabled="disabled"
        >
          <UIcon name="i-lucide-table" class="w-4 h-4" />
        </button>
      </UDropdownMenu>

      <!-- 連結 Popover -->
      <UPopover v-model:open="isLinkPopoverOpen">
        <span />
        <template #content>
          <div class="p-3 w-72 space-y-2">
            <p class="text-sm font-medium text-gray-700 dark:text-gray-300">插入連結</p>
            <UInput
              v-model="linkUrl"
              placeholder="https://example.com"
              size="sm"
              @keydown.enter="setLink"
            />
            <div class="flex justify-end gap-2">
              <UButton size="xs" variant="ghost" color="neutral" @click="isLinkPopoverOpen = false">取消</UButton>
              <UButton size="xs" @click="setLink">
                {{ linkUrl ? '套用' : '移除連結' }}
              </UButton>
            </div>
          </div>
        </template>
      </UPopover>
    </div>

    <!-- 編輯區 -->
    <EditorContent
      :editor="editor"
      class="base-rich-editor__content"
      :style="{ minHeight }"
    />
  </div>
</template>

<style scoped>
.base-rich-editor {
  border: 1px solid var(--ui-border);
  border-radius: var(--ui-radius);
  background: var(--ui-bg);
  overflow: hidden;
  transition: border-color 0.15s;
}

.base-rich-editor:focus-within {
  border-color: var(--ui-primary);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--ui-primary) 20%, transparent);
}

.base-rich-editor--disabled {
  opacity: 0.5;
  pointer-events: none;
}

.base-rich-editor__toolbar {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 4px 8px;
  border-bottom: 1px solid var(--ui-border);
  background: var(--ui-bg-elevated);
  flex-wrap: wrap;
}

.base-rich-editor__divider {
  width: 1px;
  height: 20px;
  background: var(--ui-border);
  margin: 0 4px;
}

.base-rich-editor__btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--ui-radius);
  color: var(--ui-text-muted);
  transition: all 0.15s;
  cursor: pointer;
  border: none;
  background: transparent;
}

.base-rich-editor__btn:hover {
  background: var(--ui-bg-accented);
  color: var(--ui-text);
}

.base-rich-editor__btn--active {
  background: var(--ui-primary);
  color: white;
}

.base-rich-editor__btn--active:hover {
  background: var(--ui-primary);
  color: white;
  opacity: 0.9;
}

.base-rich-editor__content {
  padding: 12px 16px;
  overflow-y: auto;
}

/* TipTap 編輯區內容樣式 */
.base-rich-editor__content :deep(.tiptap) {
  outline: none;
  min-height: inherit;
}

.base-rich-editor__content :deep(.tiptap p) {
  margin: 0 0 0.5em;
}

.base-rich-editor__content :deep(.tiptap p:last-child) {
  margin-bottom: 0;
}

.base-rich-editor__content :deep(.tiptap h1) {
  font-size: 1.5em;
  font-weight: 700;
  margin: 0 0 0.5em;
}

.base-rich-editor__content :deep(.tiptap h2) {
  font-size: 1.25em;
  font-weight: 600;
  margin: 0 0 0.5em;
}

.base-rich-editor__content :deep(.tiptap h3) {
  font-size: 1.1em;
  font-weight: 600;
  margin: 0 0 0.5em;
}

.base-rich-editor__content :deep(.tiptap ul) {
  list-style: disc;
  padding-left: 1.5em;
  margin: 0 0 0.5em;
}

.base-rich-editor__content :deep(.tiptap ol) {
  list-style: decimal;
  padding-left: 1.5em;
  margin: 0 0 0.5em;
}

.base-rich-editor__content :deep(.tiptap li) {
  margin: 0.15em 0;
}

.base-rich-editor__content :deep(.tiptap blockquote) {
  border-left: 3px solid var(--ui-border);
  padding-left: 1em;
  margin: 0.5em 0;
  color: var(--ui-text-muted);
}

.base-rich-editor__content :deep(.tiptap a) {
  color: var(--ui-primary);
  text-decoration: underline;
  cursor: pointer;
}

.base-rich-editor__content :deep(.tiptap hr) {
  border: none;
  border-top: 1px solid var(--ui-border);
  margin: 1em 0;
}

.base-rich-editor__content :deep(.tiptap code) {
  background: var(--ui-bg-accented);
  padding: 0.15em 0.4em;
  border-radius: 4px;
  font-size: 0.9em;
}

.base-rich-editor__content :deep(.tiptap pre) {
  background: var(--ui-bg-accented);
  padding: 0.75em 1em;
  border-radius: var(--ui-radius);
  overflow-x: auto;
  margin: 0.5em 0;
}

.base-rich-editor__content :deep(.tiptap pre code) {
  background: none;
  padding: 0;
}

.base-rich-editor__content :deep(.tiptap img) {
  max-width: 100%;
  height: auto;
  border-radius: var(--ui-radius);
}

/* 表格樣式（編輯器內顯示用） */
.base-rich-editor__content :deep(.tiptap table) {
  border-collapse: collapse;
  margin: 0.5em 0;
  overflow: hidden;
}

.base-rich-editor__content :deep(.tiptap th),
.base-rich-editor__content :deep(.tiptap td) {
  border: 1px solid var(--ui-border);
  padding: 8px 12px;
  text-align: left;
  vertical-align: top;
  min-width: 80px;
  position: relative;
}

.base-rich-editor__content :deep(.tiptap th) {
  background: var(--ui-bg-elevated);
  font-weight: 600;
}

.base-rich-editor__content :deep(.tiptap td p),
.base-rich-editor__content :deep(.tiptap th p) {
  margin: 0;
}

/* 選中儲存格高亮 */
.base-rich-editor__content :deep(.tiptap .selectedCell::after) {
  content: '';
  position: absolute;
  inset: 0;
  background: color-mix(in srgb, var(--ui-primary) 15%, transparent);
  pointer-events: none;
}

/* 欄寬拖拉手把 */
.base-rich-editor__content :deep(.column-resize-handle) {
  position: absolute;
  right: -2px;
  top: 0;
  bottom: -2px;
  width: 4px;
  background: var(--ui-primary);
  cursor: col-resize;
  z-index: 10;
}

.base-rich-editor__content :deep(.tableWrapper) {
  overflow-x: auto;
  margin: 0.5em 0;
}

.base-rich-editor__content :deep(.resize-cursor) {
  cursor: col-resize;
}

/* Placeholder 樣式 */
.base-rich-editor__content :deep(.tiptap p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  float: left;
  color: var(--ui-text-dimmed);
  pointer-events: none;
  height: 0;
}
</style>
