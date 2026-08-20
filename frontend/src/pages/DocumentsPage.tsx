import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { FileText, Plus, Trash2, Eye, Upload } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { documentService } from '../services';
import { StatusBadge, RiskScoreBar, formatDate, formatBytes, Spinner } from '../components/ui';

export default function DocumentsPage() {
  const navigate = useNavigate();
  const qc = useQueryClient();
  const [page, setPage] = useState(1);
  const [showUpload, setShowUpload] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ['documents', page],
    queryFn: () => documentService.list(page, 20),
  });

  const deleteMutation = useMutation({
    mutationFn: documentService.delete,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['documents'] }),
  });

  const handleDelete = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (confirm('Excluir este documento?')) deleteMutation.mutate(id);
  };

  return (
    <div className="animate-fade-in">
      <div className="page-header" style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
        <div>
          <h1 className="page-title">Documentos</h1>
          <p className="page-subtitle">Gerencie e analise seus contratos</p>
        </div>
        <button id="upload-btn" className="btn btn-primary" onClick={() => setShowUpload(true)}>
          <Plus size={16} />
          Novo Documento
        </button>
      </div>

      <div className="table-container">
        <div className="table-header">
          <span className="table-title">
            {data?.total != null && <span style={{ color: 'var(--color-text-muted)', fontWeight: 400 }}> ({data.total} total)</span>}
          </span>
        </div>

        {isLoading ? (
          <div style={{ padding: 60, display: 'flex', justifyContent: 'center' }}><Spinner /></div>
        ) : !data?.data?.length ? (
          <div className="empty-state">
            <div className="empty-state-icon"><FileText size={30} /></div>
            <div className="empty-state-title">Nenhum documento</div>
            <div className="empty-state-text">Faça upload de um contrato PDF para começar</div>
          </div>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Documento</th>
                <th>Tamanho</th>
                <th>Páginas</th>
                <th>Status</th>
                <th>Score de Risco</th>
                <th>Data</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {data.data.map((doc) => (
                <tr key={doc.id} style={{ cursor: 'pointer' }} onClick={() => navigate(`/documents/${doc.id}`)}>
                  <td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                      <div style={{
                        width: 32, height: 32, borderRadius: 'var(--radius-sm)',
                        background: 'var(--color-primary-dim)', display: 'flex',
                        alignItems: 'center', justifyContent: 'center', flexShrink: 0
                      }}>
                        <FileText size={15} color="var(--color-primary)" />
                      </div>
                      <div>
                        <div style={{ fontWeight: 500, maxWidth: 200 }} className="truncate">{doc.file_name}</div>
                        <div className="text-xs text-muted">{doc.mime_type}</div>
                      </div>
                    </div>
                  </td>
                  <td className="text-sm text-muted">{formatBytes(doc.file_size_bytes)}</td>
                  <td className="text-sm text-muted">{doc.page_count || '—'}</td>
                  <td><StatusBadge status={doc.status} /></td>
                  <td><RiskScoreBar score={doc.risk_score} /></td>
                  <td className="text-sm text-muted">{formatDate(doc.created_at)}</td>
                  <td>
                    <div style={{ display: 'flex', gap: 6 }} onClick={(e) => e.stopPropagation()}>
                      <button className="btn btn-icon btn-secondary btn-sm"
                        onClick={() => navigate(`/documents/${doc.id}`)} title="Ver detalhes">
                        <Eye size={14} />
                      </button>
                      <button className="btn btn-icon btn-danger btn-sm"
                        onClick={(e) => handleDelete(doc.id, e)} title="Excluir">
                        <Trash2 size={14} />
                      </button>
                    </div>
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

      {showUpload && <UploadModal onClose={() => setShowUpload(false)} />}
    </div>
  );
}

function UploadModal({ onClose }: { onClose: () => void }) {
  const qc = useQueryClient();
  const [dragOver, setDragOver] = useState(false);
  const [file, setFile] = useState<File | null>(null);
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState('');

  const mutation = useMutation({
    mutationFn: (f: File) => documentService.upload(f, setProgress),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['documents'] }); onClose(); },
    onError: (e: any) => setError(e.response?.data?.error || 'Erro ao fazer upload'),
  });

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    const dropped = e.dataTransfer.files[0];
    if (dropped) { setFile(dropped); setError(''); }
  };

  const handleFile = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0];
    if (f) { setFile(f); setError(''); }
  };

  const handleSubmit = () => {
    if (!file) { setError('Selecione um arquivo'); return; }
    if (file.size > 50 * 1024 * 1024) { setError('Arquivo maior que 50 MB'); return; }
    setProgress(0);
    mutation.mutate(file);
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 480 }}>
        <div className="modal-header">
          <span className="modal-title">
            <Upload size={18} style={{ marginRight: 8, verticalAlign: 'middle' }} />
            Upload de Documento
          </span>
          <button className="btn btn-icon btn-secondary btn-sm" onClick={onClose}>✕</button>
        </div>

        {/* Drop zone */}
        <div
          onDragOver={(e) => { e.preventDefault(); setDragOver(true); }}
          onDragLeave={() => setDragOver(false)}
          onDrop={handleDrop}
          style={{
            border: `2px dashed ${dragOver ? 'var(--color-primary)' : 'var(--color-border)'}`,
            borderRadius: 'var(--radius-md)',
            padding: '32px 24px',
            textAlign: 'center',
            background: dragOver ? 'var(--color-primary-dim)' : 'var(--color-surface)',
            transition: 'all 0.2s',
            cursor: 'pointer',
            marginBottom: 16,
          }}
          onClick={() => document.getElementById('file-input')?.click()}
        >
          <input id="file-input" type="file" accept=".pdf,.doc,.docx,.txt" style={{ display: 'none' }} onChange={handleFile} />
          <FileText size={36} color="var(--color-primary)" style={{ marginBottom: 12, opacity: 0.8 }} />
          {file ? (
            <>
              <div style={{ fontWeight: 600, marginBottom: 4 }}>{file.name}</div>
              <div className="text-sm text-muted">{(file.size / 1024 / 1024).toFixed(2)} MB</div>
            </>
          ) : (
            <>
              <div style={{ fontWeight: 500, marginBottom: 4 }}>Arraste o arquivo aqui</div>
              <div className="text-sm text-muted">ou clique para selecionar · PDF, DOCX, TXT · máx 50 MB</div>
            </>
          )}
        </div>

        {/* Progress bar */}
        {mutation.isPending && (
          <div style={{ marginBottom: 16 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 6 }}>
              <span className="text-sm text-muted">Enviando...</span>
              <span className="text-sm" style={{ fontWeight: 600 }}>{progress}%</span>
            </div>
            <div style={{ height: 6, borderRadius: 99, background: 'var(--color-border)', overflow: 'hidden' }}>
              <div style={{
                height: '100%', width: `${progress}%`,
                background: 'var(--color-primary)',
                borderRadius: 99,
                transition: 'width 0.2s',
              }} />
            </div>
          </div>
        )}

        {error && <div className="error-msg" style={{ marginBottom: 14 }}>{error}</div>}

        <div className="flex gap-4">
          <button className="btn btn-secondary flex-1" style={{ justifyContent: 'center' }} onClick={onClose} disabled={mutation.isPending}>
            Cancelar
          </button>
          <button
            id="doc-upload-submit"
            className="btn btn-primary flex-1"
            style={{ justifyContent: 'center' }}
            onClick={handleSubmit}
            disabled={mutation.isPending || !file}
          >
            {mutation.isPending ? <Spinner size={16} /> : <><Upload size={15} /> Enviar</>}
          </button>
        </div>
      </div>
    </div>
  );
}

