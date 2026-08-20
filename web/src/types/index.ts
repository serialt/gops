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

export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}
