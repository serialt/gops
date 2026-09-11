import type { Organization } from '@/types'
import request from '@/utils/request'

export interface OrganizationResponse extends Organization {}
export interface OrganizationListRequest { keyword: string }
export interface OrganizationCreateRequest { parent_id: number | null; name: string; code: string; manager: string; status: boolean }
export interface OrganizationUpdateRequest extends OrganizationCreateRequest { id: number }
export interface OrganizationMoveRequest { parent_id: number | null; sort_order: number }
export interface OrganizationIDRequest { id: number }

export const getOrganizations = (keyword = '') => request.post<OrganizationResponse[]>('/organizations/list', { keyword } satisfies OrganizationListRequest)
export const createOrganization = (data: OrganizationCreateRequest) => request.post<OrganizationResponse>('/organizations/create', data)
export const updateOrganization = (id: number, data: OrganizationCreateRequest) => request.post<OrganizationResponse>('/organizations/update', { ...data, id } satisfies OrganizationUpdateRequest)
export const deleteOrganization = (id: number) => request.post<null>('/organizations/delete', { id } satisfies OrganizationIDRequest)
export const moveOrganization = (id: number, data: OrganizationMoveRequest) => request.post<OrganizationResponse>('/organizations/move', { ...data, id })
