<script setup lang="ts">
import { h, resolveComponent } from 'vue'
import type { TableColumn } from '@nuxt/ui'
import type { Email } from '~/stores/emails'

definePageMeta({
  middleware: 'auth'
})

const UBadge = resolveComponent('UBadge')
const UButton = resolveComponent('UButton')
const UDropdownMenu = resolveComponent('UDropdownMenu')
const UCheckbox = resolveComponent('UCheckbox')

const emailsStore = useEmailsStore()
const router = useRouter()
const toast = useToast()

// 載入郵件和 Gmail 狀態
onMounted(async () => {
  await Promise.all([
    emailsStore.fetchEmails(),
    emailsStore.fetchGmailStatus()
  ])
})

// Table columns 定義
const columns: TableColumn<Email>[] = [{
  id: 'select',
  header: ({ table }) => h(UCheckbox, {
    'modelValue': table.getIsSomePageRowsSelected() ? 'indeterminate' : table.getIsAllPageRowsSelected(),
    'onUpdate:modelValue': (value: boolean | 'indeterminate') => table.toggleAllPageRowsSelected(!!value),
    'aria-label': 'Select all'
  }),
  cell: ({ row }) => h(UCheckbox, {
    'modelValue': row.getIsSelected(),
    'onUpdate:modelValue': (value: boolean | 'indeterminate') => row.toggleSelected(!!value),
    'aria-label': 'Select row'
  }),
  enableSorting: false,
  enableHiding: false
}, {
  accessorKey: 'from_name',
  header: '寄件者',
  cell: ({ row }) => {
    const fromName = row.original.from_name || row.original.from_email
    const isUnread = !row.original.is_read
    
    return h('div', { class: 'flex items-center gap-2' }, [
      h('div', { class: 'flex flex-col' }, [
        h('span', { 
          class: isUnread ? 'font-semibold text-highlighted' : 'text-muted'
        }, fromName),
        h('span', { class: 'text-sm text-muted' }, row.original.from_email)
      ])
    ])
  }
}, {
  accessorKey: 'subject',
  header: '主旨',
  cell: ({ row }) => {
    const subject = row.original.subject || '(無主旨)'
    const snippet = row.original.snippet || ''
    const isUnread = !row.original.is_read
    
    return h('div', { class: 'flex flex-col gap-1 max-w-md' }, [
      h('span', { 
        class: isUnread ? 'font-semibold text-highlighted' : 'text-muted',
      }, subject),
      h('span', { 
        class: 'text-sm text-muted truncate'
      }, snippet)
    ])
  }
}, {
  accessorKey: 'received_at',
  header: '時間',
  cell: ({ row }) => {
    const date = new Date(row.getValue('received_at'))
    const now = new Date()
    const diffDays = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24))
    
    let formatted = ''
    if (diffDays === 0) {
      formatted = date.toLocaleTimeString('zh-TW', { hour: '2-digit', minute: '2-digit' })
    } else if (diffDays < 7) {
      formatted = date.toLocaleDateString('zh-TW', { weekday: 'short' })
    } else {
      formatted = date.toLocaleDateString('zh-TW', { month: 'short', day: 'numeric' })
    }
    
    return h('span', { class: 'text-sm text-muted' }, formatted)
  }
}, {
  id: 'labels',
  header: '標籤',
  cell: ({ row }) => {
    const labels = row.original.labels || []
    const displayLabels = labels.filter(l => 
      !['INBOX', 'UNREAD', 'IMPORTANT', 'STARRED'].includes(l) &&
      !l.startsWith('CATEGORY_')
    ).slice(0, 2)
    
    if (displayLabels.length === 0) return null
    
    return h('div', { class: 'flex gap-1' }, 
      displayLabels.map(label => 
        h(UBadge, { 
          size: 'xs',
          variant: 'subtle',
          color: 'neutral'
        }, () => label)
      )
    )
  }
}, {
  id: 'status',
  header: '狀態',
  cell: ({ row }) => {
    const badges = []
    
    if (row.original.has_attachments) {
      badges.push(h(UBadge, {
        size: 'xs',
        variant: 'subtle',
        color: 'neutral',
        icon: 'i-lucide-paperclip'
      }))
    }
    
    if (row.original.labels?.includes('STARRED')) {
      badges.push(h(UBadge, {
        size: 'xs',
        variant: 'subtle',
        color: 'warning',
        icon: 'i-lucide-star'
      }))
    }
    
    if (row.original.case_id) {
      badges.push(h(UBadge, {
        size: 'xs',
        variant: 'subtle',
        color: 'primary'
      }, () => '已關聯案件'))
    }
    
    if (row.original.ai_analyzed) {
      badges.push(h(UBadge, {
        size: 'xs',
        variant: 'subtle',
        color: 'success',
        icon: 'i-lucide-sparkles'
      }))
    }
    
    return h('div', { class: 'flex gap-1' }, badges)
  }
}, {
  id: 'actions',
  enableHiding: false,
  cell: ({ row }) => {
    const items = [{
      type: 'label' as const,
      label: '操作'
    }, {
      label: row.original.is_read ? '標記為未讀' : '標記為已讀',
      icon: row.original.is_read ? 'i-lucide-mail' : 'i-lucide-mail-open',
      async onSelect() {
        try {
          await emailsStore.markAsRead(row.original.id, !row.original.is_read)
          toast.add({
            title: row.original.is_read ? '已標記為未讀' : '已標記為已讀',
            color: 'success'
          })
        } catch (e) {
          toast.add({
            title: '操作失敗',
            color: 'error'
          })
        }
      }
    }, {
      label: '查看詳情',
      icon: 'i-lucide-eye',
      onSelect() {
        router.push(`/emails/${row.original.id}`)
      }
    }, {
      type: 'separator' as const
    }, {
      label: '關聯到案件',
      icon: 'i-lucide-link',
      onSelect() {
        selectedEmailForLink.value = row.original
        showLinkDialog.value = true
      }
    }]

    return h('div', { class: 'text-right' }, h(UDropdownMenu, {
      'content': { align: 'end' },
      items,
      'aria-label': 'Actions dropdown'
    }, () => h(UButton, {
      'icon': 'i-lucide-ellipsis-vertical',
      'color': 'neutral',
      'variant': 'ghost',
      'class': 'ml-auto',
      'aria-label': 'Actions dropdown'
    })))
  }
}]

