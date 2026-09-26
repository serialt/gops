import type { Permission, Role } from '@/types'
import request, { type PageResponse } from '@/utils/request'

export interface RoleResponse extends Role { }
export interface RoleCreateRequest { code: string; name: string; description: string }
export interface RoleUpdateRequest extends RoleCreateRequest { id: number }
export interface RoleIDRequest { id: number }
export interface PermissionResponse extends Permission { }
export interface RolePermissionsRequest { id: number }
export interface SetRolePermissionsRequest extends RolePermissionsRequest { permissions: PermissionResponse[] }

export const getRoles = (page = 1, limit = 20) => request.post<PageResponse<RoleResponse>>('/roles/list', { page, limit })
export const createRole = (data: RoleCreateRequest) => request.post<RoleResponse>('/roles/create', data)
export const updateRole = (id: number, data: RoleCreateRequest) => request.post<RoleResponse>('/roles/update', { ...data, id } satisfies RoleUpdateRequest)
export const deleteRole = (id: number) => request.post<null>('/roles/delete', { id } satisfies RoleIDRequest)
export const getPermissionCatalog = () => request.post<PermissionResponse[]>('/permissions/catalog', {})
export const getRolePermissions = (id: number) => request.post<PermissionResponse[]>('/roles/permissions/get', { id } satisfies RolePermissionsRequest)
export const setRolePermissions = (id: number, permissions: PermissionResponse[]) => request.post<PermissionResponse[]>('/roles/permissions/set', { id, permissions } satisfies SetRolePermissionsRequest)
