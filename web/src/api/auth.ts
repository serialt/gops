import type { User } from '@/types'
import request from '@/utils/request'

export interface LoginRequest { username: string; password: string }
export interface LoginResponse { token: string }
export interface CurrentUserResponse { user: User; permissions: string[] }

export const login = (data: LoginRequest) => request.post<LoginResponse>('/auth/login', data)
export const getMe = () => request.post<CurrentUserResponse>('/auth/me', {})
