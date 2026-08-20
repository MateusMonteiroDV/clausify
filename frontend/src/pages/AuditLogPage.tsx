import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { RefreshCw, FileText, Activity } from 'lucide-react';
import { auditService } from '../services';
import { formatDate, Spinner } from '../components/ui';

export default function AuditLogPage() {
  const [page, setPage] = useState(1);
  const [filterAction, setFilterAction] = useState('');

  const { data, isLoading, refetch, isRefetching } = useQuery({
    queryKey: ['audit-logs', page],
    queryFn: () => auditService.list(page, 20),
  });

  const logs = data?.data ?? [];

  const filteredLogs = logs.filter((log) =>
    !filterAction ||
    log.action.toLowerCase().includes(filterAction.toLowerCase()) ||
    log.resource_type.toLowerCase().includes(filterAction.toLowerCase())
  );

  return (
    <div className="animate-fade-in">
      <div className="page-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <h1 className="page-title">Logs de Auditoria & Atividade</h1>
          <p className="page-subtitle">Rastreabilidade completa de ações, uploads e execuções de IA via PostgreSQL</p>
        </div>
        <button
          className="btn btn-secondary btn-sm"
          onClick={() => refetch()}
          disabled={isLoading || isRefetching}
        >
          {isRefetching ? <Spinner size={14} /> : <><RefreshCw size={14} /> Atualizar</>}
        </button>
      </div>

      {/* Filter Bar */}
      <div style={{ display: 'flex', gap: 12, marginBottom: 20 }}>
        <div style={{ flex: 1, position: 'relative' }}>
          <input
            className="form-input"
            placeholder="Filtrar por ação ou recurso (ex: DOCUMENT_UPLOADED)..."
            value={filterAction}
            onChange={(e) => setFilterAction(e.target.value)}
          />
        </div>
      </div>

      {/* Logs Table */}
      <div className="table-container">
        <div className="table-header">
          <span className="table-title">
            Histórico de Eventos do Banco de Dados
            {data?.total != null && <span style={{ color: 'var(--color-text-muted)', fontWeight: 400 }}> ({data.total} total)</span>}
          </span>
        </div>

        {isLoading ? (
          <div style={{ padding: 60, display: 'flex', justifyContent: 'center' }}><Spinner /></div>
        ) : !filteredLogs.length ? (
          <div className="empty-state">
            <div className="empty-state-icon"><Activity size={30} /></div>
            <div className="empty-state-title">Nenhum evento registrado</div>
            <div className="empty-state-text">
              {filterAction ? 'Nenhum log encontrado para esta busca.' : 'Ações no sistema (como uploads e análises) serão registradas aqui em tempo real.'}
            </div>
          </div>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Ação / Evento</th>
                <th>Tipo de Recurso</th>
                <th>ID do Recurso</th>
                <th>Ator / Usuário</th>
                <th>IP</th>
                <th>Data & Hora</th>
              </tr>
            </thead>
            <tbody>
              {filteredLogs.map((log) => (
                <tr key={log.id}>
                  <td>
                    <span className="badge badge-analyzed">
                      {log.action}
                    </span>
                  </td>
                  <td style={{ fontWeight: 600 }}>{log.resource_type}</td>
                  <td className="text-xs text-muted" style={{ fontFamily: 'monospace' }}>
                    {log.resource_id}
                  </td>
                  <td className="text-sm text-muted">
                    {log.actor?.full_name || log.actor?.email || log.actor_id || 'Sistema'}
                  </td>
                  <td className="text-xs text-muted" style={{ fontFamily: 'monospace' }}>
                    {log.ip_address || '—'}
                  </td>
                  <td className="text-xs text-muted">
                    {formatDate(log.created_at)} {new Date(log.created_at).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        {/* Pagination */}
        {data && data.total_pages > 1 && (
          <div style={{ padding: '16px 24px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', borderTop: '1px solid var(--color-border)' }}>
            <span className="text-sm text-muted">Página {page} de {data.total_pages}</span>
            <div style={{ display: 'flex', gap: 8 }}>
              <button className="btn btn-secondary btn-sm" onClick={() => setPage(p => p - 1)} disabled={page === 1}>← Anterior</button>
              <button className="btn btn-secondary btn-sm" onClick={() => setPage(p => p + 1)} disabled={page === data.total_pages}>Próxima →</button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
