import { useEffect, useState } from 'react';
import { Plus, Tags, Trash2, Pencil } from 'lucide-react';
import { api, ApiError, buildQuery, type PaginatedResponse } from '../../lib/api';
import type { ProductCategory } from '../../types';
import { useSelectedProject } from '../../hooks/useSelectedProject';
import { useProjectAutoSelect } from '../../hooks/useProjectAutoSelect';
import { usePagination, metaFrom, PER_PAGE } from '../../hooks/usePagination';
import { useDebounced } from '../../hooks/useDebounced';
import PageHeader from '../../components/PageHeader';
import StatCard from '../../components/StatCard';
import TableCard from '../../components/TableCard';
import Pagination from '../../components/Pagination';
import Modal from '../../components/Modal';
import ConfirmDialog from '../../components/ConfirmDialog';
import ProjectPicker from '../../components/ProjectPicker';
import { makeCan } from '../../lib/permissions';
import '../Dashboard.css';

interface CategoryFormState {
  name: string;
  code: string;
}

const emptyForm: CategoryFormState = { name: '', code: '' };

interface KategoriPageProps {
  /** '/super-admin' (global, ?project_id) or '/admin/projects' (path-scoped). */
  apiBasePrefix?: string;
  scopeMode?: 'query' | 'path';
  projectEndpoint?: string;
  includeAllOption?: boolean;
  showProjectPicker?: boolean;
  /** Granted permission codes. Undefined = no per-action gating (Super Admin / Admin). */
  permissions?: string[];
}

