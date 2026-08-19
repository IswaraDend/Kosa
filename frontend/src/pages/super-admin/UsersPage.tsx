import { useEffect, useState } from 'react';
import { usePagination, metaFrom, PER_PAGE } from '../../hooks/usePagination';
import { useDebounced } from '../../hooks/useDebounced';
import { Plus, Trash2 } from 'lucide-react';
import { api, ApiError, buildQuery, type PaginatedResponse } from '../../lib/api';
import type { Project, UserListItem } from '../../types';
import PageHeader from '../../components/PageHeader';
import TableCard from '../../components/TableCard';
import Pagination from '../../components/Pagination';
import Badge from '../../components/Badge';
import Modal from '../../components/Modal';
import ConfirmDialog from '../../components/ConfirmDialog';
import '../Dashboard.css';

interface UserFormState {
  name: string;
  email: string;
  password: string;
  project_id: string;
  role_name: 'admin' | 'member';
}

const emptyForm: UserFormState = { name: '', email: '', password: '', project_id: '', role_name: 'member' };

const UsersPage = () => {
  const [users, setUsers] = useState<UserListItem[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState('');
  const [search, setSearch] = useState('');
  const [roleFilter, setRoleFilter] = useState<'all' | 'admin' | 'member'>('all');
  // Search and the role filter now run in SQL: filtering the page the server
  // already sliced would only ever search the 25 rows on screen.
  const debouncedSearch = useDebounced(search);
  const { page, setPage, meta, setMeta } = usePagination(`${debouncedSearch}|${roleFilter}`);

  const [formOpen, setFormOpen] = useState(false);
  const [form, setForm] = useState<UserFormState>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState<UserListItem | null>(null);

  const loadUsers = () => {
    setLoading(true);
    const query = buildQuery({
      q: debouncedSearch,
      role: roleFilter === 'all' ? undefined : roleFilter,
      page,
      per_page: PER_PAGE,
    });
    api
      .get<PaginatedResponse<UserListItem>>(`/super-admin/users${query}`)
      .then((res) => {
        setUsers(res.data);
        setMeta(metaFrom(res));
      })
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat data user'))
      .finally(() => setLoading(false));
  };

  // The project list feeds the "create user" form, not the table — fetched
  // once rather than on every page change.
  useEffect(() => {
    api
      .get<{ data: Project[] }>('/super-admin/projects')
      .then((res) => setProjects(res.data))
      .catch(() => setProjects([]));
  }, []);

  useEffect(() => {
    loadUsers();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedSearch, roleFilter, page]);

  const openCreateForm = () => {
    setForm({ ...emptyForm, project_id: projects[0] ? String(projects[0].id) : '' });
    setFormOpen(true);
  };

  const handleSubmit = async () => {
    setSaving(true);
    setErrorMsg('');
    try {
      await api.post('/super-admin/users', {
        name: form.name,
        email: form.email,
        password: form.password,
        project_id: Number(form.project_id),
        role_name: form.role_name,
      });
      setFormOpen(false);
      loadUsers();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal membuat user');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!confirmDelete) return;
    try {
      await api.delete(`/super-admin/users/${confirmDelete.id}`);
      setConfirmDelete(null);
      loadUsers();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal menghapus user');
      setConfirmDelete(null);
    }
  };

  return (
    <div className="dashboard-content">
      <PageHeader
        title="Admin & Member"
        subtitle="Kelola akun admin dan member di setiap project"
        searchValue={search}
        onSearchChange={setSearch}
        searchPlaceholder="Cari nama atau email..."
        actions={
          <button className="btn-primary" onClick={openCreateForm}>
            <Plus size={18} />
            Tambah User
          </button>
        }
      />

      {errorMsg && (
        <div style={{ color: 'var(--danger)', marginBottom: '16px', fontSize: '14px' }}>{errorMsg}</div>
      )}

      <div className="filter-group">
        {(['all', 'admin', 'member'] as const).map((r) => (
          <button
            key={r}
            className={`filter-btn ${roleFilter === r ? 'active' : ''}`}
            onClick={() => setRoleFilter(r)}
          >
            {r === 'all' ? 'Semua' : r === 'admin' ? 'Admin' : 'Member'}
          </button>
        ))}
      </div>

      <TableCard title="Daftar User" count={meta.total}>
        <thead>
          <tr>
            <th>NAMA</th>
            <th>EMAIL</th>
            <th>ROLE</th>
            <th>PROJECT</th>
            <th>DIBUAT</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {!loading &&
            users.map((u) => (
              <tr key={u.user_role_id}>
                <td>{u.name}</td>
                <td style={{ color: 'var(--text-muted)' }}>{u.email}</td>
                <td>
                  <Badge label={u.role === 'admin' ? 'Admin' : 'Member'} variant={u.role} />
                </td>
                <td>{u.project_name}</td>
                <td style={{ color: 'var(--text-muted)' }}>
                  {new Date(u.created_at).toLocaleDateString('id-ID')}
                </td>
                <td style={{ textAlign: 'right' }}>
                  <button
                    style={{ background: 'transparent', border: 'none', color: 'var(--danger)', cursor: 'pointer' }}
                    onClick={() => setConfirmDelete(u)}
                  >
                    <Trash2 size={16} />
                  </button>
                </td>
              </tr>
            ))}
        </tbody>
      </TableCard>

      <Pagination page={page} meta={meta} onChange={setPage} label="user" />

      <Modal
        isOpen={formOpen}
        onClose={() => setFormOpen(false)}
        title="Tambah User"
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
        <div className="form-group">
          <label>Project</label>
          <select
            className="form-select"
            value={form.project_id}
            onChange={(e) => setForm({ ...form, project_id: e.target.value })}
          >
            {projects.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
        </div>
        <div className="form-group">
          <label>Role</label>
          <select
            className="form-select"
            value={form.role_name}
            onChange={(e) => setForm({ ...form, role_name: e.target.value as 'admin' | 'member' })}
          >
            <option value="admin">Admin</option>
            <option value="member">Member</option>
          </select>
        </div>
      </Modal>

      <ConfirmDialog
        isOpen={!!confirmDelete}
        title="Hapus User"
        message={`Yakin ingin menghapus user "${confirmDelete?.name}"?`}
        confirmLabel="Hapus"
        onConfirm={handleDelete}
        onCancel={() => setConfirmDelete(null)}
      />
    </div>
  );
};

export default UsersPage;
