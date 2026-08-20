# Clausify — System Architecture




## 1. Executive Summary & Problem Statement

**Clausify** is an enterprise-grade, multi-tenant B2B document intelligence and contract compliance platform. Organizations execute and store thousands of legal and operational contracts annually, suffering from:
- Silent SLA and renewal expirations (leading to unwanted auto-renewals).
- Undetected non-standard liability clauses, penalty terms, and uncapped indemnification obligations.
- Massive manual labor spent by legal/procurement teams reviewing recurring standard agreements.

Clausify automates this pipeline with end-to-end document ingestion, layout-aware OCR extraction, semantic vector chunking, structured LLM extraction (JSON Schema enforcement), automated compliance scoring, and continuous obligation tracking.

---

## 2. High-Level System Architecture

```
                                +-----------------------------------+
                                |       Next.js 14 Web Client       |
                                | (Dashboard, Annotation, Realtime) |
                                +-----------------+-----------------+
                                                  | (HTTPS / WSS)
                                                  v
+------------------------+              +-------------------+              +-------------------------+
| S3-Compatible Storage  |<-------------|  FastAPI Gateway  |------------->|  PostgreSQL 16 + RLS    |
| (AWS S3 / Cloudflare)  | (Presigned)  |  (Core API & Auth)|              |  + pgvector Extension   |
+------------------------+              +---------+---------+              +-------------------------+
                                                  |
                                                  | Enqueues Job
                                                  v
                                        +-------------------+
                                        |   Redis Broker    |
                                        | (Celery / BullMQ) |
                                        +---------+---------+
                                                  |
                                                  v
                                +-----------------------------------+
                                |      Async Ingestion Engine       |
                                | - PyMuPDF / Textract OCR Layout   |
                                | - Semantic Chunker + Embedder     |
                                | - Structured LLM Extractor        |
                                | - Risk & Obligation Rule Evaluator|
                                +-----------------+-----------------+
                                                  |
                                                  v
                                +-----------------------------------+
                                |    Notifications & Webhooks       |
                                |    (Slack, Email, Zapier, Webhook)|
                                +-----------------------------------+
```

---

## 3. Technology Stack & Component Breakdown

| Layer | Technology | Key Responsibility & Rationale |
| :--- | :--- | :--- |
| **Frontend UI** | **Next.js 14 (App Router, TS, Tailwind)** | Server-side rendering (SSR), fast client dashboard, PDF.js side-by-side view with visual bounding box highlights. |
| **API Gateway** | **FastAPI (Python 3.12)** | High-throughput asynchronous routing, strict Pydantic v2 data contracts, JWT/OAuth2 RBAC authentication. |
| **Task Queue** | **Celery / ARQ + Redis 7** | Decoupled background processing for long-running I/O and compute-heavy AI tasks. |
| **Database & Vector** | **PostgreSQL 16 + `pgvector`** | Multi-tenant transactional relational storage with Native Row-Level Security (RLS) and hybrid cosine vector indexing. |
| **Blob Storage** | **AWS S3 / Cloudflare R2** | Encrypted document storage with time-limited presigned URLs (no document passes through API server memory). |
| **OCR & Parsing** | **PyMuPDF (fitz) + Tesseract/Textract** | Native PDF text extraction and fallback OCR preserving page coordinates, font weights, and table structures. |
| **LLM & Inference** | **OpenAI API / Anthropic Claude 3.5** | Strict JSON Schema validation, structured extraction of contractual obligations and risk clauses with zero data retention. |
| **Observability** | **OpenTelemetry + Prometheus + Sentry** | Distributed tracing across API routes, queue wait latencies, LLM token usages, and error tracking. |

---

## 4. End-to-End Ingestion & Processing Pipeline

