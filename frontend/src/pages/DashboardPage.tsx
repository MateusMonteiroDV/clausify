import { useQuery } from '@tanstack/react-query';
import { FileText, CheckCircle, XCircle, Clock, TrendingUp, AlertTriangle } from 'lucide-react';
import { documentService, obligationService } from '../services';
import { useAuthStore } from '../store/authStore';
import { RiskScoreBar, StatusBadge, ObligationBadge, formatDate, Spinner } from '../components/ui';
import { useNavigate } from 'react-router-dom';

export default function DashboardPage() {
  const user = useAuthStore((s) => s.user);
  const navigate = useNavigate();

  const { data: stats, isLoading: statsLoading } = useQuery({
    queryKey: ['stats'],
    queryFn: documentService.stats,
  });

  const { data: docsData, isLoading: docsLoading } = useQuery({
    queryKey: ['documents', 1, 5],
    queryFn: () => documentService.list(1, 5),
  });

  const { data: obligData, isLoading: obligLoading } = useQuery({
    queryKey: ['obligations', 1, 5],
    queryFn: () => obligationService.list(1, 5),
  });

  const statCards = [
    { label: 'Total de contratos', value: stats?.total ?? 0, icon: FileText, color: 'var(--color-primary)', bg: 'var(--color-primary-dim)' },
    { label: 'Analisados', value: stats?.analyzed ?? 0, icon: CheckCircle, color: 'var(--color-success)', bg: 'rgba(34,197,94,0.12)' },
    { label: 'Falharam', value: stats?.failed ?? 0, icon: XCircle, color: 'var(--color-danger)', bg: 'rgba(239,68,68,0.12)' },
    { label: 'Aguardando', value: stats?.pending ?? 0, icon: Clock, color: 'var(--color-warning)', bg: 'rgba(245,158,11,0.12)' },
    { label: 'Score médio de risco', value: stats?.avg_risk_score?.toFixed(1) ?? '—', icon: TrendingUp, color: 'var(--color-info)', bg: 'rgba(59,130,246,0.12)' },
  ];

  return (
    <div className="animate-fade-in">
      <div className="page-header">
        <h1 className="page-title">Dashboard</h1>
        <p className="page-subtitle">Bem-vindo, {user?.full_name || user?.email} · {user?.org_name}</p>
      </div>

      {/* Stats */}
      <div className="stats-grid">
        {statCards.map(({ label, value, icon: Icon, color, bg }) => (
          <div key={label} className="stat-card" style={{ '--stat-color': color, '--stat-bg': bg } as React.CSSProperties}>
            <div className="stat-icon">
              <Icon size={20} />
            </div>
            <div className="stat-info">
              {statsLoading ? (
                <div style={{ display: 'flex', alignItems: 'center', height: 28 }}><Spinner size={18} /></div>
              ) : (
                <div className="stat-value">{value}</div>
              )}
              <div className="stat-label">{label}</div>
            </div>
          </div>
        ))}
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24 }}>
        {/* Recent Documents */}
        <div className="table-container">
          <div className="table-header">
            <span className="table-title">Documentos Recentes</span>
            <button className="btn btn-secondary btn-sm" onClick={() => navigate('/documents')}>
              Ver todos
            </button>
          </div>
          {docsLoading ? (
            <div style={{ padding: 40, display: 'flex', justifyContent: 'center' }}><Spinner /></div>
          ) : !docsData?.data?.length ? (
            <div className="empty-state">
              <div className="empty-state-icon"><FileText size={28} /></div>
              <div className="empty-state-title">Nenhum documento ainda</div>
              <div className="empty-state-text">Faça upload do seu primeiro contrato</div>
            </div>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>Arquivo</th>
                  <th>Status</th>
                  <th>Score</th>
                </tr>
              </thead>
              <tbody>
                {docsData.data.map((doc) => (
                  <tr key={doc.id} style={{ cursor: 'pointer' }} onClick={() => navigate(`/documents/${doc.id}`)}>
                    <td>
                      <div style={{ fontWeight: 500, fontSize: '0.875rem', maxWidth: 180 }} className="truncate">
                        {doc.file_name}
                      </div>
                      <div className="text-xs text-muted">{formatDate(doc.created_at)}</div>
                    </td>
                    <td><StatusBadge status={doc.status} /></td>
                    <td><RiskScoreBar score={doc.risk_score} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        {/* Upcoming Obligations */}
        <div className="table-container">
          <div className="table-header">
            <span className="table-title">Obrigações Próximas</span>
            <button className="btn btn-secondary btn-sm" onClick={() => navigate('/obligations')}>
              Ver todas
            </button>
          </div>
          {obligLoading ? (
            <div style={{ padding: 40, display: 'flex', justifyContent: 'center' }}><Spinner /></div>
          ) : !obligData?.data?.length ? (
            <div className="empty-state">
              <div className="empty-state-icon"><AlertTriangle size={28} /></div>
              <div className="empty-state-title">Sem obrigações pendentes</div>
              <div className="empty-state-text">Nenhum deadline próximo</div>
            </div>
          ) : (
            <div style={{ padding: '8px 0' }}>
              {obligData.data.slice(0, 5).map((o) => {
                const isDue = new Date(o.due_date) < new Date();
                const dotColor = o.status === 'COMPLETED' ? 'var(--color-success)'
                               : isDue ? 'var(--color-danger)'
                               : 'var(--color-warning)';
                return (
                  <div key={o.id} className="obligation-card" style={{ margin: '0 16px 8px', borderRadius: 'var(--radius-md)' }}>
                    <div className="obligation-due-dot" style={{ background: dotColor }} />
                    <div className="obligation-info">
                      <div className="obligation-title">{o.title}</div>
                      <div className="obligation-meta">
                        <span>{formatDate(o.due_date)}</span>
                        {o.is_recurring && <span>↻ Recorrente</span>}
                      </div>
                    </div>
                    <ObligationBadge status={o.status} />
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
