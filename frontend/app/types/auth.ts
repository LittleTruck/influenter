// 認證相關的型別定義

export interface User {
  id: string
  email: string
  name: string
  profile_picture_url?: string
  ai_instructions?: string
  googleId?: string
  aiReplyTone?: string
  timezone?: string
  notificationPrefs?: NotificationPrefs
  createdAt: string
  updatedAt: string
  lastLoginAt?: string
}

export interface NotificationPrefs {
  emailOnNewCase?: boolean
  emailOnDeadline?: boolean
  browserNotifications?: boolean
}

export interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  loading: boolean
}

export interface LoginResponse {
  user: User
  token: string
}

export interface GoogleAuthResponse {
  credential: string
  clientId: string
}

export interface ApiError {
  message: string
  code?: string
  details?: any
}

