import api from '../lib/api';
import type { AuthResponse, PaginatedResponse, Document, ExtractedClause, ContractObligation, DashboardStats } from '../types';

// ─── Auth ────────────────────────────────────────────────────────────────────
export const authService = {
  register: (data: { org_name: string; full_name: string; email: string; password: string }) =>
    api.post<AuthResponse>('/auth/register', data).then((r) => r.data),

  login: (data: { email: string; password: string }) =>
    api.post<AuthResponse>('/auth/login', data).then((r) => r.data),

  me: () => api.get('/auth/me').then((r) => r.data),
};

// ─── Documents ───────────────────────────────────────────────────────────────
export const documentService = {
  list: (page = 1, pageSize = 20) =>
    api.get<PaginatedResponse<Document>>(`/documents?page=${page}&page_size=${pageSize}`).then((r) => r.data),

  getById: (id: string) =>
    api.get<Document>(`/documents/${id}`).then((r) => r.data),

  upload: (file: File, onProgress?: (pct: number) => void) => {
    const form = new FormData();
    form.append('file', file);
    return api.post<Document>('/documents/upload', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => {
        if (onProgress && e.total) onProgress(Math.round((e.loaded * 100) / e.total));
      },
    }).then((r) => r.data);
  },

  downloadUrl: (id: string) => `/api/v1/documents/${id}/file`,

  analyze: (id: string) =>
    api.post(`/documents/${id}/analyze`).then((r) => r.data),

  updateStatus: (id: string, data: { status: string; risk_score?: number; page_count?: number }) =>
    api.patch(`/documents/${id}/status`, data).then((r) => r.data),

  delete: (id: string) => api.delete(`/documents/${id}`),

  stats: () => api.get<DashboardStats>('/documents/stats').then((r) => r.data),
};

// ─── Clauses ─────────────────────────────────────────────────────────────────
export const clauseService = {
  listByDocument: (documentId: string) =>
    api.get<{ data: ExtractedClause[]; total: number }>(`/documents/${documentId}/clauses`).then((r) => r.data),

  create: (documentId: string, data: object) =>
    api.post<ExtractedClause>(`/documents/${documentId}/clauses`, data).then((r) => r.data),
};

// ─── Obligations ─────────────────────────────────────────────────────────────
export const obligationService = {
  list: (page = 1, pageSize = 20) =>
    api.get<PaginatedResponse<ContractObligation>>(`/obligations?page=${page}&page_size=${pageSize}`).then((r) => r.data),

  create: (data: object) =>
    api.post<ContractObligation>('/obligations', data).then((r) => r.data),

  update: (id: string, data: { status?: string; notes?: string; assigned_to?: string }) =>
    api.patch<ContractObligation>(`/obligations/${id}`, data).then((r) => r.data),

  delete: (id: string) => api.delete(`/obligations/${id}`),
};
