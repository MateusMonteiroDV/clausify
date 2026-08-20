import { NavLink, useNavigate } from 'react-router-dom';
import { LayoutDashboard, FileText, AlertTriangle, Settings, Activity, LogOut, Shield } from 'lucide-react';
import { useAuthStore } from '../store/authStore';

const navItems = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/documents', icon: FileText, label: 'Documentos' },
  { to: '/obligations', icon: AlertTriangle, label: 'Obrigações' },
  { to: '/audit-log', icon: Activity, label: 'Auditoria' },
  { to: '/settings', icon: Settings, label: 'Configurações' },
];

export default function Sidebar() {
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const initials = user?.full_name
    ? user.full_name.split(' ').map((n) => n[0]).slice(0, 2).join('')
    : user?.email?.[0]?.toUpperCase() ?? 'U';

  return (
    <aside className="sidebar">
      <div className="sidebar-logo">
        <div className="sidebar-logo-icon">
          <Shield size={20} color="white" />
        </div>
        <span className="sidebar-logo-text">Clausify</span>
      </div>

      <nav className="sidebar-nav">
        {navItems.map(({ to, icon: Icon, label }) => (
          <NavLink
            key={to}
            to={to}
            end={to === '/'}
            className={({ isActive }) => `sidebar-nav-item${isActive ? ' active' : ''}`}
          >
            <Icon size={17} className="nav-icon" />
            {label}
          </NavLink>
        ))}
      </nav>

      <div className="sidebar-footer">
        <div className="sidebar-user">
          <div className="sidebar-avatar">{initials}</div>
          <div className="sidebar-user-info">
            <div className="sidebar-user-name">{user?.full_name || user?.email}</div>
            <div className="sidebar-user-role">{user?.org_name}</div>
          </div>
        </div>
        <button
          onClick={handleLogout}
          className="sidebar-nav-item w-full"
          style={{ marginTop: 4, color: 'var(--color-danger)' }}
        >
          <LogOut size={17} />
          Sair
        </button>
      </div>
    </aside>
  );
}
