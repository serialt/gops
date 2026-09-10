import type { ApiResponse, Organization, Permission, Role, User } from '@/types';
import { http } from './http';

export const login = (data: { username: string; password: string }) =>
  http.post<ApiResponse<{ token: string }>>('/auth/login', data)

export const getMe = () =>
  http.post<ApiResponse<{ user: User; permissions: string[] }>>('/auth/me', {})

export const getDashboardStats = () =>
  http.post<ApiResponse<{ users: number; enabled_users: number; roles: number; policies: number }>>('/dashboard/stats', {})

export const getUsers = (keyword = '', organizationId: number | null = null) =>
  http.post<ApiResponse<User[]>>('/users/list', { keyword, organization_id: organizationId })

export const createUser = (data: unknown) => http.post<ApiResponse<User>>('/users/create', data)
export const updateUser = (id: number, data: object) => http.post<ApiResponse<User>>('/users/update', { ...data, id })
export const deleteUser = (id: number) => http.post<ApiResponse<null>>('/users/delete', { id })

export const getRoles = () => http.post<ApiResponse<Role[]>>('/roles/list', {})
export const createRole = (data: unknown) => http.post<ApiResponse<Role>>('/roles/create', data)
export const updateRole = (id: number, data: object) => http.post<ApiResponse<Role>>('/roles/update', { ...data, id })
export const deleteRole = (id: number) => http.post<ApiResponse<null>>('/roles/delete', { id })

export const getPermissionCatalog = () => http.post<ApiResponse<Permission[]>>('/permissions/catalog', {})
export const getRolePermissions = (id: number) => http.post<ApiResponse<Permission[]>>('/roles/permissions/get', { id })
export const setRolePermissions = (id: number, permissions: Permission[]) =>
  http.post<ApiResponse<Permission[]>>('/roles/permissions/set', { id, permissions })

export const getOrganizations = (keyword = '') => http.post<ApiResponse<Organization[]>>('/organizations/list', { keyword })
export const createOrganization = (data: unknown) => http.post<ApiResponse<Organization>>('/organizations/create', data)
export const updateOrganization = (id: number, data: object) => http.post<ApiResponse<Organization>>('/organizations/update', { ...data, id })
export const deleteOrganization = (id: number) => http.post<ApiResponse<null>>('/organizations/delete', { id })
export const moveOrganization = (id: number, data: { parent_id: number | null; sort_order: number }) => http.post<ApiResponse<Organization>>('/organizations/move', { ...data, id })
