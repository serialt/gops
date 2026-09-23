import type { User } from '@/types'
import request from '@/utils/request'

export interface LoginRequest { username: string; password: string; captcha_id?: string; captcha_code?: string; email_code?: string }
export interface LoginResponse { token: string }
export interface CurrentUserResponse { user: User; permissions: string[] }
export interface ProfileUpdateRequest {
  username: string
  nickname: string
  email: string
  current_password: string
  new_password: string
}
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
export interface PasswordResetCodeRequest { username: string; captcha_id?: string; captcha_code?: string }
export interface PasswordResetVerifyRequest { username: string; code: string }
export interface PasswordResetVerifyResponse { reset_token: string }
export interface PasswordResetRequest { reset_token: string; new_password: string }

export const login = (data: LoginRequest) => request.post<LoginResponse>('/auth/login', data)
export const getMe = () => request.post<CurrentUserResponse>('/auth/me', {})
export const updateProfile = (data: ProfileUpdateRequest) => request.post<User>('/auth/profile/update', data)
export const getLoginSecurityConfig = () => request.post<LoginSecurityConfig>('/auth/security-config', {})
export const getCaptcha = () => request.post<CaptchaResponse>('/auth/captcha', {})
export const sendLoginEmailCode = (data: Omit<LoginRequest, 'email_code'>) => request.post<EmailCodeResponse>('/auth/email-code/send', data)
export const sendPasswordResetCode = (data: PasswordResetCodeRequest) => request.post<EmailCodeResponse>('/auth/password-reset/code/send', data)
export const verifyPasswordResetCode = (data: PasswordResetVerifyRequest) => request.post<PasswordResetVerifyResponse>('/auth/password-reset/code/verify', data)
export const resetPassword = (data: PasswordResetRequest) => request.post<null>('/auth/password-reset/confirm', data)
