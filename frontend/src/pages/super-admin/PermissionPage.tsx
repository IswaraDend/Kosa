import { useEffect, useState } from 'react';
import { Plus } from 'lucide-react';
import { api, ApiError, buildQuery } from '../../lib/api';
import type { MemberPermission, PermissionDef, UserListItem } from '../../types';
import PageHeader from '../../components/PageHeader';
import TableCard from '../../components/TableCard';
import Modal from '../../components/Modal';
import { MODULES, moduleLabel } from '../../lib/modules';
import '../Dashboard.css';

const PermissionPage = () => {
  const [permissions, setPermissions] = useState<PermissionDef[]>([]);
  // Separate from `permissions`: the catalogue table shows everything (Super
  // Admin is not module-gated), but only these can actually be granted to the
  // selected user — see the effect below.
  const [assignable, setAssignable] = useState<PermissionDef[]>([]);
  const [users, setUsers] = useState<UserListItem[]>([]);
  const [selectedUserRoleId, setSelectedUserRoleId] = useState<string>('');
  const [granted, setGranted] = useState<MemberPermission[]>([]);
  const [errorMsg, setErrorMsg] = useState('');
  const [saving, setSaving] = useState(false);

  const [formOpen, setFormOpen] = useState(false);
  const [newPermission, setNewPermission] = useState({ code: '', name: '', module: MODULES[0].code as string });

  const loadPermissions = () => {
    api
      .get<{ data: PermissionDef[] }>('/super-admin/permissions')
      .then((res) => setPermissions(res.data))
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat katalog permission'));
  };

  const loadGranted = (userRoleId: string) => {
    if (!userRoleId) {
      setGranted([]);
      return;
    }
    api
      .get<{ data: MemberPermission[] }>(`/super-admin/user-roles/${userRoleId}/permissions`)
      .then((res) => setGranted(res.data))
      .catch(() => setGranted([]));
  };

  useEffect(() => {
    loadPermissions();
    api
      .get<{ data: UserListItem[] }>('/super-admin/users')
      .then((res) => {
        // Permission is a member-only concept: middleware.RequirePermission
        // looks up roles.name = "member" literally, so a grant attached to an
        // admin user-role is never read by anything. Listing admins here would
        // offer checkboxes that save happily and then do nothing at all.
        const memberOnly = res.data.filter((u) => u.role === 'member');
        setUsers(memberOnly);
        if (memberOnly[0]) {
          setSelectedUserRoleId(String(memberOnly[0].user_role_id));
          loadGranted(String(memberOnly[0].user_role_id));
        }
      })
      .catch(() => setUsers([]));
  }, []);

  const handleSelectUser = (userRoleId: string) => {
    setSelectedUserRoleId(userRoleId);
    loadGranted(userRoleId);
  };

  const selectedUser = users.find((u) => String(u.user_role_id) === selectedUserRoleId);

  // What a member can actually be granted is bounded by their project's
  // enabled modules, even for a Super Admin: granting anything else creates a
  // row that RequireModule rejects at request time anyway.
  useEffect(() => {
    if (!selectedUser) {
      setAssignable([]);
      return;
    }
    api
      .get<{ data: PermissionDef[] }>(`/super-admin/permissions${buildQuery({ project_id: selectedUser.project_id })}`)
      .then((res) => setAssignable(res.data))
      .catch(() => setAssignable([]));
  }, [selectedUser]);

  const isGranted = (permissionId: number) => granted.some((g) => g.permission_id === permissionId);

  const togglePermission = async (permission: PermissionDef) => {
    if (!selectedUserRoleId) return;
    setSaving(true);
    try {
      if (isGranted(permission.id)) {
        await api.delete(`/super-admin/user-roles/${selectedUserRoleId}/permissions/${permission.id}`);
      } else {
        await api.post(`/super-admin/user-roles/${selectedUserRoleId}/permissions`, { permission_id: permission.id });
      }
      loadGranted(selectedUserRoleId);
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal mengubah permission');
    } finally {
      setSaving(false);
    }
  };

  const handleCreatePermission = async () => {
    setSaving(true);
    setErrorMsg('');
    try {
      await api.post('/super-admin/permissions', newPermission);
      setFormOpen(false);
      setNewPermission({ code: '', name: '', module: MODULES[0].code as string });
      loadPermissions();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal membuat permission');
    } finally {
      setSaving(false);
    }
  };

  const groupedByModule = assignable.reduce<Record<string, PermissionDef[]>>((acc, p) => {
    acc[p.module] = acc[p.module] || [];
    acc[p.module].push(p);
    return acc;
  }, {});

  return (
    <div className="dashboard-content">
      <PageHeader
        title="Permission"
        subtitle="Kelola katalog permission dan hak akses per user"
        actions={
          <button className="btn-primary" onClick={() => setFormOpen(true)}>
            <Plus size={18} />
            Tambah Permission
          </button>
        }
      />

      {errorMsg && (
        <div style={{ color: 'var(--danger)', marginBottom: '16px', fontSize: '14px' }}>{errorMsg}</div>
      )}

      <TableCard title="Katalog Permission" count={permissions.length}>
        <thead>
          <tr>
            <th>CODE</th>
            <th>NAMA</th>
            <th>FITUR</th>
          </tr>
        </thead>
        <tbody>
          {permissions.map((p) => (
            <tr key={p.id}>
              <td>{p.code}</td>
              <td>{p.name}</td>
              <td style={{ color: 'var(--text-muted)' }}>{moduleLabel(p.module)}</td>
            </tr>
          ))}
        </tbody>
      </TableCard>

      <div className="card" style={{ marginTop: 24 }}>
        <h3>Assign Permission ke Member</h3>
        <p style={{ color: 'var(--text-muted)', fontSize: 13, marginTop: 0 }}>
          Hanya member yang tunduk pada permission. Admin memperoleh akses dari perannya di project, dibatasi oleh
          fitur yang aktif — bukan dari daftar ini.
        </p>

        {users.length === 0 ? (
          <p style={{ color: 'var(--text-muted)', margin: 0 }}>Belum ada member di project mana pun.</p>
        ) : (
          <div className="form-group" style={{ maxWidth: 360 }}>
            <label>Pilih Member</label>
            <select
              className="form-select"
              value={selectedUserRoleId}
              onChange={(e) => handleSelectUser(e.target.value)}
            >
              {users.map((u) => (
                <option key={u.user_role_id} value={u.user_role_id}>
                  {u.name} @ {u.project_name}
                </option>
              ))}
            </select>
          </div>
        )}

        {selectedUser && assignable.length === 0 && (
          <p style={{ color: 'var(--text-muted)', fontSize: 13 }}>
            Project "{selectedUser.project_name}" belum mengaktifkan fitur apa pun, jadi belum ada permission yang
            bisa diberikan.
          </p>
        )}

        {Object.entries(groupedByModule).map(([module, perms]) => (
          <div key={module} style={{ marginTop: 16 }}>
            <div style={{ fontSize: 13, color: 'var(--text-muted)', marginBottom: 8, textTransform: 'uppercase' }}>
              {moduleLabel(module)}
            </div>
            <div className="permission-grid">
              {perms.map((p) => (
                <label className="permission-item" key={p.id}>
                  <input
                    type="checkbox"
                    checked={isGranted(p.id)}
                    disabled={saving || !selectedUserRoleId}
                    onChange={() => togglePermission(p)}
                  />
                  {p.name}
                </label>
              ))}
            </div>
          </div>
        ))}
      </div>

      <Modal
        isOpen={formOpen}
        onClose={() => setFormOpen(false)}
        title="Tambah Permission"
        footer={
          <>
            <button className="btn-secondary" onClick={() => setFormOpen(false)}>
              Batal
            </button>
            <button className="btn-primary" onClick={handleCreatePermission} disabled={saving}>
              {saving ? 'Menyimpan...' : 'Simpan'}
            </button>
          </>
        }
      >
        <div className="form-group">
          <label>Code</label>
          <input
            className="form-input"
            value={newPermission.code}
            onChange={(e) => setNewPermission({ ...newPermission, code: e.target.value })}
            placeholder="e.g. product.view"
          />
        </div>
        <div className="form-group">
          <label>Nama</label>
          <input
            className="form-input"
            value={newPermission.name}
            onChange={(e) => setNewPermission({ ...newPermission, name: e.target.value })}
          />
        </div>
        <div className="form-group">
          <label>Fitur</label>
          <select
            className="form-select"
            value={newPermission.module}
            onChange={(e) => setNewPermission({ ...newPermission, module: e.target.value })}
          >
            {MODULES.map((m) => (
              <option key={m.code} value={m.code}>
                {m.label}
              </option>
            ))}
          </select>
          <p style={{ color: 'var(--text-muted)', fontSize: 12, margin: '6px 0 0' }}>
            Permission hanya muncul untuk project yang mengaktifkan fitur ini.
          </p>
        </div>
      </Modal>
    </div>
  );
};

export default PermissionPage;
