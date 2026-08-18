import { useEffect, useState } from 'react';
import { Plus, Folder, CheckCircle, Users, UserCog, MoreVertical } from 'lucide-react';
import { api, ApiError } from '../../lib/api';
import type { Project } from '../../types';
import { MODULES, LAYERS, expandModules } from '../../lib/modules';
import type { CSSProperties } from 'react';
import PageHeader from '../../components/PageHeader';
import StatCard from '../../components/StatCard';
import TableCard from '../../components/TableCard';
import Badge from '../../components/Badge';
import Modal from '../../components/Modal';
import ConfirmDialog from '../../components/ConfirmDialog';
import '../Dashboard.css';

interface ProjectFormState {
  name: string;
  code: string;
  description: string;
  modules: string[];
}

const emptyForm: ProjectFormState = { name: '', code: '', description: '', modules: [] };

const KelolaProjectPage = () => {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState('');
  const [search, setSearch] = useState('');

  const [formOpen, setFormOpen] = useState(false);
  const [editingProject, setEditingProject] = useState<Project | null>(null);
  const [form, setForm] = useState<ProjectFormState>(emptyForm);
  const [saving, setSaving] = useState(false);

  const [menuOpenId, setMenuOpenId] = useState<number | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<Project | null>(null);

  const loadProjects = () => {
    setLoading(true);
    api
      .get<{ data: Project[] }>('/super-admin/projects')
      .then((res) => setProjects(res.data))
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat data project'))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadProjects();
  }, []);

  const filtered = projects.filter(
    (p) =>
      p.name.toLowerCase().includes(search.toLowerCase()) ||
      p.code.toLowerCase().includes(search.toLowerCase()),
  );

  const openCreateForm = () => {
    setEditingProject(null);
    setForm(emptyForm);
    setFormOpen(true);
  };

  const openEditForm = (project: Project) => {
    setEditingProject(project);
    setForm({
      name: project.name,
      code: project.code,
      description: project.description,
      modules: project.modules ?? [],
    });
    setFormOpen(true);
    setMenuOpenId(null);
  };

  const toggleModule = (code: string) => {
    const isEnabled = form.modules.includes(code);
    const next = isEnabled ? form.modules.filter((m) => m !== code) : [...form.modules, code];
    setForm({ ...form, modules: expandModules(next) });
  };

  const handleSubmit = async () => {
    setSaving(true);
    setErrorMsg('');
    try {
      if (editingProject) {
        await api.put(`/super-admin/projects/${editingProject.id}`, form);
      } else {
        await api.post('/super-admin/projects', form);
      }
      setFormOpen(false);
      loadProjects();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal menyimpan project');
    } finally {
      setSaving(false);
    }
  };

  const handleToggleStatus = async (project: Project) => {
    const nextStatus = project.status === 'aktif' ? 'nonaktif' : 'aktif';
    setMenuOpenId(null);
    try {
      await api.patch(`/super-admin/projects/${project.id}/status`, { status: nextStatus });
      loadProjects();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal mengubah status project');
    }
  };

  const handleDelete = async () => {
    if (!confirmDelete) return;
    try {
      await api.delete(`/super-admin/projects/${confirmDelete.id}`);
      setConfirmDelete(null);
      loadProjects();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal menghapus project');
      setConfirmDelete(null);
    }
  };

  const totalActive = projects.filter((p) => p.status === 'aktif').length;
  const distinctAdmins = new Set(projects.map((p) => p.admin_name).filter(Boolean)).size;
  const totalMembers = projects.reduce((sum, p) => sum + p.member_count, 0);

  return (
    <div className="dashboard-content">
      <PageHeader
        title="Kelola Project"
        subtitle="Buat project baru dan tentukan admin penanggung jawabnya"
        searchValue={search}
        onSearchChange={setSearch}
        searchPlaceholder="Cari project..."
        actions={
          <button className="btn-primary" onClick={openCreateForm}>
            <Plus size={18} />
            Tambah Project
          </button>
        }
      />

      {errorMsg && (
        <div style={{ color: 'var(--danger)', marginBottom: '16px', fontSize: '14px' }}>{errorMsg}</div>
      )}

      <div className="summary-cards">
        <StatCard label="Total Project" value={projects.length} icon={<Folder size={18} />} />
        <StatCard label="Project Aktif" value={totalActive} icon={<CheckCircle size={18} />} />
        <StatCard label="Admin Bertugas" value={distinctAdmins} icon={<UserCog size={18} />} />
        <StatCard label="Total Member" value={totalMembers} icon={<Users size={18} />} />
      </div>

      <TableCard title="Daftar Project" count={filtered.length}>
        <thead>
          <tr>
            <th>NAMA PROJECT</th>
            <th>KODE</th>
            <th>ADMIN</th>
            <th>MEMBER</th>
            <th>GUDANG</th>
            <th>STATUS</th>
            <th>DIBUAT</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {!loading &&
            filtered.map((project) => (
              <tr key={project.id}>
                <td>{project.name}</td>
                <td style={{ color: 'var(--text-muted)' }}>{project.code}</td>
                <td>{project.admin_name || '— Belum ditentukan'}</td>
                <td>{project.member_count}</td>
                <td>{project.warehouse_count}</td>
                <td>
                  <Badge label={project.status === 'aktif' ? 'Aktif' : 'Nonaktif'} variant={project.status} />
                </td>
                <td style={{ color: 'var(--text-muted)' }}>
                  {new Date(project.created_at).toLocaleDateString('id-ID')}
                </td>
                <td style={{ textAlign: 'right', position: 'relative' }}>
                  <button
                    style={{ background: 'transparent', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}
                    onClick={() => setMenuOpenId(menuOpenId === project.id ? null : project.id)}
                  >
                    <MoreVertical size={16} />
                  </button>
                  {menuOpenId === project.id && (
                    <div
                      style={{
                        position: 'absolute',
                        right: 24,
                        top: '100%',
                        background: 'var(--surface-color-light)',
                        border: '1px solid var(--border-color)',
                        borderRadius: 8,
                        overflow: 'hidden',
                        zIndex: 10,
                        minWidth: 140,
                      }}
                    >
                      <button className="dropdown-item" onClick={() => openEditForm(project)}>
                        Edit
                      </button>
                      <button className="dropdown-item" onClick={() => handleToggleStatus(project)}>
                        {project.status === 'aktif' ? 'Nonaktifkan' : 'Aktifkan'}
                      </button>
                      <button
                        className="dropdown-item"
                        style={{ color: 'var(--danger)' }}
                        onClick={() => {
                          setConfirmDelete(project);
                          setMenuOpenId(null);
                        }}
                      >
                        Hapus
                      </button>
                    </div>
                  )}
                </td>
              </tr>
            ))}
        </tbody>
      </TableCard>

      <Modal
        isOpen={formOpen}
        onClose={() => setFormOpen(false)}
        title={editingProject ? 'Edit Project' : 'Tambah Project'}
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
          <label>Nama Project</label>
          <input
            className="form-input"
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
          />
        </div>
        <div className="form-group">
          <label>Kode Project</label>
          <input
            className="form-input"
            value={form.code}
            onChange={(e) => setForm({ ...form, code: e.target.value })}
          />
        </div>
        <div className="form-group">
          <label>Deskripsi</label>
          <textarea
            className="form-textarea"
            rows={3}
            value={form.description}
            onChange={(e) => setForm({ ...form, description: e.target.value })}
          />
        </div>
        <div className="form-group">
          <label>Fitur Aktif</label>
          <p style={{ color: 'var(--text-muted)', fontSize: '13px', margin: '0 0 8px' }}>
            Pilih fitur yang bisa dipakai Admin &amp; Member project ini. Fitur yang butuh fitur lain (mis. Invoice
            butuh Produk &amp; Pelanggan) akan ikut tercentang otomatis.
          </p>
          <div className="module-picker">
            {LAYERS.map((layer) => (
              <div className="module-band" key={layer.code} style={{ '--band-c': `var(${layer.colorVar})` } as CSSProperties}>
                <div className="module-band-head">
                  {layer.name}
                  <span className="deva">{layer.deva}</span>
                </div>
                <div className="module-band-body">
                  {layer.modules.map((code) => {
                    const m = MODULES.find((mod) => mod.code === code)!;
                    const on = form.modules.includes(code);
                    return (
                      <label key={code} className={`module-check ${on ? 'on' : ''}`}>
                        <input type="checkbox" checked={on} onChange={() => toggleModule(code)} />
                        {m.label}
                      </label>
                    );
                  })}
                </div>
              </div>
            ))}
          </div>
        </div>
      </Modal>

      <ConfirmDialog
        isOpen={!!confirmDelete}
        title="Hapus Project"
        message={`Yakin ingin menghapus project "${confirmDelete?.name}"? Tindakan ini tidak bisa dibatalkan.`}
        confirmLabel="Hapus"
        onConfirm={handleDelete}
        onCancel={() => setConfirmDelete(null)}
      />
    </div>
  );
};

export default KelolaProjectPage;
