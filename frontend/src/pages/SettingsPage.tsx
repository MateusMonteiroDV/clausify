import { useState } from 'react';
import { Building, Shield, Users, Key, Save, Check } from 'lucide-react';
import { useAuthStore } from '../store/authStore';

export default function SettingsPage() {
  const user = useAuthStore((s) => s.user);
  const [saved, setSaved] = useState(false);

  const [orgName, setOrgName] = useState(user?.org_name || 'Minha Empresa');
  const [aiModel, setAiModel] = useState('gemini-1.5-pro');
  const [sensitivity, setSensitivity] = useState('HIGH');
  const [webhookUrl, setWebhookUrl] = useState('https://api.empresa.com/webhooks/clausify');

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault();
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  return (
    <div className="animate-fade-in" style={{ maxWidth: 880 }}>
      <div className="page-header">
        <h1 className="page-title">Configurações & Organização</h1>
        <p className="page-subtitle">Gerencie o perfil da sua empresa, equipe e parâmetros de IA</p>
      </div>

      {saved && (
        <div
          style={{
            padding: '12px 16px',
            borderRadius: 'var(--radius-md)',
            background: 'rgba(16,185,129,0.12)',
            border: '1px solid rgba(16,185,129,0.25)',
            color: 'var(--color-success)',
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            marginBottom: 24,
            fontSize: '0.875rem',
            fontWeight: 500,
          }}
        >
          <Check size={16} /> Configurações salvas com sucesso!
        </div>
      )}

      <form onSubmit={handleSave} style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
        {/* Organization Info */}
        <div className="card">
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 20 }}>
            <Building size={20} color="var(--color-primary)" />
            <h2 style={{ fontSize: '1.1rem', fontWeight: 700 }}>Perfil da Organização</h2>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
            <div className="form-group">
              <label className="form-label">Nome da Empresa</label>
              <input
                className="form-input"
                value={orgName}
                onChange={(e) => setOrgName(e.target.value)}
                required
              />
            </div>
            <div className="form-group">
              <label className="form-label">Plano Atual</label>
              <input
                className="form-input"
                value="ENTERPRISE · Documentos Ilimitados"
                disabled
                style={{ opacity: 0.7, cursor: 'not-allowed' }}
              />
            </div>
          </div>
        </div>

        {/* AI & Analysis Engine Settings */}
        <div className="card">
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 20 }}>
            <Shield size={20} color="var(--color-accent)" />
            <h2 style={{ fontSize: '1.1rem', fontWeight: 700 }}>Motor de Análise AI (Gemini)</h2>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginBottom: 16 }}>
            <div className="form-group">
              <label className="form-label">Modelo de IA Principal</label>
              <select className="form-input" value={aiModel} onChange={(e) => setAiModel(e.target.value)}>
                <option value="gemini-1.5-pro">Gemini 1.5 Pro (Recomendado)</option>
                <option value="gemini-1.5-flash">Gemini 1.5 Flash (Ultra rápido)</option>
                <option value="gemini-2.0-flash">Gemini 2.0 Flash (Nova Geração)</option>
              </select>
            </div>

            <div className="form-group">
              <label className="form-label">Sensibilidade de Risco</label>
              <select className="form-input" value={sensitivity} onChange={(e) => setSensitivity(e.target.value)}>
                <option value="HIGH">Rigorosa (Detecta cláusulas ambíguas)</option>
                <option value="BALANCED">Equilibrada (Padrão)</option>
                <option value="RELAXED">Permissiva (Apenas riscos críticos)</option>
              </select>
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">Webhook de Notificações de Análise</label>
            <input
              className="form-input"
              value={webhookUrl}
              onChange={(e) => setWebhookUrl(e.target.value)}
              placeholder="https://..."
            />
            <span className="text-xs text-muted" style={{ marginTop: 4 }}>
              Receba um evento via HTTP sempre que um documento for analisado ou obrigações vencerem.
            </span>
          </div>
        </div>

        {/* Team Members Mock */}
        <div className="card">
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 20 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <Users size={20} color="var(--color-primary)" />
              <h2 style={{ fontSize: '1.1rem', fontWeight: 700 }}>Membros da Equipe</h2>
            </div>
            <button type="button" className="btn btn-secondary btn-sm">
              + Convidar Membro
            </button>
          </div>

          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th>Nome</th>
                  <th>Email</th>
                  <th>Função</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td style={{ fontWeight: 600 }}>{user?.full_name || 'Usuário Admin'}</td>
                  <td className="text-muted">{user?.email}</td>
                  <td><span className="badge badge-analyzed">ADMIN</span></td>
                  <td><span className="badge badge-analyzed">ATIVO</span></td>
                </tr>
                <tr>
                  <td style={{ fontWeight: 600 }}>Auditoria Legal</td>
                  <td className="text-muted">auditoria@{orgName.toLowerCase().replace(/\s+/g, '')}.com</td>
                  <td><span className="badge badge-processing">AUDITOR</span></td>
                  <td><span className="badge badge-analyzed">ATIVO</span></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        {/* Save button */}
        <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
          <button type="submit" className="btn btn-primary" style={{ padding: '12px 32px' }}>
            <Save size={16} /> Salvar Alterações
          </button>
        </div>
      </form>
    </div>
  );
}
