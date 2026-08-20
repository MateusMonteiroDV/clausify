import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { AlertTriangle, Plus, CheckCircle, Trash2 } from 'lucide-react';
import { obligationService } from '../services';
import { ObligationBadge, formatDate, Spinner } from '../components/ui';
import type { ContractObligation } from '../types';

export default function ObligationsPage() {
  const qc = useQueryClient();
  const [page, setPage] = useState(1);
  const [showCreate, setShowCreate] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ['obligations', page],
    queryFn: () => obligationService.list(page, 20),
  });

  const completeMutation = useMutation({
    mutationFn: (id: string) => obligationService.update(id, { status: 'COMPLETED' }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['obligations'] }),
  });

  const deleteMutation = useMutation({
    mutationFn: obligationService.delete,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['obligations'] }),
  });

  const groupByStatus = (items: ContractObligation[]) => {
    const overdue  = items.filter((o) => o.status === 'OVERDUE' || (o.status === 'PENDING' && new Date(o.due_date) < new Date()));
    const pending  = items.filter((o) => o.status === 'PENDING' && new Date(o.due_date) >= new Date());
    const notified = items.filter((o) => o.status === 'NOTIFIED');
    const completed = items.filter((o) => o.status === 'COMPLETED');
    return { overdue, pending, notified, completed };
  };

  const groups = data?.data ? groupByStatus(data.data) : { overdue: [], pending: [], notified: [], completed: [] };

  return (
    <div className="animate-fade-in">
      <div className="page-header" style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
        <div>
          <h1 className="page-title">Obrigações Contratuais</h1>
          <p className="page-subtitle">Acompanhe deadlines, renovações e obrigações pendentes</p>
        </div>
        <button id="create-obligation-btn" className="btn btn-primary" onClick={() => setShowCreate(true)}>
          <Plus size={16} />
          Nova Obrigação
        </button>
      </div>

      {isLoading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: 80 }}><Spinner /></div>
      ) : !data?.data?.length ? (
        <div className="empty-state" style={{ marginTop: 40 }}>
          <div className="empty-state-icon"><AlertTriangle size={30} /></div>
          <div className="empty-state-title">Nenhuma obrigação rastreada</div>
          <div className="empty-state-text">Adicione obrigações manualmente ou deixe o sistema extrair automaticamente dos contratos</div>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 28 }}>
          {groups.overdue.length > 0 && (
            <Section title="⚠️ Vencidas" color="var(--color-danger)"
              items={groups.overdue} onComplete={completeMutation.mutate} onDelete={deleteMutation.mutate} />
          )}
          {groups.pending.length > 0 && (
            <Section title="⏳ Pendentes" color="var(--color-warning)"
              items={groups.pending} onComplete={completeMutation.mutate} onDelete={deleteMutation.mutate} />
          )}
          {groups.notified.length > 0 && (
            <Section title="🔔 Notificadas" color="var(--color-info)"
              items={groups.notified} onComplete={completeMutation.mutate} onDelete={deleteMutation.mutate} />
          )}
          {groups.completed.length > 0 && (
            <Section title="✅ Concluídas" color="var(--color-success)"
              items={groups.completed} onComplete={completeMutation.mutate} onDelete={deleteMutation.mutate} />
          )}
        </div>
      )}

      {data && data.total_pages > 1 && (
        <div style={{ display: 'flex', justifyContent: 'center', gap: 8, marginTop: 24 }}>
          <button className="btn btn-secondary btn-sm" onClick={() => setPage(p => p - 1)} disabled={page === 1}>← Anterior</button>
          <span className="text-sm text-muted" style={{ padding: '6px 12px' }}>Página {page}</span>
          <button className="btn btn-secondary btn-sm" onClick={() => setPage(p => p + 1)} disabled={page === data.total_pages}>Próxima →</button>
        </div>
      )}

      {showCreate && <CreateObligationModal onClose={() => setShowCreate(false)} />}
    </div>
  );
}

