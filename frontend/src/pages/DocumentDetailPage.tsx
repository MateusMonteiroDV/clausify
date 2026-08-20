import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, FileText, Calendar, Shield, AlertTriangle, Zap, Download } from 'lucide-react';
import { documentService, clauseService } from '../services';
import { RiskBadge, RiskScoreBar, StatusBadge, ObligationBadge, formatDate, formatBytes, Spinner } from '../components/ui';

export default function DocumentDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const qc = useQueryClient();
  const [tab, setTab] = useState<'clauses' | 'obligations'>('clauses');

  const { data: doc, isLoading } = useQuery({
    queryKey: ['document', id],
    queryFn: () => documentService.getById(id!),
    enabled: !!id,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      return status === 'QUEUED' || status === 'PROCESSING' ? 3000 : false;
    },
  });

  const analyzeMutation = useMutation({
    mutationFn: () => documentService.analyze(id!),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['document', id] });
      qc.invalidateQueries({ queryKey: ['clauses', id] });
    },
  });

  const { data: clausesData } = useQuery({
    queryKey: ['clauses', id],
    queryFn: () => clauseService.listByDocument(id!),
    enabled: !!id,
  });

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: 80 }}>
        <Spinner size={32} />
      </div>
    );
  }

  if (!doc) {
    return (
      <div className="empty-state">
        <div className="empty-state-title">Documento não encontrado</div>
        <button className="btn btn-secondary" style={{ marginTop: 16 }} onClick={() => navigate('/documents')}>
          Voltar
        </button>
      </div>
    );
  }

  const clauses = clausesData?.data ?? doc.clauses ?? [];
  const obligations = doc.obligations ?? [];

  const criticalCount = clauses.filter((c) => c.risk_level === 'CRITICAL').length;
  const highCount = clauses.filter((c) => c.risk_level === 'HIGH').length;

  return (
    <div className="animate-fade-in">
      {/* Back */}
      <button className="btn btn-secondary btn-sm" style={{ marginBottom: 20 }} onClick={() => navigate('/documents')}>
        <ArrowLeft size={14} /> Voltar
      </button>

      {/* Header */}
      <div className="card" style={{ marginBottom: 24 }}>
        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 16 }}>
          <div style={{ display: 'flex', gap: 16, alignItems: 'flex-start' }}>
            <div style={{
              width: 48, height: 48, borderRadius: 'var(--radius-md)',
              background: 'var(--color-primary-dim)', display: 'flex',
              alignItems: 'center', justifyContent: 'center', flexShrink: 0
            }}>
              <FileText size={22} color="var(--color-primary)" />
            </div>
            <div>
              <h1 style={{ fontSize: '1.3rem', fontWeight: 800 }}>{doc.file_name}</h1>
              <div className="doc-meta">
                <span className="doc-meta-item"><Calendar size={12} />{formatDate(doc.created_at)}</span>
                <span className="doc-meta-item"><Shield size={12} />{formatBytes(doc.file_size_bytes)}</span>
                {doc.page_count > 0 && <span className="doc-meta-item">{doc.page_count} páginas</span>}
                <StatusBadge status={doc.status} />
              </div>
            </div>
          </div>
          <div style={{ display: 'flex', gap: 8, flexShrink: 0, alignItems: 'flex-start' }}>
            {/* Analyze button */}
            <button
              id="btn-analyze"
              className="btn btn-primary btn-sm"
              onClick={() => analyzeMutation.mutate()}
              disabled={analyzeMutation.isPending || doc.status === 'PROCESSING'}
              title="Analisar com Gemini AI"
            >
              {(analyzeMutation.isPending || doc.status === 'PROCESSING')
                ? <><Spinner size={14} /> Analisando...</>
                : <><Zap size={14} /> Analisar</>}
            </button>
            {/* Download button */}
            <a
              id="btn-download"
              className="btn btn-secondary btn-sm"
              href={documentService.downloadUrl(doc.id)}
              download={doc.file_name}
              title="Baixar arquivo"
            >
              <Download size={14} /> Baixar
            </a>
            <div style={{ textAlign: 'right' }}>
              <div style={{ fontSize: '0.75rem', color: 'var(--color-text-muted)', marginBottom: 6 }}>Score de Risco</div>
              <RiskScoreBar score={doc.risk_score} />
            </div>
          </div>
        </div>

        {/* Risk summary */}
        {(criticalCount > 0 || highCount > 0) && (
          <div style={{
            marginTop: 16, padding: '10px 14px', borderRadius: 'var(--radius-md)',
            background: 'rgba(239,68,68,0.08)', border: '1px solid rgba(239,68,68,0.2)',
            display: 'flex', alignItems: 'center', gap: 10, fontSize: '0.875rem'
          }}>
            <AlertTriangle size={16} color="var(--color-danger)" />
            <span>
              <strong style={{ color: 'var(--color-danger)' }}>{criticalCount} cláusula{criticalCount !== 1 ? 's' : ''} crítica{criticalCount !== 1 ? 's' : ''}</strong>
              {highCount > 0 && <>, {highCount} de alto risco</>}
              {' '}detectadas neste contrato.
            </span>
          </div>
        )}
      </div>

      {/* Tabs */}
      <div className="tabs">
        <button className={`tab${tab === 'clauses' ? ' active' : ''}`} onClick={() => setTab('clauses')}>
          Cláusulas ({clauses.length})
        </button>
        <button className={`tab${tab === 'obligations' ? ' active' : ''}`} onClick={() => setTab('obligations')}>
          Obrigações ({obligations.length})
        </button>
      </div>

      {/* Clauses */}
      {tab === 'clauses' && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {clauses.length === 0 ? (
            <div className="empty-state">
              <div className="empty-state-icon"><Shield size={28} /></div>
              <div className="empty-state-title">Sem cláusulas extraídas</div>
              <div className="empty-state-text">O documento ainda não foi analisado ou não tem cláusulas identificadas</div>
            </div>
          ) : (
            clauses.map((clause) => (
              <div key={clause.id} className="clause-card">
                <div className="clause-header">
                  <span className="clause-type">{clause.clause_type.replace(/_/g, ' ')}</span>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    {clause.page_number && (
                      <span style={{ fontSize: '0.72rem', color: 'var(--color-text-muted)' }}>pág. {clause.page_number}</span>
                    )}
                    <RiskBadge level={clause.risk_level} />
                  </div>
                </div>
                {clause.summary && <p className="clause-summary">{clause.summary}</p>}
                <blockquote className="clause-text">{clause.extracted_text}</blockquote>
                <div className="confidence-bar">
                  <span className="confidence-label">Confiança</span>
                  <div className="confidence-track">
                    <div className="confidence-fill" style={{ width: `${clause.confidence * 100}%` }} />
                  </div>
                  <span className="confidence-value">{(clause.confidence * 100).toFixed(0)}%</span>
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {/* Obligations */}
      {tab === 'obligations' && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {obligations.length === 0 ? (
            <div className="empty-state">
              <div className="empty-state-icon"><AlertTriangle size={28} /></div>
              <div className="empty-state-title">Sem obrigações rastreadas</div>
              <div className="empty-state-text">Nenhuma obrigação foi extraída deste contrato</div>
            </div>
          ) : (
            obligations.map((o) => {
              const isOverdue = new Date(o.due_date) < new Date() && o.status === 'PENDING';
              return (
                <div key={o.id} className="obligation-card">
                  <div className="obligation-due-dot" style={{
                    background: isOverdue ? 'var(--color-danger)' : o.status === 'COMPLETED'
                      ? 'var(--color-success)' : 'var(--color-warning)'
                  }} />
                  <div className="obligation-info">
                    <div className="obligation-title">{o.title}</div>
                    {o.description && <div className="text-sm text-muted">{o.description}</div>}
                    <div className="obligation-meta">
                      <span>Vence em: <strong>{formatDate(o.due_date)}</strong></span>
                      {o.is_recurring && <span>↻ {o.recurrence_interval}</span>}
                    </div>
                  </div>
                  <ObligationBadge status={o.status} />
                </div>
              );
            })
          )}
        </div>
      )}
    </div>
  );
}
