/**
 * 清理郵件 HTML 內容，移除可能影響頁面的標籤和樣式
 */
export const useEmailSanitizer = () => {
  /**
   * 清理 HTML 內容
   * @param html 原始 HTML 字串
   * @returns 清理後的 HTML 字串
   */
  const sanitizeHtml = (html: string): string => {
    if (!html) return ''
    
    // 移除 <style> 標籤及其內容
    let cleaned = html.replace(/<style[^>]*>[\s\S]*?<\/style>/gi, '')
    
    // 移除 <script> 標籤及其內容
    cleaned = cleaned.replace(/<script[^>]*>[\s\S]*?<\/script>/gi, '')
    
    // 移除 <link> 標籤（外部樣式表）
    cleaned = cleaned.replace(/<link[^>]*>/gi, '')
    
    // 移除可能的 style 屬性中的 position、z-index 等會影響佈局的屬性
    cleaned = cleaned.replace(/style="([^"]*)"/gi, (match, styleContent) => {
      // 保留安全的樣式，移除危險的樣式
      const safeStyles = styleContent
        .split(';')
        .filter((style: string) => {
          const prop = style.trim().toLowerCase()
          return !prop.startsWith('position') &&
                 !prop.startsWith('z-index') &&
                 !prop.startsWith('!important')
        })
        .join(';')
      return safeStyles ? `style="${safeStyles}"` : ''
    })
    
    return cleaned
  }

  return {
    sanitizeHtml
  }
}

