export interface Organization {
  id: string;
  name: string;
  slug: string;
  plan_tier: string;
  max_monthly_documents: number;
  created_at: string;
}

export interface User {
  id: string;
  org_id: string;
  email: string;
  full_name: string;
  role: 'admin' | 'auditor' | 'member';
  is_active: boolean;
  created_at: string;
  organization?: Organization;
}

export type DocumentStatus = 'QUEUED' | 'PROCESSING' | 'ANALYZED' | 'FAILED';

export interface Document {
  id: string;
  org_id: string;
  uploaded_by?: string;
  file_name: string;
  storage_path: string;
  file_hash: string;
  file_size_bytes: number;
  mime_type: string;
  status: DocumentStatus;
  risk_score?: number;
  page_count: number;
  analyzed_at?: string;
  created_at: string;
  updated_at: string;
  clauses?: ExtractedClause[];
  obligations?: ContractObligation[];
}

export type RiskLevel = 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';

export interface ExtractedClause {
  id: string;
  org_id: string;
  document_id: string;
  clause_type: string;
  extracted_text: string;
  risk_level: RiskLevel;
  confidence: number;
  summary: string;
  page_number?: number;
  metadata: Record<string, unknown>;
  created_at: string;
}

export type ObligationStatus = 'PENDING' | 'NOTIFIED' | 'COMPLETED' | 'OVERDUE';

export interface ContractObligation {
  id: string;
  org_id: string;
  document_id: string;
  assigned_to?: string;
  title: string;
  description: string;
  due_date: string;
  is_recurring: boolean;
  recurrence_interval?: 'ANNUAL' | 'MONTHLY' | 'QUARTERLY';
  status: ObligationStatus;
  notified_at?: string;
  completed_at?: string;
  notes: string;
  created_at: string;
  updated_at: string;
  document?: Document;
  assignee?: User;
}

export interface AuthResponse {
  token: string;
  user: UserProfile;
}

export interface UserProfile {
  id: string;
  email: string;
  full_name: string;
  role: string;
  org_id: string;
  org_name: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface DashboardStats {
  total: number;
  analyzed: number;
  failed: number;
  pending: number;
  avg_risk_score: number;
}
