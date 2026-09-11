export interface Role {
  id: number
  code: string
  name: string
  description: string
  created_at: string
  updated_at: string
}

export interface User {
  id: number
  username: string
  nickname: string
  status: boolean
  roles: Role[]
  organizations: Organization[]
  created_at: string
  updated_at: string
}

export interface Permission {
  group: string
  name: string
  path: string
  method: string
  key: string
}

export interface Organization { id: number; parent_id: number | null; sort_order: number; name: string; code: string; manager: string; status: boolean; created_at: string; updated_at: string }
