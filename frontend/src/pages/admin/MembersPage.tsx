import { useEffect, useState } from 'react';
import { usePagination, metaFrom, PER_PAGE } from '../../hooks/usePagination';
import { useDebounced } from '../../hooks/useDebounced';
import { Plus, Trash2 } from 'lucide-react';
import { api, ApiError, buildQuery, type PaginatedResponse } from '../../lib/api';
import type { UserListItem } from '../../types';
import { useSelectedProject } from '../../hooks/useSelectedProject';
import { useProjectAutoSelect } from '../../hooks/useProjectAutoSelect';
import PageHeader from '../../components/PageHeader';
import TableCard from '../../components/TableCard';
import Pagination from '../../components/Pagination';
import Badge from '../../components/Badge';
import Modal from '../../components/Modal';
import ConfirmDialog from '../../components/ConfirmDialog';
import '../Dashboard.css';

interface MemberFormState {
  name: string;
  email: string;
  password: string;
}

const emptyForm: MemberFormState = { name: '', email: '', password: '' };

const MembersPage = () => {
  const { selectedProjectId } = useSelectedProject();
  useProjectAutoSelect('/admin/projects', true);
  const [members, setMembers] = useState<UserListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState('');
  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounced(search);
  const { page, setPage, meta, setMeta } = usePagination(`${selectedProjectId}|${debouncedSearch}`);

  const [formOpen, setFormOpen] = useState(false);
  const [form, setForm] = useState<MemberFormState>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState<UserListItem | null>(null);

  const loadMembers = () => {
    if (selectedProjectId === 'all' || !selectedProjectId) {
      setMembers([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    api
      .get<PaginatedResponse<UserListItem>>(
        `/admin/projects/${selectedProjectId}/members${buildQuery({ q: debouncedSearch, page, per_page: PER_PAGE })}`,
      )
      .then((res) => {
        setMembers(res.data);
        setMeta(metaFrom(res));
      })
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat data member'))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadMembers();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId, debouncedSearch, page]);



  const openCreateForm = () => {
    setForm(emptyForm);
    setFormOpen(true);
  };

  const handleSubmit = async () => {
    setSaving(true);
    setErrorMsg('');
    try {
      await api.post(`/admin/projects/${selectedProjectId}/members`, form);
      setFormOpen(false);
      loadMembers();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal menambahkan member');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!confirmDelete) return;
    try {
      await api.delete(`/admin/projects/${selectedProjectId}/members/${confirmDelete.user_role_id}`);
      setConfirmDelete(null);
      loadMembers();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal mencabut member');
      setConfirmDelete(null);
    }
  };

  return (
    <div className="dashboard-content">
      <PageHeader
        title="Members"
        subtitle="Kelola member di project yang Anda kelola"
        searchValue={search}
        onSearchChange={setSearch}
        searchPlaceholder="Cari nama atau email..."
        actions={
          <button className="btn-primary" onClick={openCreateForm} disabled={!selectedProjectId || selectedProjectId === 'all'}>
            <Plus size={18} />
            Tambah Member
          </button>
        }
      />

      {errorMsg && (
        <div style={{ color: 'var(--danger)', marginBottom: '16px', fontSize: '14px' }}>{errorMsg}</div>
      )}

      <TableCard title="Daftar Member" count={meta.total}>
        <thead>
          <tr>
            <th>NAMA</th>
            <th>EMAIL</th>
            <th>ROLE</th>
            <th>DIBUAT</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {!loading &&
            members.map((m) => (
              <tr key={m.user_role_id}>
                <td>{m.name}</td>
                <td style={{ color: 'var(--text-muted)' }}>{m.email}</td>
                <td>
                  <Badge label={m.role === 'admin' ? 'Admin' : 'Member'} variant={m.role} />
                </td>
                <td style={{ color: 'var(--text-muted)' }}>
                  {new Date(m.created_at).toLocaleDateString('id-ID')}
                </td>
                <td style={{ textAlign: 'right' }}>
                  {m.role === 'member' && (
                    <button
                      style={{ background: 'transparent', border: 'none', color: 'var(--danger)', cursor: 'pointer' }}
                      onClick={() => setConfirmDelete(m)}
                    >
                      <Trash2 size={16} />
                    </button>
                  )}
                </td>
              </tr>
            ))}
        </tbody>
      </TableCard>

      <Pagination page={page} meta={meta} onChange={setPage} label="member" />

      <Modal
        isOpen={formOpen}
        onClose={() => setFormOpen(false)}
        title="Tambah Member"
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
          <label>Nama</label>
          <input className="form-input" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
        </div>
        <div className="form-group">
          <label>Email</label>
          <input
            type="email"
            className="form-input"
            value={form.email}
            onChange={(e) => setForm({ ...form, email: e.target.value })}
          />
        </div>
        <div className="form-group">
          <label>Password</label>
          <input
            type="password"
            className="form-input"
            value={form.password}
            onChange={(e) => setForm({ ...form, password: e.target.value })}
          />
        </div>
      </Modal>

      <ConfirmDialog
        isOpen={!!confirmDelete}
        title="Cabut Member"
        message={`Yakin ingin mencabut akses "${confirmDelete?.name}" dari project ini?`}
        confirmLabel="Cabut"
        onConfirm={handleDelete}
        onCancel={() => setConfirmDelete(null)}
      />
    </div>
  );
};

export default MembersPage;
