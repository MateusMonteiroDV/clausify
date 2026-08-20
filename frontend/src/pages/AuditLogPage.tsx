import { useState } from 'react';
import { Activity, ShieldAlert, FileText, UserCheck, RefreshCw, Filter } from 'lucide-react';
import { formatDate } from '../components/ui';

interface AuditLogEntry {
  id: string;
  action: string;
  user_email: string;
  resource: string;
  ip_address: string;
  created_at: string;
  severity: 'INFO' | 'WARNING' | 'CRITICAL';
}

const mockLogs: AuditLogEntry[] = [
  {
    id: '1',
    action: 'DOCUMENT_ANALYZED',
    user_email: 'mateus@clausify.io',
    resource: 'Contrato_Prestacao_Servicos_v2.pdf',
    ip_address: '189.120.45.12',
    created_at: new Date().toISOString(),
    severity: 'INFO',
  },
  {
    id: '2',
    action: 'HIGH_RISK_DETECTED',
    user_email: 'system.ai@clausify.io',
    resource: 'Acordo_NDA_Confidencialidade.pdf (Cláusula Penal $500k)',
    ip_address: '10.0.4.1',
    created_at: new Date(Date.now() - 1000 * 60 * 45).toISOString(),
    severity: 'CRITICAL',
  },
  {
    id: '3',
    action: 'OBLIGATION_COMPLETED',
    user_email: 'mateus@clausify.io',
    resource: 'Renovação da Licença de Software',
    ip_address: '189.120.45.12',
    created_at: new Date(Date.now() - 1000 * 60 * 180).toISOString(),
    severity: 'INFO',
  },
  {
    id: '4',
    action: 'DOCUMENT_UPLOADED',
    user_email: 'auditoria@clausify.io',
    resource: 'Aditivo_Contratual_Fornecedor.pdf',
    ip_address: '200.18.90.11',
    created_at: new Date(Date.now() - 1000 * 60 * 360).toISOString(),
    severity: 'INFO',
  },
  {
    id: '5',
    action: 'USER_LOGIN_SUCCESS',
    user_email: 'mateus@clausify.io',
    resource: 'Sessão Web (JWT Autenticado)',
    ip_address: '189.120.45.12',
    created_at: new Date(Date.now() - 1000 * 60 * 720).toISOString(),
    severity: 'INFO',
  },
];

export default function AuditLogPage() {
  const [filterAction, setFilterAction] = useState('');

  const filteredLogs = mockLogs.filter((log) =>
    !filterAction || log.action.includes(filterAction) || log.resource.toLowerCase().includes(filterAction.toLowerCase())
  );

  return (
    <div className="animate-fade-in">
      <div className="page-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <h1 className="page-title">Logs de Auditoria & Atividade</h1>
          <p className="page-subtitle">Rastreabilidade completa de ações, uploads e execuções de IA</p>
        </div>
        <button className="btn btn-secondary btn-sm" onClick={() => window.location.reload()}>
          <RefreshCw size={14} /> Atualizar
        </button>
      </div>

      {/* Filter Bar */}
      <div style={{ display: 'flex', gap: 12, marginBottom: 20 }}>
        <div style={{ flex: 1, position: 'relative' }}>
          <input
            className="form-input"
            placeholder="Filtrar por ação ou recurso..."
            value={filterAction}
            onChange={(e) => setFilterAction(e.target.value)}
          />
        </div>
      </div>

      {/* Logs Table */}
      <div className="table-container">
        <div className="table-header">
          <span className="table-title">Histórico de Eventos ({filteredLogs.length})</span>
        </div>

        <table>
          <thead>
            <tr>
              <th>Evento / Ação</th>
              <th>Recurso</th>
              <th>Usuário</th>
              <th>IP</th>
              <th>Data & Hora</th>
            </tr>
          </thead>
          <tbody>
            {filteredLogs.map((log) => (
              <tr key={log.id}>
                <td>
                  <span
                    className={`badge ${
                      log.severity === 'CRITICAL' ? 'badge-critical' : 'badge-analyzed'
                    }`}
                  >
                    {log.action}
                  </span>
                </td>
                <td style={{ fontWeight: 500 }} className="truncate">
                  {log.resource}
                </td>
                <td className="text-sm text-muted">{log.user_email}</td>
                <td className="text-xs text-muted" style={{ fontFamily: 'monospace' }}>
                  {log.ip_address}
                </td>
                <td className="text-xs text-muted">
                  {formatDate(log.created_at)} {new Date(log.created_at).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
