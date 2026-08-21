package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/clausify/backend/internal/models"
	"github.com/clausify/backend/internal/repository"
	"github.com/ledongthuc/pdf"
	"go.uber.org/zap"
	"google.golang.org/genai"
)

// ─── Gemini response structures ──────────────────────────────────────────────

type geminiClause struct {
	ClauseType    string  `json:"clause_type"`
	ExtractedText string  `json:"extracted_text"`
	Summary       string  `json:"summary"`
	RiskLevel     string  `json:"risk_level"` // LOW | MEDIUM | HIGH | CRITICAL
	Confidence    float64 `json:"confidence"`
	PageNumber    *int    `json:"page_number"`
}

type geminiObligation struct {
	Title               string  `json:"title"`
	Description         string  `json:"description"`
	DueDate             string  `json:"due_date"` // YYYY-MM-DD
	IsRecurring         bool    `json:"is_recurring"`
	RecurrenceInterval  *string `json:"recurrence_interval"` // DAILY|WEEKLY|MONTHLY|YEARLY
}

type geminiAnalysisResult struct {
	Clauses     []geminiClause     `json:"clauses"`
	Obligations []geminiObligation `json:"obligations"`
	RiskScore   float64            `json:"risk_score"` // 0-100
	PageCount   int                `json:"page_count"`
}

// ─── AnalysisService ─────────────────────────────────────────────────────────

type AnalysisService struct {
	docRepo    *repository.DocumentRepository
	clauseRepo *repository.ClauseRepository
	obligRepo  *repository.ObligationRepository
	apiKey     string
	logger     *zap.Logger
}

func NewAnalysisService(
	docRepo *repository.DocumentRepository,
	clauseRepo *repository.ClauseRepository,
	obligRepo *repository.ObligationRepository,
	apiKey string,
	logger *zap.Logger,
) *AnalysisService {
	return &AnalysisService{
		docRepo:    docRepo,
		clauseRepo: clauseRepo,
		obligRepo:  obligRepo,
		apiKey:     apiKey,
		logger:     logger,
	}
}

// Analyze runs the full pipeline asynchronously. Call this in a goroutine after upload.
func (s *AnalysisService) Analyze(doc *models.Document) {
	ctx := context.Background()
	log := s.logger.With(zap.String("doc_id", doc.ID.String()))

	// Mark as PROCESSING
	_ = s.docRepo.UpdateStatus(doc.ID, models.StatusProcessing, nil, 0)
	log.Info("analysis started")

	result, err := s.runAnalysis(ctx, doc)
	if err != nil {
		log.Error("analysis failed", zap.Error(err))
		_ = s.docRepo.UpdateStatus(doc.ID, models.StatusFailed, nil, 0)
		return
	}

	// Persist clauses
	for _, gc := range result.Clauses {
		pg := gc.PageNumber
		clause := &models.ExtractedClause{
			OrgID:         doc.OrgID,
			DocumentID:    doc.ID,
			ClauseType:    gc.ClauseType,
			ExtractedText: gc.ExtractedText,
			Summary:       gc.Summary,
			RiskLevel:     models.RiskLevel(gc.RiskLevel),
			Confidence:    gc.Confidence,
			PageNumber:    pg,
		}
		if err := s.clauseRepo.Create(clause); err != nil {
			log.Warn("failed to save clause", zap.Error(err))
		}
	}

	// Persist obligations
	for _, go_ := range result.Obligations {
		dueDate, err := time.Parse("2006-01-02", go_.DueDate)
		if err != nil {
			// Skip obligations with invalid dates
			log.Warn("invalid due_date from gemini", zap.String("date", go_.DueDate))
			continue
		}
		o := &models.ContractObligation{
			OrgID:       doc.OrgID,
			DocumentID:  doc.ID,
			Title:       go_.Title,
			Description: go_.Description,
			DueDate:     dueDate,
			IsRecurring: go_.IsRecurring,
			Status:      models.ObligationPending,
		}
		if go_.RecurrenceInterval != nil {
			ri := models.RecurrenceInterval(*go_.RecurrenceInterval)
			o.RecurrenceInterval = &ri
		}
		if err := s.obligRepo.Create(o); err != nil {
			log.Warn("failed to save obligation", zap.Error(err))
		}
	}

	// Update document status to ANALYZED
	_ = s.docRepo.UpdateStatus(doc.ID, models.StatusAnalyzed, &result.RiskScore, result.PageCount)
	log.Info("analysis complete",
		zap.Int("clauses", len(result.Clauses)),
		zap.Int("obligations", len(result.Obligations)),
		zap.Float64("risk_score", result.RiskScore),
	)
}

