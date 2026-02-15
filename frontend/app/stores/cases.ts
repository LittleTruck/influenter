import { defineStore } from 'pinia'
import { logError } from '~/utils/errorUtils'
import { casesStorage, tasksStorage, generateTempId } from '~/utils/localStorage'
import type {
  Case,
  CaseDetail,
  CaseQueryParams,
  CaseListResponse,
  CreateCaseRequest,
  UpdateCaseRequest,
  Task,
  CreateTaskRequest,
  UpdateTaskRequest,
  ReorderTasksRequest
} from '~/types/cases'

export const useCasesStore = defineStore('cases', () => {
  // State
  const cases: Ref<Case[]> = ref([])
  const currentCase: Ref<CaseDetail | null> = ref(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Pagination
  const pagination = ref({
    page: 1,
    per_page: 20,
    total: 0,
    total_pages: 0
  })

  // Filters
  const filters = ref<CaseQueryParams>({
    page: 1,
    per_page: 20,
    sort: 'updated_at_desc'
  })

  // Current view (board or list)
  const currentView = ref<'board' | 'list'>('board')

  // Actions
  const fetchCases = async (params?: Partial<CaseQueryParams>) => {
    loading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      // 合併篩選參數，並過濾掉 undefined 值
      const mergedParams = { ...filters.value, ...params }

      const queryParams: Record<string, string | number> = {}
      Object.keys(mergedParams).forEach(key => {
        const value = mergedParams[key as keyof CaseQueryParams]
        if (value !== undefined) {
          queryParams[key] = value
        }
      })

      const data = await $fetch<CaseListResponse>(`${config.public.apiBase}/api/v1/cases`, {
        method: 'GET',
        params: queryParams,
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      cases.value = data.data
      pagination.value = data.pagination
      filters.value = { ...queryParams }
      
      // 同步到 localStorage
      casesStorage.setCases(data.data)
    } catch (e: unknown) {
      error.value = logError(e, '取得案件列表失敗', { component: 'casesStore', action: 'fetchCases' })
      // Fallback 到 localStorage
      const localCases = casesStorage.getCases()
      if (localCases.length > 0) {
        cases.value = localCases as Case[]
        pagination.value = {
          page: 1,
          per_page: 20,
          total: localCases.length,
          total_pages: Math.ceil(localCases.length / 20)
        }
      }
    } finally {
      loading.value = false
    }
  }

  const fetchCase = async (id: string) => {
    loading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const data = await $fetch<CaseDetail>(`${config.public.apiBase}/api/v1/cases/${id}`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      currentCase.value = data

      // 同時更新列表中的案件（如果存在）
      const index = cases.value.findIndex(c => c.id === id)
      if (index !== -1) {
        cases.value[index] = { ...cases.value[index], ...data }
      }

      // 同步到 localStorage
      casesStorage.setCaseDetail(id, data)
      casesStorage.updateCase(id, data)

      return data
    } catch (e: unknown) {
      error.value = logError(e, '取得案件詳情失敗', { component: 'casesStore', action: 'fetchCase' })
      // Fallback 到 localStorage
      const localDetail = casesStorage.getCaseDetail(id)
      if (localDetail) {
        currentCase.value = localDetail as CaseDetail
        return localDetail as CaseDetail
      }
      return null
    } finally {
      loading.value = false
    }
  }

  const createCase = async (data: CreateCaseRequest) => {
    loading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const newCase = await $fetch<Case>(`${config.public.apiBase}/api/v1/cases`, {
        method: 'POST',
        body: data,
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      // 添加到列表開頭
      cases.value.unshift(newCase)
      pagination.value.total += 1

      // 同步到 localStorage
      casesStorage.addCase(newCase)
      casesStorage.setCaseDetail(newCase.id, { ...newCase, tasks: [], emails: [] })

      return newCase
    } catch (e: unknown) {
      error.value = logError(e, '建立案件失敗（已儲存到本地）', { component: 'casesStore', action: 'createCase' })
      // Fallback 到 localStorage：建立臨時案件
      const tempId = generateTempId()
      const now = new Date().toISOString()
      const newCase: Case = {
        id: tempId,
        ...data,
        status: data.status || 'to_confirm',
        created_at: now,
        updated_at: now,
        email_count: 0,
        task_count: 0,
        completed_task_count: 0
      } as Case

      // 添加到列表開頭
      cases.value.unshift(newCase)
      pagination.value.total += 1

      // 儲存到 localStorage
      casesStorage.addCase(newCase)
      casesStorage.setCaseDetail(tempId, { ...newCase, tasks: [], emails: [] } as CaseDetail)

      return newCase
    } finally {
      loading.value = false
    }
  }

  const updateCase = async (id: string, data: UpdateCaseRequest) => {
    loading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const updatedCase = await $fetch<Case>(`${config.public.apiBase}/api/v1/cases/${id}`, {
        method: 'PATCH',
        body: data,
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      // 更新列表中的案件
      const index = cases.value.findIndex(c => c.id === id)
      if (index !== -1) {
        cases.value[index] = updatedCase
      }

      // 更新當前案件
      if (currentCase.value?.id === id) {
        currentCase.value = { ...currentCase.value, ...updatedCase }
      }

      // 同步到 localStorage
      casesStorage.updateCase(id, updatedCase)
      const detail = casesStorage.getCaseDetail(id)
      if (detail) {
        casesStorage.setCaseDetail(id, { ...detail, ...updatedCase })
      }

      return updatedCase
    } catch (e: unknown) {
      error.value = logError(e, '更新案件失敗（已儲存到本地）', { component: 'casesStore', action: 'updateCase' })
      // Fallback 到 localStorage：更新本地資料
      const index = cases.value.findIndex(c => c.id === id)
      if (index !== -1) {
        const updated = { ...cases.value[index], ...data, updated_at: new Date().toISOString() }
        cases.value[index] = updated as Case
        casesStorage.updateCase(id, updated)
      }

      if (currentCase.value?.id === id) {
        currentCase.value = { ...currentCase.value, ...data, updated_at: new Date().toISOString() } as CaseDetail
        casesStorage.setCaseDetail(id, currentCase.value)
      }

      return cases.value.find(c => c.id === id) as Case
    } finally {
      loading.value = false
    }
  }

  const updateCaseStatus = async (id: string, status: Case['status']) => {
    return updateCase(id, { status })
  }

  const deleteCase = async (id: string) => {
    loading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      await $fetch(`${config.public.apiBase}/api/v1/cases/${id}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      // 從列表中移除
      cases.value = cases.value.filter(c => c.id !== id)
      pagination.value.total -= 1

      // 如果刪除的是當前案件，清空
      if (currentCase.value?.id === id) {
        currentCase.value = null
      }

      // 從 localStorage 刪除
      casesStorage.deleteCase(id)
    } catch (e: unknown) {
      error.value = logError(e, '刪除案件失敗（已從本地移除）', { component: 'casesStore', action: 'deleteCase' })
      // Fallback 到 localStorage：從本地刪除
      cases.value = cases.value.filter(c => c.id !== id)
      pagination.value.total = Math.max(0, pagination.value.total - 1)
      if (currentCase.value?.id === id) {
        currentCase.value = null
      }
      casesStorage.deleteCase(id)
    } finally {
      loading.value = false
    }
  }

  const linkEmailToCase = async (caseId: string, emailId: string) => {
    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      await $fetch(`${config.public.apiBase}/api/v1/cases/${caseId}/emails`, {
        method: 'POST',
        body: { email_id: emailId },
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      // 重新載入案件詳情以更新郵件列表
      if (currentCase.value?.id === caseId) {
        await fetchCase(caseId)
      }
    } catch (e: unknown) {
      error.value = logError(e, '關聯郵件失敗', { component: 'casesStore', action: 'linkEmailToCase' })
      throw e
    }
  }

  const unlinkEmailFromCase = async (caseId: string, emailId: string) => {
    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      await $fetch(`${config.public.apiBase}/api/v1/cases/${caseId}/emails/${emailId}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      // 重新載入案件詳情
      if (currentCase.value?.id === caseId) {
        await fetchCase(caseId)
      }
    } catch (e: unknown) {
      error.value = logError(e, '取消關聯郵件失敗', { component: 'casesStore', action: 'unlinkEmailFromCase' })
      throw e
    }
  }

  const fetchCaseEmails = async (caseId: string) => {
    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const data = await $fetch<{ data: CaseDetail['emails'] }>(
        `${config.public.apiBase}/api/v1/cases/${caseId}/emails`,
        {
          method: 'GET',
          headers: {
            Authorization: `Bearer ${authStore.token}`
          }
        }
      )

      // 更新當前案件的郵件列表
      if (currentCase.value?.id === caseId) {
        currentCase.value.emails = data.data
      }

      return data.data
    } catch (e: unknown) {
      error.value = logError(e, '取得案件郵件失敗', { component: 'casesStore', action: 'fetchCaseEmails' })
      return []
    }
  }

  const fetchCaseTasks = async (caseId: string) => {
    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const data = await $fetch<{ data: Task[] }>(
        `${config.public.apiBase}/api/v1/cases/${caseId}/tasks`,
        {
          method: 'GET',
          headers: {
            Authorization: `Bearer ${authStore.token}`
          }
        }
      )

      // 更新當前案件的任務列表
      if (currentCase.value?.id === caseId) {
        currentCase.value.tasks = data.data
      }

      // 同步到 localStorage
      tasksStorage.setTasks(caseId, data.data)

      return data.data
    } catch (e: unknown) {
      error.value = logError(e, '取得案件任務失敗', { component: 'casesStore', action: 'fetchCaseTasks' })
      // Fallback 到 localStorage
      const localTasks = tasksStorage.getTasks(caseId)
      if (currentCase.value?.id === caseId) {
        currentCase.value.tasks = localTasks as Task[]
      }
      return localTasks as Task[]
    }
  }

  const createTask = async (caseId: string, data: CreateTaskRequest) => {
    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const newTask = await $fetch<Task>(`${config.public.apiBase}/api/v1/cases/${caseId}/tasks`, {
        method: 'POST',
        body: data,
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      // 更新當前案件的任務列表
      if (currentCase.value?.id === caseId) {
        if (!currentCase.value.tasks) {
          currentCase.value.tasks = []
        }
        currentCase.value.tasks.push(newTask)
      }

      // 同步到 localStorage
      tasksStorage.addTask(caseId, newTask)

      return newTask
    } catch (e: unknown) {
      error.value = logError(e, '建立任務失敗（已儲存到本地）', { component: 'casesStore', action: 'createTask' })
      // Fallback 到 localStorage：建立臨時任務
      const tempId = generateTempId()
      const now = new Date().toISOString()
      const newTask: Task = {
        id: tempId,
        case_id: caseId,
        ...data,
        is_completed: false,
        order: 0,
        created_at: now,
        updated_at: now
      } as Task

      if (currentCase.value?.id === caseId) {
        if (!currentCase.value.tasks) {
          currentCase.value.tasks = []
        }
        currentCase.value.tasks.push(newTask)
      }

      tasksStorage.addTask(caseId, newTask)

      return newTask
    }
  }

  const updateTask = async (taskId: string, data: UpdateTaskRequest) => {
    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const updatedTask = await $fetch<Task>(`${config.public.apiBase}/api/v1/tasks/${taskId}`, {
        method: 'PATCH',
        body: data,
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      // 更新當前案件的任務列表
      if (currentCase.value?.tasks) {
        const index = currentCase.value.tasks.findIndex(t => t.id === taskId)
        if (index !== -1) {
          currentCase.value.tasks[index] = updatedTask
        }
      }

      // 同步到 localStorage
      if (currentCase.value?.id) {
        tasksStorage.updateTask(currentCase.value.id, taskId, updatedTask)
      }

      return updatedTask
    } catch (e: unknown) {
      error.value = logError(e, '更新任務失敗（已儲存到本地）', { component: 'casesStore', action: 'updateTask' })
      // Fallback 到 localStorage
      if (currentCase.value?.tasks) {
        const index = currentCase.value.tasks.findIndex(t => t.id === taskId)
        if (index !== -1) {
          const updated = { ...currentCase.value.tasks[index], ...data, updated_at: new Date().toISOString() }
          currentCase.value.tasks[index] = updated as Task
          if (currentCase.value.id) {
            tasksStorage.updateTask(currentCase.value.id, taskId, updated)
          }
          return updated as Task
        }
      }
      throw e
    }
  }

  const deleteTask = async (taskId: string) => {
    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      await $fetch(`${config.public.apiBase}/api/v1/tasks/${taskId}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      // 從當前案件的任務列表中移除
      if (currentCase.value?.tasks) {
        currentCase.value.tasks = currentCase.value.tasks.filter(t => t.id !== taskId)
      }

      // 從 localStorage 刪除
      if (currentCase.value?.id) {
        tasksStorage.deleteTask(currentCase.value.id, taskId)
      }
    } catch (e: unknown) {
      error.value = logError(e, '刪除任務失敗（已從本地移除）', { component: 'casesStore', action: 'deleteTask' })
      // Fallback 到 localStorage
      if (currentCase.value?.tasks) {
        currentCase.value.tasks = currentCase.value.tasks.filter(t => t.id !== taskId)
      }
      if (currentCase.value?.id) {
        tasksStorage.deleteTask(currentCase.value.id, taskId)
      }
    }
  }

  const completeTask = async (taskId: string) => {
    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const completedTask = await $fetch<Task>(
        `${config.public.apiBase}/api/v1/tasks/${taskId}/complete`,
        {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${authStore.token}`
          }
        }
      )

      // 更新當前案件的任務列表
      if (currentCase.value?.tasks) {
        const index = currentCase.value.tasks.findIndex(t => t.id === taskId)
        if (index !== -1) {
          currentCase.value.tasks[index] = completedTask
        }
      }

      // 同步到 localStorage
      if (currentCase.value?.id) {
        tasksStorage.updateTask(currentCase.value.id, taskId, completedTask)
      }

      return completedTask
    } catch (e: unknown) {
      error.value = logError(e, '完成任務失敗（已儲存到本地）', { component: 'casesStore', action: 'completeTask' })
      // Fallback 到 localStorage
      if (currentCase.value?.tasks) {
        const index = currentCase.value.tasks.findIndex(t => t.id === taskId)
        if (index !== -1) {
          const updated = { ...currentCase.value.tasks[index], is_completed: true, updated_at: new Date().toISOString() }
          currentCase.value.tasks[index] = updated as Task
          if (currentCase.value.id) {
            tasksStorage.updateTask(currentCase.value.id, taskId, updated)
          }
          return updated as Task
        }
      }
      throw e
    }
  }

  const reorderTasks = async (caseId: string, taskIds: string[]) => {
    try {
      const config = useRuntimeConfig()
      const authStore = useAuthStore()

      const data: ReorderTasksRequest = { task_ids: taskIds }

      await $fetch(`${config.public.apiBase}/api/v1/cases/${caseId}/tasks/reorder`, {
        method: 'PATCH',
        body: data,
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      // 重新載入任務列表以獲取更新後的順序
      await fetchCaseTasks(caseId)
    } catch (e: unknown) {
      error.value = logError(e, '重新排序任務失敗（已儲存到本地）', { component: 'casesStore', action: 'reorderTasks' })
      // Fallback 到 localStorage：重新排序本地任務
      const tasks = tasksStorage.getTasks(caseId)
      const reordered = taskIds.map(id => tasks.find((t: any) => t.id === id)).filter(Boolean)
      tasksStorage.setTasks(caseId, reordered)
      
      // 更新當前案件的任務列表
      if (currentCase.value?.id === caseId) {
        currentCase.value.tasks = reordered as Task[]
      }
    }
  }

  // Getters
  const hasCases = computed(() => {
    return cases.value.length > 0
  })

  const casesByStatus = computed(() => {
    const grouped: Record<Case['status'], Case[]> = {
      to_confirm: [],
      in_progress: [],
      completed: [],
      cancelled: [],
      other: []
    }

    cases.value.forEach(c => {
      if (grouped[c.status]) {
        grouped[c.status].push(c)
      }
    })

    return grouped
  })

  // 重置狀態
  const reset = () => {
    cases.value = []
    currentCase.value = null
    loading.value = false
    error.value = null
    pagination.value = {
      page: 1,
      per_page: 20,
      total: 0,
      total_pages: 0
    }
    filters.value = {
      page: 1,
      per_page: 20,
      sort: 'updated_at_desc'
    }
  }

  return {
    // State
    cases,
    currentCase,
    loading,
    error,
    pagination,
    filters,
    currentView,

    // Getters
    hasCases,
    casesByStatus,

    // Actions
    fetchCases,
    fetchCase,
    createCase,
    updateCase,
    updateCaseStatus,
    deleteCase,
    linkEmailToCase,
    unlinkEmailFromCase,
    fetchCaseEmails,
    fetchCaseTasks,
    createTask,
    updateTask,
    deleteTask,
    completeTask,
    reorderTasks,
    reset
  }
})

