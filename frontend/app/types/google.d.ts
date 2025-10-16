// Google Identity Services 類型聲明

interface Window {
  google?: {
    accounts: {
      id: {
        initialize: (config: GoogleInitConfig) => void
        renderButton: (parent: HTMLElement, options: GoogleButtonConfig) => void
        prompt: () => void
      }
    }
  }
}

interface GoogleInitConfig {
  client_id: string
  callback: (response: GoogleCallbackResponse) => void
  auto_select?: boolean
  cancel_on_tap_outside?: boolean
}

interface GoogleButtonConfig {
  theme?: 'outline' | 'filled_blue' | 'filled_black'
  size?: 'large' | 'medium' | 'small'
  type?: 'standard' | 'icon'
  shape?: 'rectangular' | 'pill' | 'circle' | 'square'
  text?: 'signin_with' | 'signup_with' | 'continue_with' | 'signin'
  width?: number
  logo_alignment?: 'left' | 'center'
}

interface GoogleCallbackResponse {
  credential: string
  select_by: string
  clientId?: string
}

export {}

