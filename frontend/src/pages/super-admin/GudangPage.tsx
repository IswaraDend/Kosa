import { useEffect, useState } from 'react';
import { Plus, Box, Trash2, Pencil } from 'lucide-react';
import { api, ApiError, buildQuery , type PaginatedResponse } from '../../lib/api';
import type { Warehouse } from '../../types';
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
import { makeCan } from '../../lib/permissions';
import '../Dashboard.css';

interface WarehouseFormState {
  name: string;
  code: string;
  address: string;
}

const emptyForm: WarehouseFormState = { name: '', code: '', address: '' };

interface GudangPageProps {
  /** '/super-admin' (global, project scoped via ?project_id query) or '/admin/projects' (path-scoped: /admin/projects/:id/...) */
  apiBasePrefix?: string;
  scopeMode?: 'query' | 'path';
  projectEndpoint?: string;
  includeAllOption?: boolean;
  showProjectPicker?: boolean;
  /** Granted permission codes. Undefined = no per-action gating (Super Admin / Admin). */
  permissions?: string[];
}

const GudangPage = ({
  apiBasePrefix = '/super-admin',
  scopeMode = 'query',
  projectEndpoint = '/super-admin/projects',
  includeAllOption = true,
  showProjectPicker = true,
  permissions,
}: GudangPageProps) => {
  const can = makeCan(permissions);
  const { selectedProjectId, setSelectedProjectId } = useSelectedProject();
  useProjectAutoSelect(projectEndpoint, !showProjectPicker);
  const { page, setPage, meta, setMeta } = usePagination(selectedProjectId);
  const [warehouses, setWarehouses] = useState<Warehouse[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState('');

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Warehouse | null>(null);
  const [form, setForm] = useState<WarehouseFormState>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState<Warehouse | null>(null);

  const resourceBase = scopeMode === 'path' ? `${apiBasePrefix}/${selectedProjectId}` : apiBasePrefix;
  // Separate from listQuery on purpose: listQuery still scopes the auxiliary
  // dropdown fetches (warehouses, items, customers) which must stay complete,
  // while only the main table asks for a page.
  const pagedQuery = buildQuery({
    ...(scopeMode === 'query' ? { project_id: selectedProjectId } : {}),
    page,
    per_page: PER_PAGE,
  });


  const loadWarehouses = () => {
    if (selectedProjectId === 'all') {
      setWarehouses([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    api
      .get<PaginatedResponse<Warehouse>>(`${resourceBase}/warehouses${pagedQuery}`)
      .then((res) => {
        setWarehouses(res.data);
        setMeta(metaFrom(res));
      })
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat data gudang'))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadWarehouses();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId, apiBasePrefix, scopeMode, page]);

  const openCreateForm = () => {
    setEditing(null);
    setForm(emptyForm);
    setFormOpen(true);
  };

  const openEditForm = (w: Warehouse) => {
    setEditing(w);
    setForm({ name: w.name, code: w.code, address: w.address });
    setFormOpen(true);
  };

  const handleSubmit = async () => {
    setSaving(true);
    setErrorMsg('');
    try {
      if (editing) {
        await api.put(`${resourceBase}/warehouses/${editing.id}`, form);
      } else {
        await api.post(`${resourceBase}/warehouses`, { ...form, project_id: Number(selectedProjectId) });
      }
      setFormOpen(false);
      loadWarehouses();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal menyimpan gudang');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!confirmDelete) return;
    try {
      await api.delete(`${resourceBase}/warehouses/${confirmDelete.id}`);
      setConfirmDelete(null);
      loadWarehouses();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal menghapus gudang');
      setConfirmDelete(null);
    }
  };

  return (
    <div className="dashboard-content">
      <PageHeader
        title="Gudang"
        subtitle="Kelola gudang penyimpanan untuk project terpilih"
        actions={
          can('warehouse.create') && (
            <button className="btn-primary" onClick={openCreateForm} disabled={selectedProjectId === 'all'}>
              <Plus size={18} />
              Tambah Gudang
            </button>
          )
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
        <p style={{ color: 'var(--text-muted)' }}>Pilih project terlebih dahulu untuk melihat daftar gudang.</p>
      ) : (
        <>
          <div className="summary-cards" style={{ gridTemplateColumns: 'repeat(2, 1fr)' }}>
            <StatCard label="Total Gudang" value={meta.total} icon={<Box size={18} />} />
            <StatCard label="Gudang Aktif" value={meta.total} icon={<Box size={18} />} />
          </div>

          <TableCard title="Daftar Gudang" count={meta.total}>
            <thead>
              <tr>
                <th>NAMA GUDANG</th>
                <th>KODE</th>
                <th>ALAMAT</th>
                <th>DIBUAT</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {!loading &&
                warehouses.map((w) => (
                  <tr key={w.id}>
                    <td>{w.name}</td>
                    <td style={{ color: 'var(--text-muted)' }}>{w.code}</td>
                    <td style={{ color: 'var(--text-muted)' }}>{w.address}</td>
                    <td style={{ color: 'var(--text-muted)' }}>
                      {new Date(w.created_at).toLocaleDateString('id-ID')}
                    </td>
                    <td style={{ textAlign: 'right', display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
                      {can('warehouse.update') && (
                        <button
                          style={{ background: 'transparent', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}
                          onClick={() => openEditForm(w)}
                        >
                          <Pencil size={16} />
                        </button>
                      )}
                      {can('warehouse.delete') && (
                        <button
                          style={{ background: 'transparent', border: 'none', color: 'var(--danger)', cursor: 'pointer' }}
                          onClick={() => setConfirmDelete(w)}
                        >
                          <Trash2 size={16} />
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
            </tbody>
          </TableCard>

          <Pagination page={page} meta={meta} onChange={setPage} label="gudang" />
        </>
      )}

      <Modal
        isOpen={formOpen}
        onClose={() => setFormOpen(false)}
        title={editing ? 'Edit Gudang' : 'Tambah Gudang'}
        footer={
          <>
            <button className="btn-secondary" onClick={() => setFormOpen(false)}>
              Batal
            </button>
            <button className="btn-primary" onClick={handleSubmit} disabled={saving}>
              {saving ? 'Menyimpan...' : 'Simpan'}
            </button>
          </>
        }
      >
        <div className="form-group">
          <label>Nama Gudang</label>
          <input className="form-input" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
        </div>
        <div className="form-group">
          <label>Kode</label>
          <input className="form-input" value={form.code} onChange={(e) => setForm({ ...form, code: e.target.value })} />
        </div>
        <div className="form-group">
          <label>Alamat</label>
          <textarea
            className="form-textarea"
            rows={2}
            value={form.address}
            onChange={(e) => setForm({ ...form, address: e.target.value })}
          />
        </div>
      </Modal>

      <ConfirmDialog
        isOpen={!!confirmDelete}
        title="Hapus Gudang"
        message={`Yakin ingin menghapus gudang "${confirmDelete?.name}"?`}
        confirmLabel="Hapus"
        onConfirm={handleDelete}
        onCancel={() => setConfirmDelete(null)}
      />
    </div>
  );
};

export default GudangPage;