const KategoriPage = ({
  apiBasePrefix = '/super-admin',
  scopeMode = 'query',
  projectEndpoint = '/super-admin/projects',
  includeAllOption = true,
  showProjectPicker = true,
  permissions,
}: KategoriPageProps) => {
  // Categories have no permissions of their own: they are part of the Product
  // catalogue and reuse product.* one for one, exactly as the member routes do.
  const can = makeCan(permissions);
  const canCreate = can('product.create');
  const canUpdate = can('product.update');
  const canDelete = can('product.delete');

  const { selectedProjectId, setSelectedProjectId } = useSelectedProject();
  useProjectAutoSelect(projectEndpoint, !showProjectPicker);
  const [search, setSearch] = useState('');
  // Searching in SQL, not on screen: filtering the page the server already
  // sliced would only ever look at the 25 rows currently rendered.
  const debouncedSearch = useDebounced(search);
  const { page, setPage, meta, setMeta } = usePagination(`${selectedProjectId}|${debouncedSearch}`);

  const [categories, setCategories] = useState<ProductCategory[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState('');

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<ProductCategory | null>(null);
  const [form, setForm] = useState<CategoryFormState>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState<ProductCategory | null>(null);

  const resourceBase = scopeMode === 'path' ? `${apiBasePrefix}/${selectedProjectId}` : apiBasePrefix;

  const loadCategories = () => {
    if (selectedProjectId === 'all') {
      setCategories([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    const query = buildQuery({
      ...(scopeMode === 'query' ? { project_id: selectedProjectId } : {}),
      q: debouncedSearch,
      page,
      per_page: PER_PAGE,
    });
    api
      .get<PaginatedResponse<ProductCategory>>(`${resourceBase}/product-categories${query}`)
      .then((res) => {
        setCategories(res.data);
        setMeta(metaFrom(res));
      })
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat data kategori'))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadCategories();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId, apiBasePrefix, scopeMode, page, debouncedSearch]);

  const openCreateForm = () => {
    setEditing(null);
    setForm(emptyForm);
    setErrorMsg('');
    setFormOpen(true);
  };

  const openEditForm = (cat: ProductCategory) => {
    setEditing(cat);
    setForm({ name: cat.name, code: cat.code });
    setErrorMsg('');
    setFormOpen(true);
  };

  const handleSubmit = async () => {
    if (!form.name.trim() || !form.code.trim()) return;
    setSaving(true);
    setErrorMsg('');
    try {
      const payload = { name: form.name.trim(), code: form.code.trim() };
      if (editing) {
        await api.put(`${resourceBase}/product-categories/${editing.id}`, payload);
      } else {
        await api.post(`${resourceBase}/product-categories`, {
          ...payload,
          project_id: Number(selectedProjectId),
        });
      }
      setFormOpen(false);
      loadCategories();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal menyimpan kategori');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!confirmDelete) return;
    try {
      await api.delete(`${resourceBase}/product-categories/${confirmDelete.id}`);
      setConfirmDelete(null);
      loadCategories();
    } catch (err) {
      // The 409 from a category still in use is the useful case here, and its
      // message already says what to do about it.
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal menghapus kategori');
      setConfirmDelete(null);
    }
  };

  return (
    <div className="dashboard-content">
      <PageHeader
        title="Kategori"
        subtitle="Kelompok produk, dipakai juga sebagai filter di API eksternal"
        searchValue={search}
        onSearchChange={setSearch}
        searchPlaceholder="Cari nama atau kode..."
        actions={
          canCreate && (
            <button className="btn-primary" onClick={openCreateForm} disabled={selectedProjectId === 'all'}>
              <Plus size={18} />
              Tambah Kategori
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
        <p style={{ color: 'var(--text-muted)' }}>Pilih project terlebih dahulu untuk melihat daftar kategori.</p>
      ) : (
        <>
          <div className="summary-cards" style={{ gridTemplateColumns: 'repeat(1, 1fr)' }}>
            <StatCard label="Total Kategori" value={meta.total} icon={<Tags size={18} />} />
          </div>

          <TableCard title="Daftar Kategori" count={meta.total}>
            <thead>
              <tr>
                <th>KODE</th>
                <th>NAMA KATEGORI</th>
                <th>DIBUAT</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {!loading && categories.length === 0 && (
                <tr>
                  <td colSpan={4} style={{ color: 'var(--text-muted)' }}>
                    {debouncedSearch ? 'Tidak ada kategori yang cocok.' : 'Belum ada kategori di project ini.'}
                  </td>
                </tr>
              )}
              {!loading &&
                categories.map((cat) => (
                  <tr key={cat.id}>
                    <td style={{ fontFamily: 'monospace', fontSize: 12 }}>{cat.code}</td>
                    <td>{cat.name}</td>
                    <td style={{ color: 'var(--text-muted)' }}>
                      {new Date(cat.created_at).toLocaleDateString('id-ID')}
                    </td>
                    <td style={{ textAlign: 'right', display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
                      {canUpdate && (
                        <button
                          style={{ background: 'transparent', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}
                          onClick={() => openEditForm(cat)}
                          title="Ubah Kategori"
                        >
                          <Pencil size={16} />
                        </button>
                      )}
                      {canDelete && (
                        <button
                          style={{ background: 'transparent', border: 'none', color: 'var(--danger)', cursor: 'pointer' }}
                          onClick={() => setConfirmDelete(cat)}
                          title="Hapus Kategori"
                        >
                          <Trash2 size={16} />
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
            </tbody>
          </TableCard>

          <Pagination page={page} meta={meta} onChange={setPage} label="kategori" />
        </>
      )}

      <Modal
        isOpen={formOpen}
        onClose={() => setFormOpen(false)}
        title={editing ? 'Edit Kategori' : 'Tambah Kategori'}
        footer={
          <>
            <button className="btn-secondary" onClick={() => setFormOpen(false)}>
              Batal
            </button>
            <button
              className="btn-primary"
              onClick={handleSubmit}
              disabled={saving || !form.name.trim() || !form.code.trim()}
            >
              {saving ? 'Menyimpan...' : 'Simpan'}
            </button>
          </>
        }
      >
        <div className="form-group">
          <label>Nama Kategori</label>
          <input
            className="form-input"
            placeholder="mis. Minuman, Merchandise"
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
          />
        </div>
        <div className="form-group">
          <label>Kode</label>
          <input
            className="form-input"
            placeholder="mis. MIN"
            maxLength={32}
            value={form.code}
            onChange={(e) => setForm({ ...form, code: e.target.value })}
          />
          <p style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 6 }}>
            Dipakai pihak ketiga sebagai <code>?category={form.code.trim().toUpperCase() || 'MIN'}</code> di API
            eksternal. Disimpan dalam huruf besar, dan harus unik di project ini. Kode tetap sama walau nama
            kategorinya diubah, jadi tautan di sisi mereka tidak ikut mati.
          </p>
        </div>
      </Modal>

      <ConfirmDialog
        isOpen={!!confirmDelete}
        title="Hapus Kategori"
        message={`Yakin ingin menghapus kategori "${confirmDelete?.name}"?`}
        confirmLabel="Hapus"
        onConfirm={handleDelete}
        onCancel={() => setConfirmDelete(null)}
      />
    </div>
  );
};

export default KategoriPage;