const table = useTemplateRef('table')
const rowSelection = ref<Record<string, boolean>>({})

// 篩選相關
const filterIsRead = ref<'all' | 'read' | 'unread'>('all')
const searchQuery = ref('')

// 關聯案件對話框
const showLinkDialog = ref(false)
const selectedEmailForLink = ref<Email | null>(null)

const handleLinked = async (caseId: string) => {
  // 重新載入郵件列表
  await emailsStore.fetchEmails()
  toast.add({
    title: '案件關聯成功',
    color: 'success'
  })
}

// 監聽篩選變更
watch([filterIsRead, searchQuery], async () => {
  const params: any = {}
  
  if (filterIsRead.value !== 'all') {
    params.is_read = filterIsRead.value === 'read'
  }
  
  if (searchQuery.value) {
    params.subject = searchQuery.value
  }
  
  await emailsStore.fetchEmails(params)
})

// 刷新郵件列表
const refreshEmails = async () => {
  await emailsStore.fetchEmails()
  toast.add({
    title: '郵件列表已更新',
    color: 'success'
  })
}

// 觸發同步
const handleSync = async () => {
  try {
    await emailsStore.triggerSync()
    toast.add({
      title: '開始同步郵件',
      description: '這可能需要幾分鐘，請稍候',
      color: 'info'
    })
  } catch (e) {
    // error 已經在 store 中處理
  }
}

