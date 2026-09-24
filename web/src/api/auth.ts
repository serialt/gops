import type { User } from '@/types'
import request from '@/utils/request'

export interface LoginRequest { username: string; password: string; captcha_id?: string; captcha_code?: string; email_code?: string }
export type MFAMethod = 'code' | 'totp' | 'passkey'
export interface LoginResponse { token:string;access_token:string;refresh_token:string;expires_in:number;refresh_expires_in:number;mfa_required:boolean;mfa_token:string;mfa_methods:MFAMethod[] }
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
export interface MFAPasskey { id:number;name:string;mfa_enabled:boolean;direct_login_enabled:boolean }
export interface MFAStatus { code_available:boolean;email_available:boolean;email:string;email_mfa_enabled:boolean;email_direct_login_enabled:boolean;totp_configured:boolean;totp_enabled:boolean;passkeys:MFAPasskey[] }
export interface TOTPSetup { secret:string;uri:string;qr_code:string }

export const login = (data: LoginRequest) => request.post<LoginResponse>('/auth/login', data)
export const refreshToken = (refresh_token:string) => request.post<LoginResponse>('/auth/refresh', { refresh_token })
export const getMe = () => request.post<CurrentUserResponse>('/auth/me', {})
export const updateProfile = (data: ProfileUpdateRequest) => request.post<User>('/auth/profile/update', data)
export const getLoginSecurityConfig = () => request.post<LoginSecurityConfig>('/auth/security-config', {})
export const getCaptcha = () => request.post<CaptchaResponse>('/auth/captcha', {})
export const sendLoginEmailCode = (data: Omit<LoginRequest, 'email_code'>) => request.post<EmailCodeResponse>('/auth/email-code/send', data)
export const sendPasswordResetCode = (data: PasswordResetCodeRequest) => request.post<EmailCodeResponse>('/auth/password-reset/code/send', data)
export const verifyPasswordResetCode = (data: PasswordResetVerifyRequest) => request.post<PasswordResetVerifyResponse>('/auth/password-reset/code/verify', data)
export const resetPassword = (data: PasswordResetRequest) => request.post<null>('/auth/password-reset/confirm', data)
export const sendMFACode = (mfa_token:string) => request.post<EmailCodeResponse>('/auth/mfa/code/send', { mfa_token })
export const verifyMFA = (data:{mfa_token:string;method:'code'|'totp';code:string}) => request.post<LoginResponse>('/auth/mfa/verify', data)
export const getMFAStatus = () => request.post<MFAStatus>('/auth/mfa/status', {})
export const setupTOTP = () => request.post<TOTPSetup>('/auth/mfa/totp/setup', {})
export const confirmTOTP = (code:string) => request.post<null>('/auth/mfa/totp/confirm', { code })
export const disableTOTP = (code:string) => request.post<null>('/auth/mfa/totp/disable', { code })
export const enableTOTP = (code:string) => request.post<null>('/auth/mfa/totp/enable', { code })
export const deleteTOTP = (code:string) => request.post<null>('/auth/mfa/totp/delete', { code })
export const beginPasskeyRegistration = (name:string) => request.post<{registration_token:string;options:any}>('/auth/mfa/passkeys/register/begin', { name })
export const finishPasskeyRegistration = (data:{registration_token:string;name:string;credential:unknown}) => request.post<null>('/auth/mfa/passkeys/register/finish', data)
export const deletePasskey = (id:number) => request.post<null>('/auth/mfa/passkeys/delete', { id })
export const updatePasskeyStatus = (data:{id:number;mfa_enabled:boolean;direct_login_enabled:boolean}) => request.post<null>('/auth/mfa/passkeys/status', data)
export const renamePasskey = (id:number, name:string) => request.post<null>('/auth/mfa/passkeys/rename', { id, name })
export const updateEmailMFAStatus = (data:{mfa_enabled:boolean;direct_login_enabled:boolean}) => request.post<null>('/auth/mfa/email/status', data)
export const beginPasskeyLogin = (mfa_token:string) => request.post<{options:any}>('/auth/mfa/passkey/login/begin', { mfa_token })
export const finishPasskeyLogin = (data:{mfa_token:string;credential:unknown}) => request.post<LoginResponse>('/auth/mfa/passkey/login/finish', data)
export const beginDirectPasskeyLogin = (data:{username:string;captcha_id?:string;captcha_code?:string}) => request.post<{login_token:string;options:any}>('/auth/passkey/login/begin', data)
export const finishDirectPasskeyLogin = (data:{login_token:string;credential:unknown}) => request.post<LoginResponse>('/auth/passkey/login/finish', data)
export const beginDirectEmailLogin = (data:{username:string;captcha_id?:string;captcha_code?:string}) => request.post<{login_token:string;sent:boolean;retry_after_seconds:number;expires_in_seconds:number}>('/auth/email/login/begin', data)
export const finishDirectEmailLogin = (data:{login_token:string;code:string}) => request.post<LoginResponse>('/auth/email/login/finish', data)
