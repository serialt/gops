import type { User } from '@/types'
import request, { type PageResponse } from '@/utils/request'

export interface UserResponse extends User { }
export interface UserListRequest { keyword: string; organization_id: number | null; page: number; limit: number }
export interface UserCreateRequest { username: string; email: string; password: string; nickname: string; status: boolean; role_ids: number[]; organization_ids: number[] }
export interface UserUpdateRequest extends UserCreateRequest { id: number }
export interface UserIDRequest { id: number }

export const getUsers = (keyword = '', organizationId: number | null = null, page = 1, limit = 20) => request.post<PageResponse<UserResponse>>('/users/list', { keyword, organization_id: organizationId, page, limit } satisfies UserListRequest)
export const createUser = (data: UserCreateRequest) => request.post<UserResponse>('/users/create', data)
export const updateUser = (id: number, data: UserCreateRequest) => request.post<UserResponse>('/users/update', { ...data, id } satisfies UserUpdateRequest)
export const deleteUser = (id: number) => request.post<null>('/users/delete', { id } satisfies UserIDRequest)
