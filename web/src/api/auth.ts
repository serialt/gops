import type { User } from '@/types'
import request from '@/utils/request'

export interface LoginRequest { username: string; password: string; captcha_id?: string; captcha_code?: string; email_code?: string }
export interface LoginResponse { token: string }
export interface CurrentUserResponse { user: User; permissions: string[] }
export interface LoginSecurityConfig {
  captcha_enabled: boolean
  login_verification_enabled: boolean
  login_verification_method: '' | 'email' | 'webhook'
}
export interface CaptchaResponse { captcha_id: string; image: string }
export interface EmailCodeResponse {
  required: boolean
  sent: boolean
  retry_after_seconds: number
  expires_in_seconds: number
}

export const login = (data: LoginRequest) => request.post<LoginResponse>('/auth/login', data)
export const getMe = () => request.post<CurrentUserResponse>('/auth/me', {})
export const getLoginSecurityConfig = () => request.post<LoginSecurityConfig>('/auth/security-config', {})
export const getCaptcha = () => request.post<CaptchaResponse>('/auth/captcha', {})
export const sendLoginEmailCode = (data: Omit<LoginRequest, 'email_code'>) => request.post<EmailCodeResponse>('/auth/email-code/send', data)
