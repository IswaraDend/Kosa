import { useEffect, useState } from 'react';
import { Plus, KeyRound, Copy, Eye, Ban } from 'lucide-react';
import { api, ApiError, buildQuery, type PaginatedResponse } from '../../lib/api';
import type { ProjectApiKey } from '../../types';
import { useSelectedProject } from '../../hooks/useSelectedProject';
import { useProjectAutoSelect } from '../../hooks/useProjectAutoSelect';
import { usePagination, metaFrom, PER_PAGE } from '../../hooks/usePagination';
import PageHeader from '../../components/PageHeader';
import StatCard from '../../components/StatCard';
import TableCard from '../../components/TableCard';
import Pagination from '../../components/Pagination';
import Modal from '../../components/Modal';
import ConfirmDialog from '../../components/ConfirmDialog';
import ProjectPicker from '../../components/ProjectPicker';
import Badge from '../../components/Badge';
import '../Dashboard.css';

/** `kosa_sk_ab12••••` — what the list shows for a key nobody is revealing. */
const maskedKey = (prefix: string) => `kosa_sk_${prefix}••••••••`;

interface ApiKeyPageProps {
  /** '/super-admin' (global, ?project_id) or '/admin/projects' (path-scoped). */
  apiBasePrefix?: string;
  scopeMode?: 'query' | 'path';
  projectEndpoint?: string;
  includeAllOption?: boolean;
  showProjectPicker?: boolean;
  /**
   * Whether this dashboard may read an existing key back in plaintext. Super
   * Admin only — the backend refuses it on the /admin routes regardless, so
   * this just keeps a button off the screen that would always fail.
   */
  canReveal?: boolean;
}