```
[1. Upload Request]  --> API issues short-lived Presigned S3 PUT URL
[2. S3 Direct Upload]--> Client pushes binary directly to Object Storage
[3. Ingestion Event] --> Client triggers `POST /api/v1/documents/process`
[4. Dispatch]        --> API computes SHA-256 hash, creates record (status=QUEUED), enqueues task
[5. Layout Parsing]  --> Worker pulls PDF, parses text blocks, extracts visual bounding boxes
[6. Chunking & Embed]--> Hierarchical semantic chunking (clauses/sections) -> Generates 1536-d embeddings
[7. Risk Evaluation] --> Hybrid RAG retrieval + LLM structured extraction (JSON Schema enforcement)
[8. Audit & Commit]  --> Worker persists extracted clauses, updates document status to `ANALYZED`
[9. Notification]    --> Event broadcast via WebSockets to client; triggers webhooks / Slack alerts
```

---

## 5. Multi-Tenancy & Security Architecture

### 5.1 Row-Level Security (RLS)
Data isolation is strictly enforced at the database engine level. The API gateway sets a session variable on every database connection:
```sql
SET LOCAL app.current_tenant_id = 'tenant_uuid_here';
```

### 5.2 Zero Data Retention (ZDR) & Compliance
- Enterprise LLM endpoints are configured with **Zero Data Retention** agreements to guarantee uploaded contracts are never used for model retraining.
- PII Redaction module masks financial account numbers, Brazilian CPF/CNPJ, and sensitive personal identifiers prior to external model transit where required.
- All documents receive a cryptographic `SHA-256` checksum at ingestion to establish non-repudiation audit logs.

---

## 6. Database Schema Definition

```sql
-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";

-- 1. Organizations (Tenants)
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    plan_tier VARCHAR(50) NOT NULL DEFAULT 'starter',
    max_monthly_documents INT NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Users with RBAC
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member', -- 'admin', 'auditor', 'member'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Documents
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    storage_path TEXT NOT NULL,
    file_hash VARCHAR(64) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'QUEUED', -- 'QUEUED', 'PROCESSING', 'ANALYZED', 'FAILED'
    risk_score INT DEFAULT NULL,
    page_count INT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. Semantic Chunks & Embeddings
CREATE TABLE document_chunks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL,
    content TEXT NOT NULL,
    bounding_box JSONB, -- {"page": 1, "x": 50, "y": 120, "w": 400, "h": 60}
    embedding vector(1536),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 5. Extracted Clauses & Compliance Flags
CREATE TABLE extracted_clauses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    clause_type VARCHAR(100) NOT NULL, -- e.g., 'LIMITATION_OF_LIABILITY', 'INDEMNITY', 'TERMINATION'
    extracted_text TEXT NOT NULL,
    risk_level VARCHAR(20) NOT NULL, -- 'LOW', 'MEDIUM', 'HIGH', 'CRITICAL'
    confidence NUMERIC(4, 3) NOT NULL,
    summary TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 6. Tracked Obligations & Deadlines
CREATE TABLE contract_obligations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    due_date DATE NOT NULL,
    is_recurring BOOLEAN NOT NULL DEFAULT FALSE,
    recurrence_interval VARCHAR(50), -- 'ANNUAL', 'MONTHLY', 'QUARTERLY'
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING', -- 'PENDING', 'NOTIFIED', 'COMPLETED'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Enable Row-Level Security
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE document_chunks ENABLE ROW LEVEL SECURITY;
ALTER TABLE extracted_clauses ENABLE ROW LEVEL SECURITY;
ALTER TABLE contract_obligations ENABLE ROW LEVEL SECURITY;

-- Sample RLS Policy
CREATE POLICY tenant_isolation_policy ON documents
    FOR ALL
    USING (org_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID);
```

---

## 7. Resilience, Fault Tolerance & Observability

- **Idempotency:** Every ingestion job is keyed by `SHA-256(file_binary) + org_id`. Repetitive uploads prevent duplicated background jobs and duplicate billing charges.
- **Dead Letter Queue (DLQ):** Tasks that fail OCR parsing or exceed LLM rate-limit retries are automatically moved to a Redis DLQ with full stack traces for forensic debugging.
- **Circuit Breakers & Exponential Backoff:** External AI API requests utilize the `tenacity` library with jittered exponential backoff to handle transient provider downtime without losing document state.
- **Metrics & Tracing:** OpenTelemetry traces document processing end-to-end, measuring:
  - Presigned S3 URL generation latency.
  - OCR extraction throughput (pages/sec).
  - LLM token consumption & inference latency.
  - Database vector search lookup speeds.