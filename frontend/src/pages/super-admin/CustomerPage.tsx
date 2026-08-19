import { useEffect, useState } from 'react';
import { Plus, Users, Trash2, Pencil } from 'lucide-react';
import { api, ApiError, buildQuery , type PaginatedResponse } from '../../lib/api';
import type { Customer } from '../../types';
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

interface CustomerFormState {
  name: string;
  phone: string;
  email: string;
  address: string;
}

const emptyForm: CustomerFormState = { name: '', phone: '', email: '', address: '' };

interface CustomerPageProps {
  apiBasePrefix?: string;
  scopeMode?: 'query' | 'path';
  projectEndpoint?: string;
  includeAllOption?: boolean;
  showProjectPicker?: boolean;
  /** Granted permission codes. Undefined = no per-action gating (Super Admin / Admin). */
  permissions?: string[];
}

const CustomerPage = ({
  apiBasePrefix = '/super-admin',
  scopeMode = 'query',
  projectEndpoint = '/super-admin/projects',
  includeAllOption = true,
  showProjectPicker = true,
  permissions,
}: CustomerPageProps) => {
  const can = makeCan(permissions);
  const { selectedProjectId, setSelectedProjectId } = useSelectedProject();
  useProjectAutoSelect(projectEndpoint, !showProjectPicker);
  const { page, setPage, meta, setMeta } = usePagination(selectedProjectId);
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState('');

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Customer | null>(null);
  const [form, setForm] = useState<CustomerFormState>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState<Customer | null>(null);

  const resourceBase = scopeMode === 'path' ? `${apiBasePrefix}/${selectedProjectId}` : apiBasePrefix;
  // Separate from listQuery on purpose: listQuery still scopes the auxiliary
  // dropdown fetches (warehouses, items, customers) which must stay complete,
  // while only the main table asks for a page.
  const pagedQuery = buildQuery({
    ...(scopeMode === 'query' ? { project_id: selectedProjectId } : {}),
    page,
    per_page: PER_PAGE,
  });


  const loadCustomers = () => {
    if (selectedProjectId === 'all') {
      setCustomers([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    api
      .get<PaginatedResponse<Customer>>(`${resourceBase}/customers${pagedQuery}`)
      .then((res) => {
        setCustomers(res.data);
        setMeta(metaFrom(res));
      })
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat data customer'))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadCustomers();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId, apiBasePrefix, scopeMode, page]);

  const openCreateForm = () => {
    setEditing(null);
    setForm(emptyForm);
    setFormOpen(true);
  };

  const openEditForm = (customer: Customer) => {
    setEditing(customer);
    setForm({ name: customer.name, phone: customer.phone, email: customer.email, address: customer.address });
    setFormOpen(true);
  };

  const handleSubmit = async () => {
    setSaving(true);
    setErrorMsg('');
    try {
      if (editing) {
        await api.put(`${resourceBase}/customers/${editing.id}`, form);
      } else {
        await api.post(`${resourceBase}/customers`, { ...form, project_id: Number(selectedProjectId) });
      }
      setFormOpen(false);
      loadCustomers();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal menyimpan customer');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!confirmDelete) return;
    try {
      await api.delete(`${resourceBase}/customers/${confirmDelete.id}`);
      setConfirmDelete(null);
      loadCustomers();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal menghapus customer');
      setConfirmDelete(null);
    }
  };

  return (
    <div className="dashboard-content">
      <PageHeader
        title="Pelanggan"
        subtitle="Kelola data customer untuk keperluan invoice"
        actions={
          can('customer.create') && (
            <button className="btn-primary" onClick={openCreateForm} disabled={selectedProjectId === 'all'}>
              <Plus size={18} />
              Tambah Customer
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
        <p style={{ color: 'var(--text-muted)' }}>Pilih project terlebih dahulu untuk melihat daftar customer.</p>
      ) : (
        <>
          <div className="summary-cards" style={{ gridTemplateColumns: 'repeat(1, 1fr)' }}>
            <StatCard label="Total Customer" value={meta.total} icon={<Users size={18} />} />
          </div>

          <TableCard title="Daftar Customer" count={meta.total}>
            <thead>
              <tr>
                <th>NAMA</th>
                <th>TELEPON</th>
                <th>EMAIL</th>
                <th>ALAMAT</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {!loading &&
                customers.map((customer) => (
                  <tr key={customer.id}>
                    <td>{customer.name}</td>
                    <td style={{ color: 'var(--text-muted)' }}>{customer.phone || '-'}</td>
                    <td style={{ color: 'var(--text-muted)' }}>{customer.email || '-'}</td>
                    <td style={{ color: 'var(--text-muted)' }}>{customer.address || '-'}</td>
                    <td style={{ textAlign: 'right', display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
                      {can('customer.update') && (
                        <button
                          style={{ background: 'transparent', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}
                          onClick={() => openEditForm(customer)}
                        >
                          <Pencil size={16} />
                        </button>
                      )}
                      {can('customer.delete') && (
                        <button
                          style={{ background: 'transparent', border: 'none', color: 'var(--danger)', cursor: 'pointer' }}
                          onClick={() => setConfirmDelete(customer)}
                        >
                          <Trash2 size={16} />
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
            </tbody>
          </TableCard>

          <Pagination page={page} meta={meta} onChange={setPage} label="customer" />
        </>
      )}

      <Modal
        isOpen={formOpen}
        onClose={() => setFormOpen(false)}
        title={editing ? 'Edit Customer' : 'Tambah Customer'}
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
          <label>Telepon</label>
          <input className="form-input" value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} />
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
        title="Hapus Customer"
        message={`Yakin ingin menghapus customer "${confirmDelete?.name}"?`}
        confirmLabel="Hapus"
        onConfirm={handleDelete}
        onCancel={() => setConfirmDelete(null)}
      />
    </div>
  );
};

export default CustomerPage;
