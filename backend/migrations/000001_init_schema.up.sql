-- Enable required PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 1. Organizations (Tenants)
CREATE TABLE organizations (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(100) UNIQUE NOT NULL,
    plan_tier   VARCHAR(50) NOT NULL DEFAULT 'starter',
    max_monthly_documents INT NOT NULL DEFAULT 100,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Users with RBAC
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id        UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name     VARCHAR(255) NOT NULL DEFAULT '',
    role          VARCHAR(50) NOT NULL DEFAULT 'member', -- 'admin', 'auditor', 'member'
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_org_id ON users(org_id);
CREATE INDEX idx_users_email ON users(email);

-- 3. Documents
CREATE TABLE documents (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id       UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    uploaded_by  UUID REFERENCES users(id) ON DELETE SET NULL,
    file_name    VARCHAR(255) NOT NULL,
    storage_path TEXT NOT NULL,
    file_hash    VARCHAR(64) NOT NULL,
    file_size_bytes BIGINT DEFAULT 0,
    mime_type    VARCHAR(100) DEFAULT 'application/pdf',
    status       VARCHAR(50) NOT NULL DEFAULT 'QUEUED',
    -- QUEUED | PROCESSING | ANALYZED | FAILED
    risk_score   NUMERIC(5,2) DEFAULT NULL,
    risk_score_version VARCHAR(20) DEFAULT NULL,
    page_count   INT DEFAULT 0,
    analyzed_at  TIMESTAMPTZ DEFAULT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_documents_org_id ON documents(org_id);
CREATE INDEX idx_documents_status ON documents(status);
CREATE UNIQUE INDEX idx_documents_hash_org ON documents(file_hash, org_id);

-- 4. Extracted Clauses & Compliance Flags
CREATE TABLE extracted_clauses (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    document_id    UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    clause_type    VARCHAR(100) NOT NULL,
    -- LIMITATION_OF_LIABILITY | INDEMNITY | TERMINATION | RENEWAL | PENALTY | CONFIDENTIALITY
    extracted_text TEXT NOT NULL,
    risk_level     VARCHAR(20) NOT NULL DEFAULT 'LOW',
    -- LOW | MEDIUM | HIGH | CRITICAL
    confidence     NUMERIC(4,3) NOT NULL DEFAULT 0.000,
    summary        TEXT NOT NULL DEFAULT '',
    page_number    INT DEFAULT NULL,
    metadata       JSONB NOT NULL DEFAULT '{}',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_clauses_document_id ON extracted_clauses(document_id);
CREATE INDEX idx_clauses_org_id ON extracted_clauses(org_id);
CREATE INDEX idx_clauses_risk_level ON extracted_clauses(risk_level);

-- 5. Tracked Obligations & Deadlines
CREATE TABLE contract_obligations (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    document_id         UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    assigned_to         UUID REFERENCES users(id) ON DELETE SET NULL,
    title               VARCHAR(255) NOT NULL,
    description         TEXT DEFAULT '',
    due_date            DATE NOT NULL,
    is_recurring        BOOLEAN NOT NULL DEFAULT FALSE,
    recurrence_interval VARCHAR(50) DEFAULT NULL, -- ANNUAL | MONTHLY | QUARTERLY
    status              VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    -- PENDING | NOTIFIED | COMPLETED | OVERDUE
    notified_at         TIMESTAMPTZ DEFAULT NULL,
    completed_at        TIMESTAMPTZ DEFAULT NULL,
    completed_by        UUID REFERENCES users(id) ON DELETE SET NULL,
    notes               TEXT DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_obligations_org_id ON contract_obligations(org_id);
CREATE INDEX idx_obligations_document_id ON contract_obligations(document_id);
CREATE INDEX idx_obligations_due_date ON contract_obligations(due_date);
CREATE INDEX idx_obligations_status ON contract_obligations(status);

-- 6. Audit Log
CREATE TABLE audit_logs (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id        UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    actor_id      UUID REFERENCES users(id) ON DELETE SET NULL,
    action        VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id   UUID NOT NULL,
    old_value     JSONB DEFAULT NULL,
    new_value     JSONB DEFAULT NULL,
    ip_address    INET DEFAULT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_org_id ON audit_logs(org_id);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);

-- 7. RLS policies
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE extracted_clauses ENABLE ROW LEVEL SECURITY;
ALTER TABLE contract_obligations ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON documents
    FOR ALL USING (org_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID);

CREATE POLICY tenant_isolation ON extracted_clauses
    FOR ALL USING (org_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID);

CREATE POLICY tenant_isolation ON contract_obligations
    FOR ALL USING (org_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID);

CREATE POLICY tenant_isolation ON audit_logs
    FOR ALL USING (org_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID);

CREATE POLICY users_own_org ON users
    FOR ALL USING (org_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID);
