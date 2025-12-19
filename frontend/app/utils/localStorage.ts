/**
 * localStorage 工具函數
 * 用於在前端持久化數據，當後端 API 不可用時作為 fallback
 */

const STORAGE_KEYS = {
  CASES: 'influenter_cases',
  CASE_DETAILS: 'influenter_case_details',
  CASE_FIELDS: 'influenter_case_fields',
  TASKS: 'influenter_tasks',
  COLLABORATION_ITEMS: 'influenter_collaboration_items'
} as const

/**
 * 生成臨時 ID（用於前端建立的資料）
 */
export const generateTempId = (): string => {
  return `temp_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
}

/**
 * 檢查是否為臨時 ID
 */
export const isTempId = (id: string): boolean => {
  return id.startsWith('temp_')
}

/**
 * 案件相關的 localStorage 操作
 */
export const casesStorage = {
  /**
   * 取得所有案件
   */
  getCases(): any[] {
    if (typeof window === 'undefined') return []
    try {
      const data = localStorage.getItem(STORAGE_KEYS.CASES)
      return data ? JSON.parse(data) : []
    } catch {
      return []
    }
  },

  /**
   * 儲存案件列表
   */
  setCases(cases: any[]): void {
    if (typeof window === 'undefined') return
    try {
      localStorage.setItem(STORAGE_KEYS.CASES, JSON.stringify(cases))
    } catch (error) {
      console.error('Failed to save cases to localStorage:', error)
    }
  },

  /**
   * 取得單一案件詳情
   */
  getCaseDetail(id: string): any | null {
    if (typeof window === 'undefined') return null
    try {
      const data = localStorage.getItem(STORAGE_KEYS.CASE_DETAILS)
      const details = data ? JSON.parse(data) : {}
      return details[id] || null
    } catch {
      return null
    }
  },

  /**
   * 儲存案件詳情
   */
  setCaseDetail(id: string, detail: any): void {
    if (typeof window === 'undefined') return
    try {
      const data = localStorage.getItem(STORAGE_KEYS.CASE_DETAILS)
      const details = data ? JSON.parse(data) : {}
      details[id] = detail
      localStorage.setItem(STORAGE_KEYS.CASE_DETAILS, JSON.stringify(details))
    } catch (error) {
      console.error('Failed to save case detail to localStorage:', error)
    }
  },

  /**
   * 刪除案件詳情
   */
  deleteCaseDetail(id: string): void {
    if (typeof window === 'undefined') return
    try {
      const data = localStorage.getItem(STORAGE_KEYS.CASE_DETAILS)
      const details = data ? JSON.parse(data) : {}
      delete details[id]
      localStorage.setItem(STORAGE_KEYS.CASE_DETAILS, JSON.stringify(details))
    } catch (error) {
      console.error('Failed to delete case detail from localStorage:', error)
    }
  },

  /**
   * 新增案件到列表
   */
  addCase(caseData: any): void {
    const cases = this.getCases()
    cases.unshift(caseData)
    this.setCases(cases)
  },

  /**
   * 更新案件
   */
  updateCase(id: string, updates: any): void {
    const cases = this.getCases()
    const index = cases.findIndex((c: any) => c.id === id)
    if (index !== -1) {
      cases[index] = { ...cases[index], ...updates }
      this.setCases(cases)
    }

    // 同時更新詳情
    const detail = this.getCaseDetail(id)
    if (detail) {
      this.setCaseDetail(id, { ...detail, ...updates })
    }
  },

  /**
   * 刪除案件
   */
  deleteCase(id: string): void {
    const cases = this.getCases()
    const filtered = cases.filter((c: any) => c.id !== id)
    this.setCases(filtered)
    this.deleteCaseDetail(id)
  }
}

/**
 * 任務相關的 localStorage 操作
 */
export const tasksStorage = {
  /**
   * 取得案件的所有任務
   */
  getTasks(caseId: string): any[] {
    if (typeof window === 'undefined') return []
    try {
      const data = localStorage.getItem(STORAGE_KEYS.TASKS)
      const tasks = data ? JSON.parse(data) : {}
      return tasks[caseId] || []
    } catch {
      return []
    }
  },

  /**
   * 儲存任務列表
   */
  setTasks(caseId: string, tasks: any[]): void {
    if (typeof window === 'undefined') return
    try {
      const data = localStorage.getItem(STORAGE_KEYS.TASKS)
      const allTasks = data ? JSON.parse(data) : {}
      allTasks[caseId] = tasks
      localStorage.setItem(STORAGE_KEYS.TASKS, JSON.stringify(allTasks))
    } catch (error) {
      console.error('Failed to save tasks to localStorage:', error)
    }
  },

  /**
   * 新增任務
   */
  addTask(caseId: string, task: any): void {
    const tasks = this.getTasks(caseId)
    tasks.push(task)
    this.setTasks(caseId, tasks)
  },

  /**
   * 更新任務
   */
  updateTask(caseId: string, taskId: string, updates: any): void {
    const tasks = this.getTasks(caseId)
    const index = tasks.findIndex((t: any) => t.id === taskId)
    if (index !== -1) {
      tasks[index] = { ...tasks[index], ...updates }
      this.setTasks(caseId, tasks)
    }
  },

  /**
   * 刪除任務
   */
  deleteTask(caseId: string, taskId: string): void {
    const tasks = this.getTasks(caseId)
    const filtered = tasks.filter((t: any) => t.id !== taskId)
    this.setTasks(caseId, filtered)
  },

  /**
   * 重新排序任務
   */
  reorderTasks(caseId: string, taskIds: string[]): void {
    const tasks = this.getTasks(caseId)
    const reordered = taskIds.map(id => tasks.find((t: any) => t.id === id)).filter(Boolean)
    this.setTasks(caseId, reordered)
  }
}

/**
 * 屬性相關的 localStorage 操作
 */
export const fieldsStorage = {
  /**
   * 取得所有屬性
   */
  getFields(): { system_fields: any[]; custom_fields: any[] } {
    if (typeof window === 'undefined') return { system_fields: [], custom_fields: [] }
    try {
      const data = localStorage.getItem(STORAGE_KEYS.CASE_FIELDS)
      return data ? JSON.parse(data) : { system_fields: [], custom_fields: [] }
    } catch {
      return { system_fields: [], custom_fields: [] }
    }
  },

  /**
   * 儲存屬性列表
   */
  setFields(fields: { system_fields: any[]; custom_fields: any[] }): void {
    if (typeof window === 'undefined') return
    try {
      localStorage.setItem(STORAGE_KEYS.CASE_FIELDS, JSON.stringify(fields))
    } catch (error) {
      console.error('Failed to save fields to localStorage:', error)
    }
  },

  /**
   * 新增自定義屬性
   */
  addField(field: any): void {
    const { system_fields, custom_fields } = this.getFields()
    custom_fields.push(field)
    this.setFields({ system_fields, custom_fields })
  },

  /**
   * 更新屬性
   */
  updateField(id: string, updates: any): void {
    const { system_fields, custom_fields } = this.getFields()
    const allFields = [...system_fields, ...custom_fields]
    const index = allFields.findIndex((f: any) => f.id === id)
    if (index !== -1) {
      const field = allFields[index]
      const updated = { ...field, ...updates }
      if (field.is_system) {
        const sysIndex = system_fields.findIndex((f: any) => f.id === id)
        if (sysIndex !== -1) {
          system_fields[sysIndex] = updated
        }
      } else {
        const customIndex = custom_fields.findIndex((f: any) => f.id === id)
        if (customIndex !== -1) {
          custom_fields[customIndex] = updated
        }
      }
      this.setFields({ system_fields, custom_fields })
    }
  },

  /**
   * 刪除屬性（只可能是自定義屬性）
   */
  deleteField(id: string): void {
    const { system_fields, custom_fields } = this.getFields()
    const filtered = custom_fields.filter((f: any) => f.id !== id)
    this.setFields({ system_fields, custom_fields: filtered })
  }
}

/**
 * 合作項目相關的 localStorage 操作
 */
export const collaborationItemsStorage = {
  /**
   * 取得所有合作項目（扁平結構）
   */
  getItems(): any[] {
    if (typeof window === 'undefined') return []
    try {
      const data = localStorage.getItem(STORAGE_KEYS.COLLABORATION_ITEMS)
      return data ? JSON.parse(data) : []
    } catch {
      return []
    }
  },

  /**
   * 儲存合作項目列表（扁平結構）
   */
  setItems(items: any[]): void {
    if (typeof window === 'undefined') return
    try {
      localStorage.setItem(STORAGE_KEYS.COLLABORATION_ITEMS, JSON.stringify(items))
    } catch (error) {
      console.error('Failed to save collaboration items to localStorage:', error)
    }
  },

  /**
   * 新增合作項目
   */
  addItem(item: any): void {
    const items = this.getItems()
    items.push(item)
    this.setItems(items)
  },

  /**
   * 更新合作項目
   */
  updateItem(id: string, updates: any): void {
    const items = this.getItems()
    const index = items.findIndex((i: any) => i.id === id)
    if (index !== -1) {
      items[index] = { ...items[index], ...updates }
      this.setItems(items)
    }
  },

  /**
   * 刪除合作項目（包括所有子項目）
   */
  deleteItem(id: string): void {
    const items = this.getItems()
    // 找出所有子項目
    const getAllChildren = (parentId: string): string[] => {
      const children = items.filter((i: any) => i.parent_id === parentId)
      const childIds = children.map((c: any) => c.id)
      children.forEach((child: any) => {
        childIds.push(...getAllChildren(child.id))
      })
      return childIds
    }
    const allIdsToDelete = [id, ...getAllChildren(id)]
    const filtered = items.filter((i: any) => !allIdsToDelete.includes(i.id))
    this.setItems(filtered)
  }
}