function Section({ title, color, items, onComplete, onDelete }: {
  title: string;
  color: string;
  items: ContractObligation[];
  onComplete: (id: string) => void;
  onDelete: (id: string) => void;
}) {
  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 12 }}>
        <h2 style={{ fontSize: '0.9rem', fontWeight: 700, color }}>{title}</h2>
        <span style={{
          background: `${color}22`, color, border: `1px solid ${color}44`,
          padding: '1px 8px', borderRadius: 100, fontSize: '0.72rem', fontWeight: 700
        }}>{items.length}</span>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
        {items.map((o) => (
          <div key={o.id} className="obligation-card">
            <div className="obligation-due-dot" style={{ background: color }} />
            <div className="obligation-info">
              <div className="obligation-title">{o.title}</div>
              {o.description && <div className="text-xs text-muted" style={{ marginBottom: 2 }}>{o.description}</div>}
              <div className="obligation-meta">
                <span>Vence: <strong>{formatDate(o.due_date)}</strong></span>
                {o.is_recurring && <span>↻ {o.recurrence_interval}</span>}
                {o.notes && <span title={o.notes}>📝 Nota</span>}
              </div>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <ObligationBadge status={o.status} />
              {o.status !== 'COMPLETED' && (
                <button className="btn btn-icon btn-secondary btn-sm" title="Marcar como concluída"
                  onClick={() => onComplete(o.id)}>
                  <CheckCircle size={14} />
                </button>
              )}
              <button className="btn btn-icon btn-danger btn-sm" title="Excluir"
                onClick={() => confirm('Excluir?') && onDelete(o.id)}>
                <Trash2 size={14} />
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function CreateObligationModal({ onClose }: { onClose: () => void }) {
  const qc = useQueryClient();
  const [form, setForm] = useState({
    document_id: '', title: '', description: '', due_date: '',
    is_recurring: false, recurrence_interval: '', notes: ''
  });
  const [error, setError] = useState('');

  const mutation = useMutation({
    mutationFn: () => obligationService.create({
      ...form,
      recurrence_interval: form.is_recurring && form.recurrence_interval ? form.recurrence_interval : undefined,
    }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['obligations'] }); onClose(); },
    onError: (e: any) => setError(e.response?.data?.error || 'Erro ao criar obrigação'),
  });

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" style={{ maxWidth: 520 }} onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <span className="modal-title"><AlertTriangle size={18} style={{ marginRight: 8, verticalAlign: 'middle' }} />Nova Obrigação</span>
          <button className="btn btn-icon btn-secondary btn-sm" onClick={onClose}>✕</button>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          {error && <div className="error-msg">{error}</div>}
          <div className="form-group">
            <label className="form-label">ID do Documento (UUID)</label>
            <input id="oblig-doc" className="form-input" placeholder="uuid-do-documento"
              value={form.document_id} onChange={(e) => setForm(p => ({ ...p, document_id: e.target.value }))} />
          </div>
          <div className="form-group">
            <label className="form-label">Título *</label>
            <input id="oblig-title" className="form-input" placeholder="Renovação anual do contrato"
              value={form.title} onChange={(e) => setForm(p => ({ ...p, title: e.target.value }))} required />
          </div>
          <div className="form-group">
            <label className="form-label">Descrição</label>
            <input className="form-input" placeholder="Detalhes adicionais..."
              value={form.description} onChange={(e) => setForm(p => ({ ...p, description: e.target.value }))} />
          </div>
          <div className="grid-2">
            <div className="form-group">
              <label className="form-label">Data de vencimento *</label>
              <input id="oblig-date" className="form-input" type="date"
                value={form.due_date} onChange={(e) => setForm(p => ({ ...p, due_date: e.target.value }))} required />
            </div>
            <div className="form-group">
              <label className="form-label">Recorrência</label>
              <select className="form-input"
                value={form.recurrence_interval}
                onChange={(e) => setForm(p => ({ ...p, recurrence_interval: e.target.value, is_recurring: !!e.target.value }))}>
                <option value="">Não recorrente</option>
                <option value="MONTHLY">Mensal</option>
                <option value="QUARTERLY">Trimestral</option>
                <option value="ANNUAL">Anual</option>
              </select>
            </div>
          </div>
          <div className="form-group">
            <label className="form-label">Notas</label>
            <input className="form-input" placeholder="Observações..."
              value={form.notes} onChange={(e) => setForm(p => ({ ...p, notes: e.target.value }))} />
          </div>
          <div className="flex gap-4" style={{ marginTop: 4 }}>
            <button className="btn btn-secondary flex-1" style={{ justifyContent: 'center' }} onClick={onClose}>Cancelar</button>
            <button id="oblig-submit" className="btn btn-primary flex-1" style={{ justifyContent: 'center' }}
              onClick={() => mutation.mutate()} disabled={mutation.isPending}>
              {mutation.isPending ? <Spinner size={16} /> : 'Criar Obrigação'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
