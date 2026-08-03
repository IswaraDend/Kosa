import { useEffect, useState } from 'react';
import { Plus, Package, Trash2, Pencil } from 'lucide-react';
import { api, ApiError, buildQuery } from '../../lib/api';
import type { Item } from '../../types';
import { useSelectedProject } from '../../hooks/useSelectedProject';
import { useProjectAutoSelect } from '../../hooks/useProjectAutoSelect';
import PageHeader from '../../components/PageHeader';
import StatCard from '../../components/StatCard';
import TableCard from '../../components/TableCard';
import Modal from '../../components/Modal';
import ConfirmDialog from '../../components/ConfirmDialog';
import ProjectPicker from '../../components/ProjectPicker';
import '../Dashboard.css';

interface ItemFormState {
  sku: string;
  name: string;
  unit: string;
}

const emptyForm: ItemFormState = { sku: '', name: '', unit: '' };

interface ItemPageProps {
  apiBasePrefix?: string;
  scopeMode?: 'query' | 'path';
  projectEndpoint?: string;
  includeAllOption?: boolean;
  showProjectPicker?: boolean;
}

const ItemPage = ({
  apiBasePrefix = '/super-admin',
  scopeMode = 'query',
  projectEndpoint = '/super-admin/projects',
  includeAllOption = true,
  showProjectPicker = true,
}: ItemPageProps) => {
  const { selectedProjectId, setSelectedProjectId } = useSelectedProject();
  useProjectAutoSelect(projectEndpoint, !showProjectPicker);
  const [items, setItems] = useState<Item[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState('');

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Item | null>(null);
  const [form, setForm] = useState<ItemFormState>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState<Item | null>(null);

  const resourceBase = scopeMode === 'path' ? `${apiBasePrefix}/${selectedProjectId}` : apiBasePrefix;
  const listQuery = scopeMode === 'query' ? buildQuery({ project_id: selectedProjectId }) : '';

  const loadItems = () => {
    if (selectedProjectId === 'all') {
      setItems([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    api
      .get<{ data: Item[] }>(`${resourceBase}/items${listQuery}`)
      .then((res) => setItems(res.data))
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat data item'))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadItems();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId, apiBasePrefix, scopeMode]);

  const openCreateForm = () => {
    setEditing(null);
    setForm(emptyForm);
    setFormOpen(true);
  };

  const openEditForm = (item: Item) => {
    setEditing(item);
    setForm({ sku: item.sku, name: item.name, unit: item.unit });
    setFormOpen(true);
  };

  const handleSubmit = async () => {
    setSaving(true);
    setErrorMsg('');
    try {
      if (editing) {
        await api.put(`${resourceBase}/items/${editing.id}`, form);
      } else {
        await api.post(`${resourceBase}/items`, { ...form, project_id: Number(selectedProjectId) });
      }
      setFormOpen(false);
      loadItems();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal menyimpan item');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!confirmDelete) return;
    try {
      await api.delete(`${resourceBase}/items/${confirmDelete.id}`);
      setConfirmDelete(null);
      loadItems();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal menghapus item');
      setConfirmDelete(null);
    }
  };

  return (
    <div className="dashboard-content">
      <PageHeader
        title="Item"
        subtitle="Kelola master data barang untuk project terpilih"
        actions={
          <button className="btn-primary" onClick={openCreateForm} disabled={selectedProjectId === 'all'}>
            <Plus size={18} />
            Tambah Item
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
        <p style={{ color: 'var(--text-muted)' }}>Pilih project terlebih dahulu untuk melihat daftar item.</p>
      ) : (
        <>
          <div className="summary-cards" style={{ gridTemplateColumns: 'repeat(1, 1fr)' }}>
            <StatCard label="Total Item" value={items.length} icon={<Package size={18} />} />
          </div>

          <TableCard title="Daftar Item" count={items.length}>
            <thead>
              <tr>
                <th>SKU</th>
                <th>NAMA ITEM</th>
                <th>SATUAN</th>
                <th>DIBUAT</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {!loading &&
                items.map((item) => (
                  <tr key={item.id}>
                    <td style={{ color: 'var(--text-muted)' }}>{item.sku}</td>
                    <td>{item.name}</td>
                    <td style={{ color: 'var(--text-muted)' }}>{item.unit}</td>
                    <td style={{ color: 'var(--text-muted)' }}>
                      {new Date(item.created_at).toLocaleDateString('id-ID')}
                    </td>
                    <td style={{ textAlign: 'right', display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
                      <button
                        style={{ background: 'transparent', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}
                        onClick={() => openEditForm(item)}
                      >
                        <Pencil size={16} />
                      </button>
                      <button
                        style={{ background: 'transparent', border: 'none', color: 'var(--danger)', cursor: 'pointer' }}
                        onClick={() => setConfirmDelete(item)}
                      >
                        <Trash2 size={16} />
                      </button>
                    </td>
                  </tr>
                ))}
            </tbody>
          </TableCard>
        </>
      )}

      <Modal
        isOpen={formOpen}
        onClose={() => setFormOpen(false)}
        title={editing ? 'Edit Item' : 'Tambah Item'}
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
          <label>SKU</label>
          <input className="form-input" value={form.sku} onChange={(e) => setForm({ ...form, sku: e.target.value })} />
        </div>
        <div className="form-group">
          <label>Nama Item</label>
          <input className="form-input" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
        </div>
        <div className="form-group">
          <label>Satuan</label>
          <input
            className="form-input"
            placeholder="pcs, box, kg, dst"
            value={form.unit}
            onChange={(e) => setForm({ ...form, unit: e.target.value })}
          />
        </div>
      </Modal>

      <ConfirmDialog
        isOpen={!!confirmDelete}
        title="Hapus Item"
        message={`Yakin ingin menghapus item "${confirmDelete?.name}"?`}
        confirmLabel="Hapus"
        onConfirm={handleDelete}
        onCancel={() => setConfirmDelete(null)}
      />
    </div>
  );
};

export default ItemPage;
