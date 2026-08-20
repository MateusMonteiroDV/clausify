import type { RiskLevel, DocumentStatus, ObligationStatus } from '../types';

export function RiskBadge({ level }: { level: RiskLevel }) {
  return (
    <span className={`badge badge-${level.toLowerCase()}`}>
      <span style={{
        width: 6, height: 6, borderRadius: '50%',
        background: 'currentColor', display: 'inline-block'
      }} />
      {level}
    </span>
  );
}

export function StatusBadge({ status }: { status: DocumentStatus }) {
  return (
    <span className={`badge badge-${status.toLowerCase()}`}>
      {status === 'PROCESSING' && (
        <span style={{
          width: 6, height: 6, borderRadius: '50%',
          background: 'currentColor', display: 'inline-block',
          animation: 'pulse 1.5s ease infinite'
        }} />
      )}
      {status}
    </span>
  );
}

export function ObligationBadge({ status }: { status: ObligationStatus }) {
  return (
    <span className={`badge badge-${status.toLowerCase()}`}>
      {status}
    </span>
  );
}

export function RiskScoreBar({ score }: { score?: number }) {
  if (score == null) return <span style={{ color: 'var(--color-text-muted)', fontSize: '0.8rem' }}>—</span>;

  const pct = Math.min(score, 100);
  const color = pct >= 75 ? 'var(--risk-critical)'
              : pct >= 50 ? 'var(--risk-high)'
              : pct >= 25 ? 'var(--risk-medium)'
              : 'var(--risk-low)';

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
      <div className="risk-bar">
        <div className="risk-bar-fill" style={{ width: `${pct}%`, background: color }} />
      </div>
      <span style={{ fontSize: '0.8rem', fontWeight: 600, color }}>{score.toFixed(0)}</span>
    </div>
  );
}

export function Spinner({ size = 20 }: { size?: number }) {
  return (
    <div style={{ width: size, height: size, border: '2px solid rgba(255,255,255,0.1)',
      borderTopColor: 'var(--color-primary)', borderRadius: '50%',
      animation: 'spin 0.7s linear infinite' }} />
  );
}

export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('pt-BR', {
    day: '2-digit', month: 'short', year: 'numeric'
  });
}