// 點擊郵件行
function onSelectRow(e: Event, row: any) {
  router.push(`/emails/${row.original.id}`)
}
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Header -->
    <div class="flex items-center justify-between px-6 py-4 border-b border-default">
      <div class="flex items-center gap-4">
        <h1 class="text-2xl font-bold text-highlighted">郵件</h1>
        
        <!-- Gmail 狀態指示器 -->
        <UBadge 
          v-if="emailsStore.isConnected"
          :color="emailsStore.gmailStatus?.sync_status === 'active' ? 'success' : 'warning'"
          variant="subtle"
        >
          <template #leading>
            <span class="w-2 h-2 rounded-full" 
              :class="{
                'bg-success animate-pulse': emailsStore.gmailStatus?.sync_status === 'active',
                'bg-warning': emailsStore.gmailStatus?.sync_status !== 'active'
              }"
            />
          </template>
          {{ emailsStore.gmailStatus?.email }}
        </UBadge>
      </div>

      <div class="flex items-center gap-2">
        <!-- 同步按鈕 -->
        <UButton
          v-if="emailsStore.isConnected"
          icon="i-lucide-refresh-cw"
          color="neutral"
          variant="outline"
          :loading="emailsStore.syncing"
          :disabled="!emailsStore.canSync"
          @click="handleSync"
        >
          {{ emailsStore.syncing ? '同步中...' : '同步郵件' }}
        </UButton>

        <!-- 重新整理 -->
        <UButton
          icon="i-lucide-rotate-cw"
          color="neutral"
          variant="ghost"
          :loading="emailsStore.loading"
          @click="refreshEmails"
          aria-label="重新整理"
        />
      </div>
    </div>

    <!-- Filters -->
    <div class="flex items-center gap-2 px-6 py-3 border-b border-default bg-elevated/50">
      <!-- 搜尋框 -->
      <UInput
        v-model="searchQuery"
        icon="i-lucide-search"
        placeholder="搜尋主旨..."
        class="max-w-sm"
      />

      <!-- 已讀/未讀篩選 -->
      <USelectMenu
        v-model="filterIsRead"
        :options="[
          { value: 'all', label: '全部' },
          { value: 'unread', label: '未讀' },
          { value: 'read', label: '已讀' }
        ]"
        value-attribute="value"
        option-attribute="label"
        class="w-32"
      />

      <!-- 統計資訊 -->
      <div class="ml-auto text-sm text-muted">
        未讀: <span class="font-semibold text-highlighted">{{ emailsStore.unreadCount }}</span> / 
        總計: <span class="font-semibold text-highlighted">{{ emailsStore.pagination.total }}</span>
      </div>
    </div>

    <!-- Table -->
    <div class="flex-1 overflow-hidden">
      <UTable
        ref="table"
        v-model:row-selection="rowSelection"
        :data="emailsStore.emails"
        :columns="columns"
        :loading="emailsStore.loading"
        sticky
        class="h-full"
        @select="onSelectRow"
      />
    </div>

    <!-- Pagination -->
    <div v-if="emailsStore.pagination.total_pages > 1" 
      class="flex justify-center px-6 py-4 border-t border-default"
    >
      <UPagination
        :model-value="emailsStore.pagination.page"
        :items-per-page="emailsStore.pagination.page_size"
        :total="emailsStore.pagination.total"
        @update:model-value="(page) => emailsStore.fetchEmails({ page })"
      />
    </div>

    <!-- Empty state -->
    <div 
      v-if="!emailsStore.loading && emailsStore.emails.length === 0" 
      class="flex flex-col items-center justify-center h-full py-12"
    >
      <div class="text-center">
        <UIcon name="i-lucide-inbox" class="w-16 h-16 mx-auto mb-4 text-muted" />
        
        <h3 class="text-lg font-semibold text-highlighted mb-2">
          {{ emailsStore.isConnected ? '目前沒有郵件' : '尚未連接 Gmail' }}
        </h3>
        
        <p class="text-muted mb-4">
          {{ emailsStore.isConnected 
            ? '嘗試同步郵件或調整篩選條件' 
            : '請先在設定中連接您的 Gmail 帳號' 
          }}
        </p>
        
        <UButton
          v-if="emailsStore.isConnected"
          icon="i-lucide-refresh-cw"
          color="primary"
          @click="handleSync"
        >
          立即同步
        </UButton>
        
        <UButton
          v-else
          icon="i-lucide-settings"
          color="primary"
          to="/settings"
        >
          前往設定
        </UButton>
      </div>
    </div>

    <!-- Link to Case Dialog -->
    <EmailLinkToCaseDialog
      v-if="selectedEmailForLink"
      :email-id="selectedEmailForLink.id"
      :open="showLinkDialog"
      @close="showLinkDialog = false"
      @linked="handleLinked"
    />
  </div>
</template>

