import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { Shield, Eye, EyeOff } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { authService } from '../services';
import { useAuthStore } from '../store/authStore';
import { Spinner } from '../components/ui';

type Tab = 'login' | 'register';

export default function AuthPage() {
  const [tab, setTab] = useState<Tab>('login');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();
  const login = useAuthStore((s) => s.login);

  // Login
  const [loginForm, setLoginForm] = useState({ email: '', password: '' });
  const loginMutation = useMutation({
    mutationFn: () => authService.login(loginForm),
    onSuccess: (data) => { login(data.token, data.user); navigate('/'); },
    onError: (e: any) => setError(e.response?.data?.error || 'Erro ao entrar'),
  });

  // Register
  const [regForm, setRegForm] = useState({ org_name: '', full_name: '', email: '', password: '' });
  const regMutation = useMutation({
    mutationFn: () => authService.register(regForm),
    onSuccess: (data) => { login(data.token, data.user); navigate('/'); },
    onError: (e: any) => setError(e.response?.data?.error || 'Erro ao criar conta'),
  });

  const handleTab = (t: Tab) => { setTab(t); setError(''); };

  return (
    <div className="auth-page">
      <div className="auth-card animate-fade-in">
        <div className="auth-logo">
          <div className="auth-logo-icon">
            <Shield size={22} color="white" />
          </div>
          <span className="auth-logo-text">Clausify</span>
        </div>

        {/* Tabs */}
        <div className="tabs" style={{ marginBottom: 28 }}>
          <button className={`tab${tab === 'login' ? ' active' : ''}`} onClick={() => handleTab('login')}>
            Entrar
          </button>
          <button className={`tab${tab === 'register' ? ' active' : ''}`} onClick={() => handleTab('register')}>
            Criar conta
          </button>
        </div>

        {tab === 'login' ? (
          <>
            <p className="auth-subtitle">Bem-vindo de volta. Entre na sua conta.</p>
            <form className="auth-form" onSubmit={(e) => { e.preventDefault(); setError(''); loginMutation.mutate(); }}>
              {error && <div className="error-msg">{error}</div>}
              <div className="form-group">
                <label className="form-label">E-mail</label>
                <input id="login-email" className="form-input" type="email" placeholder="voce@empresa.com"
                  value={loginForm.email}
                  onChange={(e) => setLoginForm((p) => ({ ...p, email: e.target.value }))} required />
              </div>
              <div className="form-group">
                <label className="form-label">Senha</label>
                <div style={{ position: 'relative' }}>
                  <input id="login-password" className="form-input" type={showPassword ? 'text' : 'password'}
                    placeholder="••••••••" value={loginForm.password}
                    onChange={(e) => setLoginForm((p) => ({ ...p, password: e.target.value }))}
                    style={{ paddingRight: 42 }} required />
                  <button type="button" onClick={() => setShowPassword(!showPassword)}
                    style={{ position: 'absolute', right: 12, top: '50%', transform: 'translateY(-50%)', background: 'none', border: 'none', color: 'var(--color-text-muted)' }}>
                    {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                  </button>
                </div>
              </div>
              <button id="login-submit" type="submit" className="btn btn-primary w-full"
                style={{ justifyContent: 'center', marginTop: 4 }}
                disabled={loginMutation.isPending}>
                {loginMutation.isPending ? <Spinner size={16} /> : 'Entrar'}
              </button>
            </form>
          </>
        ) : (
          <>
            <p className="auth-subtitle">Crie sua organização e comece a analisar contratos.</p>
            <form className="auth-form" onSubmit={(e) => { e.preventDefault(); setError(''); regMutation.mutate(); }}>
              {error && <div className="error-msg">{error}</div>}
              <div className="form-group">
                <label className="form-label">Nome da organização</label>
                <input id="reg-org" className="form-input" type="text" placeholder="Acme Ltda."
                  value={regForm.org_name} onChange={(e) => setRegForm((p) => ({ ...p, org_name: e.target.value }))} required />
              </div>
              <div className="form-group">
                <label className="form-label">Seu nome</label>
                <input id="reg-name" className="form-input" type="text" placeholder="João Silva"
                  value={regForm.full_name} onChange={(e) => setRegForm((p) => ({ ...p, full_name: e.target.value }))} required />
              </div>
              <div className="form-group">
                <label className="form-label">E-mail</label>
                <input id="reg-email" className="form-input" type="email" placeholder="voce@empresa.com"
                  value={regForm.email} onChange={(e) => setRegForm((p) => ({ ...p, email: e.target.value }))} required />
              </div>
              <div className="form-group">
                <label className="form-label">Senha</label>
                <input id="reg-password" className="form-input" type="password" placeholder="Mínimo 8 caracteres"
                  value={regForm.password} onChange={(e) => setRegForm((p) => ({ ...p, password: e.target.value }))} minLength={8} required />
              </div>
              <button id="reg-submit" type="submit" className="btn btn-primary w-full"
                style={{ justifyContent: 'center', marginTop: 4 }} disabled={regMutation.isPending}>
                {regMutation.isPending ? <Spinner size={16} /> : 'Criar conta'}
              </button>
            </form>
          </>
        )}
      </div>
    </div>
  );
}