func (s *AnalysisService) runAnalysis(ctx context.Context, doc *models.Document) (*geminiAnalysisResult, error) {
	// 1. Extract text from PDF
	text, pageCount, err := extractPDFText(doc.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("pdf extraction: %w", err)
	}
	if len(strings.TrimSpace(text)) == 0 {
		return nil, fmt.Errorf("no text extracted from PDF")
	}

	// Trim to ~30k chars to stay within Gemini token limits
	const maxChars = 30000
	if len(text) > maxChars {
		text = text[:maxChars] + "\n[... documento truncado para análise ...]"
	}

	// 2. Call Gemini
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  s.apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini client: %w", err)
	}

	prompt := buildPrompt(text, pageCount)
	resp, err := client.Models.GenerateContent(
		ctx,
		"gemini-1.5-flash",
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("gemini generate: %w", err)
	}

	// 3. Parse response
	if resp == nil || len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("empty gemini response")
	}

	rawJSON := resp.Text()
	rawJSON = extractJSON(rawJSON)

	var result geminiAnalysisResult
	if err := json.Unmarshal([]byte(rawJSON), &result); err != nil {
		return nil, fmt.Errorf("parse gemini json: %w — raw: %s", err, rawJSON[:min(200, len(rawJSON))])
	}

	result.PageCount = pageCount
	return &result, nil
}

// ─── PDF text extraction ─────────────────────────────────────────────────────

func extractPDFText(filePath string) (string, int, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		// Fallback: try reading as plain text file (useful for test text files or mock contracts)
		data, readErr := os.ReadFile(filePath)
		if readErr == nil && len(data) > 0 && !bytes.Contains(data, []byte{0}) {
			return string(data), 1, nil
		}
		return "", 0, fmt.Errorf("open pdf: %w", err)
	}
	defer f.Close()

	pageCount := r.NumPage()
	var buf bytes.Buffer

	for i := 1; i <= pageCount; i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}

	return buf.String(), pageCount, nil
}

// ─── Prompt builder ──────────────────────────────────────────────────────────

func buildPrompt(contractText string, pageCount int) string {
	today := time.Now().Format("2006-01-02")
	return fmt.Sprintf(`Você é um especialista jurídico em análise de contratos empresariais brasileiros.

Analise o contrato abaixo e retorne um JSON com a seguinte estrutura exata:

{
  "clauses": [
    {
      "clause_type": "string (ex: PENALIDADE, RESCISÃO, PRAZO, PAGAMENTO, CONFIDENCIALIDADE, GARANTIA, RENOVAÇÃO, OUTRO)",
      "extracted_text": "texto exato da cláusula",
      "summary": "resumo em 1-2 frases",
      "risk_level": "LOW | MEDIUM | HIGH | CRITICAL",
      "confidence": 0.95,
      "page_number": 1
    }
  ],
  "obligations": [
    {
      "title": "título curto da obrigação",
      "description": "descrição detalhada",
      "due_date": "YYYY-MM-DD",
      "is_recurring": false,
      "recurrence_interval": null
    }
  ],
  "risk_score": 45.0
}

Regras:
- Extraia TODAS as cláusulas relevantes de risco (penalidades, multas, prazos críticos, rescisão, confidencialidade)
- Para obrigações, inclua apenas as que têm prazo definido ou recorrência clara
- Se uma data não for explícita, estime com base no contexto (hoje é %s)
- risk_score é de 0 a 100 (0 = sem risco, 100 = risco extremo)
- recurrence_interval pode ser: DAILY, WEEKLY, MONTHLY, YEARLY ou null
- Responda APENAS com o JSON, sem markdown, sem texto adicional
- O contrato tem %d páginas

CONTRATO:
%s`, today, pageCount, contractText)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// extractJSON strips markdown code fences from the response if present.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	// Remove ```json ... ``` or ``` ... ```
	if strings.HasPrefix(s, "```") {
		end := strings.LastIndex(s, "```")
		if end > 3 {
			s = s[3:end]
		}
		// Remove optional language tag (e.g. "json\n")
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		}
	}
	return strings.TrimSpace(s)
}

