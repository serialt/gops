import request from '@/utils/request'
import type { PageResponse } from '@/utils/request'

export type InspectionStatus = 'running' | 'completed' | 'failed'
export type CheckStatus = 'success' | 'warning' | 'error'
export interface InspectionCheck { category: string; name: string; status: CheckStatus; value: string; expected: string; detail: string; suggestion: string }
export interface InspectionMetric { label: string; value: string; tone?: string }
export interface InspectionSection { key: string; title: string; metrics?: InspectionMetric[]; details?: Record<string, unknown>[]; checks: InspectionCheck[] }
export interface InspectionSummary { total: number; passed: number; warning: number; failed: number; duration: number }
export interface InspectionReport { cluster_id: number; cluster_name: string; score: number; summary: InspectionSummary; sections: InspectionSection[] }
export interface InspectionHistory { id: number; cluster_id: number; cluster_name: string; status: InspectionStatus; score: number; check_count: number; pass_count: number; warning_count: number; fail_count: number; duration: number; started_at: string; ended_at: string }
export interface InspectionProgress { id: number; status: InspectionStatus; progress: number; current_step: string }

export const startInspection = (cluster_id: number) => request.post<{ id: number }>('/kubernetes/inspections/start', { cluster_id })
export const getInspectionProgress = (id: number) => request.post<InspectionProgress>('/kubernetes/inspections/progress', { id })
export const getInspectionResult = (id: number) => request.post<{ inspection: InspectionHistory; report: InspectionReport }>('/kubernetes/inspections/result', { id })
export const listInspectionHistory = (cluster_id = 0, page = 1, limit = 20) => request.post<PageResponse<InspectionHistory>>('/kubernetes/inspections/history', { cluster_id, page, limit })
export const deleteInspection = (id: number) => request.post<null>('/kubernetes/inspections/delete', { id })