const ApiKeyPage = ({
  apiBasePrefix = '/super-admin',
  scopeMode = 'query',
  projectEndpoint = '/super-admin/projects',
  includeAllOption = false,
  showProjectPicker = true,
  canReveal = true,
}: ApiKeyPageProps) => {
  const { selectedProjectId, setSelectedProjectId } = useSelectedProject();
  useProjectAutoSelect(projectEndpoint, !showProjectPicker);
  const { page, setPage, meta, setMeta } = usePagination(selectedProjectId);
  const [keys, setKeys] = useState<ProjectApiKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState('');

  const [formOpen, setFormOpen] = useState(false);
  const [name, setName] = useState('');
  const [saving, setSaving] = useState(false);
  const [confirmRevoke, setConfirmRevoke] = useState<ProjectApiKey | null>(null);

  // The plaintext currently on screen, either just created or just revealed.
  // Kept in state and never written anywhere else — closing the dialog is all
  // it takes for it to be gone from the page.
  const [secret, setSecret] = useState<{ value: string; fresh: boolean } | null>(null);
  const [copied, setCopied] = useState(false);

  // Keys are addressed globally on the Super Admin routes and per project on
  // the Admin ones, so unlike every other page here the two scopes do not
  // share one URL shape.
  const listPath = scopeMode === 'path' ? `${apiBasePrefix}/${selectedProjectId}/api-keys` : `${apiBasePrefix}/api-keys`;
  const createPath =
    scopeMode === 'path'
      ? `${apiBasePrefix}/${selectedProjectId}/api-keys`
      : `${apiBasePrefix}/projects/${selectedProjectId}/api-keys`;
  const keyPath = (id: number) =>
    scopeMode === 'path' ? `${apiBasePrefix}/${selectedProjectId}/api-keys/${id}` : `${apiBasePrefix}/api-keys/${id}`;

  const pagedQuery = buildQuery({
    ...(scopeMode === 'query' ? { project_id: selectedProjectId } : {}),
    page,
    per_page: PER_PAGE,
  });

  const loadKeys = () => {
    if (selectedProjectId === 'all') {
      setKeys([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    api
      .get<PaginatedResponse<ProjectApiKey>>(`${listPath}${pagedQuery}`)
      .then((res) => {
        setKeys(res.data);
        setMeta(metaFrom(res));
      })
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat data secret key'))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadKeys();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId, apiBasePrefix, scopeMode, page]);

  const handleCreate = async () => {
    if (!name.trim()) return;
    setSaving(true);
    setErrorMsg('');
    try {
      const res = await api.post<{ secret: string }>(createPath, { name: name.trim() });
      setFormOpen(false);
      setName('');
      setSecret({ value: res.secret, fresh: true });
      setCopied(false);
      loadKeys();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal membuat secret key');
    } finally {
      setSaving(false);
    }
  };

  const handleReveal = async (key: ProjectApiKey) => {
    setErrorMsg('');
    try {
      const res = await api.post<{ secret: string }>(`${keyPath(key.id)}/reveal`, {});
      setSecret({ value: res.secret, fresh: false });
      setCopied(false);
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal menampilkan secret key');
    }
  };

  const handleRevoke = async () => {
    if (!confirmRevoke) return;
    try {
      await api.delete(keyPath(confirmRevoke.id));
      setConfirmRevoke(null);
      loadKeys();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal mencabut secret key');
      setConfirmRevoke(null);
    }
  };

  const copySecret = async () => {
    if (!secret) return;
    try {
      await navigator.clipboard.writeText(secret.value);
      setCopied(true);
    } catch {
      // Clipboard access needs a secure context, so this fails on plain HTTP.
      // The value is on screen and selectable either way.
      setErrorMsg('Tidak bisa menyalin otomatis, salin manual dari kotak di atas.');
    }
  };

  const activeCount = keys.filter((k) => k.status === 'active').length;

  return (
    <div className="dashboard-content">
      <PageHeader
        title="Secret Key"
        subtitle="Kunci untuk memanggil API Kosa dari luar, satu atau lebih per project"
        actions={
          <button
            className="btn-primary"
            onClick={() => {
              setName('');
              setFormOpen(true);
            }}
            disabled={selectedProjectId === 'all'}
          >
            <Plus size={18} />
            Buat Secret Key
          </button>
        }
      />

      {showProjectPicker && (
        <div className="form-group" style={{ maxWidth: 280 }}>
          <label>Project</label>
          <ProjectPicker
            value={selectedProjectId}
            onChange={setSelectedProjectId}
            includeAllOption={includeAllOption}
            endpoint={projectEndpoint}
          />
        </div>
      )}

      {errorMsg && (
        <div style={{ color: 'var(--danger)', marginBottom: '16px', fontSize: '14px' }}>{errorMsg}</div>
      )}

      {selectedProjectId === 'all' ? (
        <p style={{ color: 'var(--text-muted)' }}>Pilih project terlebih dahulu untuk melihat secret key-nya.</p>
      ) : (
        <>
          <div className="summary-cards" style={{ gridTemplateColumns: 'repeat(2, 1fr)' }}>
            <StatCard label="Total Secret Key" value={meta.total} icon={<KeyRound size={18} />} />
            <StatCard label="Masih Aktif" value={activeCount} icon={<KeyRound size={18} />} />
          </div>

          <TableCard title="Daftar Secret Key" count={meta.total}>
            <thead>
              <tr>
                <th>NAMA</th>
                <th>KEY</th>
                <th>STATUS</th>
                <th>TERAKHIR DIPAKAI</th>
                <th>DIBUAT</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {!loading && keys.length === 0 && (
                <tr>
                  <td colSpan={6} style={{ color: 'var(--text-muted)' }}>
                    Belum ada secret key untuk project ini.
                  </td>
                </tr>
              )}
              {!loading &&
                keys.map((key) => (
                  <tr key={key.id}>
                    <td>{key.name}</td>
                    <td style={{ color: 'var(--text-muted)', fontFamily: 'monospace', fontSize: 12 }}>
                      {maskedKey(key.key_prefix)}
                    </td>
                    <td>
                      <Badge
                        label={key.status === 'active' ? 'Aktif' : 'Dicabut'}
                        variant={key.status === 'active' ? 'aktif' : 'keluar'}
                      />
                    </td>
                    <td style={{ color: 'var(--text-muted)' }}>
                      {key.last_used_at ? new Date(key.last_used_at).toLocaleString('id-ID') : 'Belum pernah'}
                    </td>
                    <td style={{ color: 'var(--text-muted)' }}>
                      {new Date(key.created_at).toLocaleDateString('id-ID')}
                    </td>
                    <td style={{ textAlign: 'right', display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
                      {canReveal && (
                        <button
                          style={{ background: 'transparent', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}
                          onClick={() => handleReveal(key)}
                          title="Lihat secret key (tercatat di audit)"
                        >
                          <Eye size={16} />
                        </button>
                      )}
                      {key.status === 'active' && (
                        <button
                          style={{ background: 'transparent', border: 'none', color: 'var(--danger)', cursor: 'pointer' }}
                          onClick={() => setConfirmRevoke(key)}
                          title="Cabut secret key"
                        >
                          <Ban size={16} />
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
            </tbody>
          </TableCard>

          <Pagination page={page} meta={meta} onChange={setPage} label="secret key" />
        </>
      )}

      <Modal
        isOpen={formOpen}
        onClose={() => setFormOpen(false)}
        title="Buat Secret Key"
        footer={
          <>
            <button className="btn-secondary" onClick={() => setFormOpen(false)}>
              Batal
            </button>
            <button className="btn-primary" onClick={handleCreate} disabled={saving || !name.trim()}>
              {saving ? 'Membuat...' : 'Buat'}
            </button>
          </>
        }
      >
        <p style={{ fontSize: 13, color: 'var(--text-muted)', marginBottom: 12 }}>
          Nama hanya untuk Anda sendiri — supaya jelas kunci ini dipakai siapa. Satu project boleh punya lebih
          dari satu kunci aktif, jadi rotasi tidak perlu membuat pihak luar mati dulu.
        </p>
        <div className="form-group">
          <label>Nama</label>
          <input
            className="form-input"
            placeholder="mis. Landing page, Storefront Shopee"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
        </div>
      </Modal>

      <Modal
        isOpen={!!secret}
        onClose={() => setSecret(null)}
        title={secret?.fresh ? 'Secret Key Baru' : 'Secret Key'}
        footer={
          <button className="btn-secondary" onClick={() => setSecret(null)}>
            Tutup
          </button>
        }
      >
        <p style={{ fontSize: 13, color: 'var(--text-muted)', marginBottom: 12 }}>
          {secret?.fresh
            ? 'Simpan dan kirimkan ke pihak yang akan memakainya. Kunci ini bisa dilihat lagi nanti oleh Super Admin, dan setiap kali dilihat tercatat di jejak audit.'
            : 'Setiap kali kunci ini ditampilkan, siapa dan kapan tercatat di jejak audit.'}
        </p>
        <div
          style={{
            padding: '12px',
            background: 'var(--bg-color)',
            border: '1px solid var(--border-color)',
            borderRadius: 8,
            fontFamily: 'monospace',
            fontSize: 12,
            wordBreak: 'break-all',
            marginBottom: 12,
          }}
        >
          {secret?.value}
        </div>
        <button className="btn-primary" onClick={copySecret}>
          <Copy size={16} />
          {copied ? 'Tersalin' : 'Salin'}
        </button>
      </Modal>

      <ConfirmDialog
        isOpen={!!confirmRevoke}
        title="Cabut Secret Key"
        message={`Cabut "${confirmRevoke?.name}"? Pemanggil yang masih memakainya akan langsung ditolak, dan kunci ini tidak bisa diaktifkan lagi.`}
        confirmLabel="Cabut"
        onConfirm={handleRevoke}
        onCancel={() => setConfirmRevoke(null)}
      />
    </div>
  );
};

export default ApiKeyPage;
