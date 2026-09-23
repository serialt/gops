import request from '@/utils/request'

export interface SystemConfigResponse {
  system_name: string
  system_description: string
  password_min_length: number
  password_require_digit: boolean
  password_require_letter: boolean
  password_require_upper: boolean
  password_require_lower: boolean
  password_require_special: boolean
  session_timeout_minutes: number
  captcha_enabled: boolean
  captcha_expire_minutes: number
  login_verification_enabled: boolean
  login_verification_method: '' | 'email' | 'webhook'
  max_login_failures: number
  account_lock_minutes: number
  verification_email_enabled: boolean
  verification_email_sender_name: string
  verification_smtp_host: string
  verification_smtp_port: number
  verification_smtp_username: string
  verification_smtp_password: string
  verification_smtp_tls: boolean
  verification_webhook_enabled: boolean
  verification_webhook_url: string
  verification_webhook_secret: string
}
export interface SystemConfigUpdateRequest extends SystemConfigResponse {}
export interface SystemConfigEmailTestRequest { recipient: string }

export const getSystemConfig = () => request.post<SystemConfigResponse>('/system-config/get', {})
export const updateSystemConfig = (data: SystemConfigUpdateRequest) => request.post<SystemConfigResponse>('/system-config/update', data)
export const testSystemConfigEmail = (data: SystemConfigEmailTestRequest) => request.post<null>('/system-config/email/test', data)
