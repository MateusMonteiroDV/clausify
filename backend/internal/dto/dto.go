package dto

import "github.com/google/uuid"

// ─── Auth ───────────────────────────────────────────────────────────────────

type RegisterRequest struct {
	OrgName  string `json:"org_name"  binding:"required,min=2,max=100"`
	FullName string `json:"full_name"  binding:"required,min=2,max=100"`
	Email    string `json:"email"     binding:"required,email"`
	Password string `json:"password"  binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  UserProfile `json:"user"`
}

type UserProfile struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	FullName string    `json:"full_name"`
	Role     string    `json:"role"`
	OrgID    uuid.UUID `json:"org_id"`
	OrgName  string    `json:"org_name"`
}

// ─── Documents ──────────────────────────────────────────────────────────────

type CreateDocumentRequest struct {
	FileName      string `json:"file_name"       binding:"required"`
	StoragePath   string `json:"storage_path"    binding:"required"`
	FileHash      string `json:"file_hash"       binding:"required"`
	FileSizeBytes int64  `json:"file_size_bytes"`
	MimeType      string `json:"mime_type"`
}

type UpdateDocumentStatusRequest struct {
	Status    string   `json:"status"     binding:"required,oneof=QUEUED PROCESSING ANALYZED FAILED"`
	RiskScore *float64 `json:"risk_score"`
	PageCount int      `json:"page_count"`
}

// ─── Clauses ────────────────────────────────────────────────────────────────

type CreateClauseRequest struct {
	DocumentID    string  `json:"document_id"    binding:"required,uuid"`
	ClauseType    string  `json:"clause_type"    binding:"required"`
	ExtractedText string  `json:"extracted_text" binding:"required"`
	RiskLevel     string  `json:"risk_level"     binding:"required,oneof=LOW MEDIUM HIGH CRITICAL"`
	Confidence    float64 `json:"confidence"     binding:"required,min=0,max=1"`
	Summary       string  `json:"summary"`
	PageNumber    *int    `json:"page_number"`
}

// ─── Obligations ─────────────────────────────────────────────────────────────

type CreateObligationRequest struct {
	DocumentID         string  `json:"document_id"          binding:"required,uuid"`
	Title              string  `json:"title"                binding:"required,min=3"`
	Description        string  `json:"description"`
	DueDate            string  `json:"due_date"             binding:"required"` // YYYY-MM-DD
	IsRecurring        bool    `json:"is_recurring"`
	RecurrenceInterval *string `json:"recurrence_interval"`
	AssignedTo         *string `json:"assigned_to"`
}

type UpdateObligationRequest struct {
	Status      *string `json:"status"       binding:"omitempty,oneof=PENDING NOTIFIED COMPLETED OVERDUE"`
	Notes       *string `json:"notes"`
	AssignedTo  *string `json:"assigned_to"`
}

// ─── Pagination ──────────────────────────────────────────────────────────────

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}
