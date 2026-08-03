import { useEffect, useState } from 'react';
import { api, ApiError } from '../../lib/api';
import type { MemberPermission, PermissionDef, UserListItem } from '../../types';
import { useSelectedProject } from '../../hooks/useSelectedProject';
import { useProjectAutoSelect } from '../../hooks/useProjectAutoSelect';
import PageHeader from '../../components/PageHeader';
import TableCard from '../../components/TableCard';
import '../Dashboard.css';

const PermissionPage = () => {
  const { selectedProjectId } = useSelectedProject();
  useProjectAutoSelect('/admin/projects', true);
  const [permissions, setPermissions] = useState<PermissionDef[]>([]);
  const [members, setMembers] = useState<UserListItem[]>([]);
  const [selectedUserRoleId, setSelectedUserRoleId] = useState<string>('');
  const [granted, setGranted] = useState<MemberPermission[]>([]);
  const [errorMsg, setErrorMsg] = useState('');
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    api
      .get<{ data: PermissionDef[] }>('/admin/permissions')
      .then((res) => setPermissions(res.data))
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat katalog permission'));
  }, []);

  const loadGranted = (userRoleId: string) => {
    if (!userRoleId || !selectedProjectId || selectedProjectId === 'all') {
      setGranted([]);
      return;
    }
    api
      .get<{ data: MemberPermission[] }>(`/admin/projects/${selectedProjectId}/user-roles/${userRoleId}/permissions`)
      .then((res) => setGranted(res.data))
      .catch(() => setGranted([]));
  };

  useEffect(() => {
    if (selectedProjectId === 'all' || !selectedProjectId) {
      setMembers([]);
      return;
    }
    api
      .get<{ data: UserListItem[] }>(`/admin/projects/${selectedProjectId}/members`)
      .then((res) => {
        const memberOnly = res.data.filter((m) => m.role === 'member');
        setMembers(memberOnly);
        if (memberOnly[0]) {
          setSelectedUserRoleId(String(memberOnly[0].user_role_id));
          loadGranted(String(memberOnly[0].user_role_id));
        } else {
          setSelectedUserRoleId('');
          setGranted([]);
        }
      })
      .catch(() => setMembers([]));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId]);

  const handleSelectUser = (userRoleId: string) => {
    setSelectedUserRoleId(userRoleId);
    loadGranted(userRoleId);
  };

  const isGranted = (permissionId: number) => granted.some((g) => g.permission_id === permissionId);

  const togglePermission = async (permission: PermissionDef) => {
    if (!selectedUserRoleId || !selectedProjectId) return;
    setSaving(true);
    try {
      if (isGranted(permission.id)) {
        await api.delete(`/admin/projects/${selectedProjectId}/user-roles/${selectedUserRoleId}/permissions/${permission.id}`);
      } else {
        await api.post(`/admin/projects/${selectedProjectId}/user-roles/${selectedUserRoleId}/permissions`, {
          permission_id: permission.id,
        });
      }
      loadGranted(selectedUserRoleId);
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal mengubah permission');
    } finally {
      setSaving(false);
    }
  };

  const groupedByModule = permissions.reduce<Record<string, PermissionDef[]>>((acc, p) => {
    acc[p.module] = acc[p.module] || [];
    acc[p.module].push(p);
    return acc;
  }, {});

  return (
    <div className="dashboard-content">
      <PageHeader title="Permission" subtitle="Atur hak akses member di project yang Anda kelola" />

      {errorMsg && (
        <div style={{ color: 'var(--danger)', marginBottom: '16px', fontSize: '14px' }}>{errorMsg}</div>
      )}

      <TableCard title="Katalog Permission" count={permissions.length}>
        <thead>
          <tr>
            <th>CODE</th>
            <th>NAMA</th>
            <th>MODULE</th>
          </tr>
        </thead>
        <tbody>
          {permissions.map((p) => (
            <tr key={p.id}>
              <td>{p.code}</td>
              <td>{p.name}</td>
              <td style={{ color: 'var(--text-muted)' }}>{p.module}</td>
            </tr>
          ))}
        </tbody>
      </TableCard>

      <div className="card" style={{ marginTop: 24 }}>
        <h3>Assign Permission ke Member</h3>
        {members.length === 0 ? (
          <p style={{ color: 'var(--text-muted)' }}>Belum ada member di project ini.</p>
        ) : (
          <>
            <div className="form-group" style={{ maxWidth: 360 }}>
              <label>Pilih Member</label>
              <select
                className="form-select"
                value={selectedUserRoleId}
                onChange={(e) => handleSelectUser(e.target.value)}
              >
                {members.map((m) => (
                  <option key={m.user_role_id} value={m.user_role_id}>
                    {m.name} ({m.email})
                  </option>
                ))}
              </select>
            </div>

            {Object.entries(groupedByModule).map(([module, perms]) => (
              <div key={module} style={{ marginTop: 16 }}>
                <div style={{ fontSize: 13, color: 'var(--text-muted)', marginBottom: 8, textTransform: 'uppercase' }}>
                  {module}
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
          </>
        )}
      </div>
    </div>
  );
};

export default PermissionPage;
