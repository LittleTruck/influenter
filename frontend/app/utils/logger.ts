/**
 * 統一日誌處理工具
 * 在生產環境中可配置為發送到錯誤追蹤服務
 */

type LogLevel = 'debug' | 'info' | 'warn' | 'error'

interface LogContext {
  component?: string
  action?: string
  [key: string]: unknown
}

class Logger {
  private isDevelopment = process.env.NODE_ENV === 'development'

  private log(level: LogLevel, message: string, context?: LogContext, error?: Error) {
    if (!this.isDevelopment && level === 'debug') {
      return
    }

    const timestamp = new Date().toISOString()
    const contextStr = context ? ` [${JSON.stringify(context)}]` : ''
    const errorStr = error ? `\n${error.stack}` : ''

    const logMessage = `[${timestamp}] [${level.toUpperCase()}] ${message}${contextStr}${errorStr}`

    switch (level) {
      case 'debug':
        console.debug(logMessage)
        break
      case 'info':
        console.info(logMessage)
        break
      case 'warn':
        console.warn(logMessage)
        break
      case 'error':
        console.error(logMessage)
        // 在生產環境中可以發送到錯誤追蹤服務
        // if (!this.isDevelopment) {
        //   this.sendToErrorTracking(level, message, context, error)
        // }
        break
    }
  }

  debug(message: string, context?: LogContext) {
    this.log('debug', message, context)
  }

  info(message: string, context?: LogContext) {
    this.log('info', message, context)
  }

  warn(message: string, context?: LogContext) {
    this.log('warn', message, context)
  }

  error(message: string, error?: Error, context?: LogContext) {
    this.log('error', message, context, error)
  }
}

export const logger = new Logger()



