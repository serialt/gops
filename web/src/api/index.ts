import type { ApiResponse, Permission, Role, User } from '@/types';
import { http } from './http';

export const login = (data: { username: string; password: string }) =>
  http.post<ApiResponse<{ token: string }>>('/auth/login', data)

export const getMe = () =>
  http.get<ApiResponse<{ user: User; permissions: string[] }>>('/auth/me')

export const getDashboardStats = () =>
  http.get<ApiResponse<{ users: number; enabled_users: number; roles: number; policies: number }>>('/dashboard/stats')

export const getUsers = (keyword = '') =>
  http.get<ApiResponse<User[]>>('/users', { params: { keyword } })

export const createUser = (data: unknown) => http.post<ApiResponse<User>>('/users', data)
export const updateUser = (id: number, data: unknown) => http.put<ApiResponse<User>>(`/users/${id}`, data)
export const deleteUser = (id: number) => http.delete<ApiResponse<null>>(`/users/${id}`)

export const getRoles = () => http.get<ApiResponse<Role[]>>('/roles')
export const createRole = (data: unknown) => http.post<ApiResponse<Role>>('/roles', data)
export const updateRole = (id: number, data: unknown) => http.put<ApiResponse<Role>>(`/roles/${id}`, data)
export const deleteRole = (id: number) => http.delete<ApiResponse<null>>(`/roles/${id}`)

export const getPermissionCatalog = () => http.get<ApiResponse<Permission[]>>('/permissions/catalog')
export const getRolePermissions = (id: number) => http.get<ApiResponse<Permission[]>>(`/roles/${id}/permissions`)
export const setRolePermissions = (id: number, permissions: Permission[]) =>
  http.put<ApiResponse<Permission[]>>(`/roles/${id}/permissions`, { permissions })
