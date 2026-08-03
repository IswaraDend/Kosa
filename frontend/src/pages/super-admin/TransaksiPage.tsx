import { useEffect, useState } from 'react';
import { Plus, ArrowDownCircle, ArrowUpCircle, ArrowRightLeft } from 'lucide-react';
import { api, ApiError, buildQuery } from '../../lib/api';
import type { Item, TransactionRecord, TransactionType, Warehouse } from '../../types';
import { useSelectedProject } from '../../hooks/useSelectedProject';
import { useProjectAutoSelect } from '../../hooks/useProjectAutoSelect';
import PageHeader from '../../components/PageHeader';
import StatCard from '../../components/StatCard';
import TableCard from '../../components/TableCard';
import Badge from '../../components/Badge';
import Modal from '../../components/Modal';
import ProjectPicker from '../../components/ProjectPicker';
import '../Dashboard.css';

const typeLabel: Record<TransactionType, string> = { in: 'Masuk', out: 'Keluar', transfer: 'Transfer' };
const typeVariant: Record<TransactionType, string> = { in: 'masuk', out: 'keluar', transfer: 'transfer' };

interface TxFormState {
  type: TransactionType;
  source_warehouse_id: string;
  dest_warehouse_id: string;
  item_id: string;
  quantity: string;
  note: string;
}

const emptyForm: TxFormState = {
  type: 'in',
  source_warehouse_id: '',
  dest_warehouse_id: '',
  item_id: '',
  quantity: '',
  note: '',
};

interface TransaksiPageProps {
  apiBasePrefix?: string;
  scopeMode?: 'query' | 'path';
  projectEndpoint?: string;
  includeAllOption?: boolean;
  canCreate?: boolean;
  showProjectPicker?: boolean;
}

const TransaksiPage = ({
  apiBasePrefix = '/super-admin',
  scopeMode = 'query',
  projectEndpoint = '/super-admin/projects',
  includeAllOption = true,
  canCreate = true,
  showProjectPicker = true,
}: TransaksiPageProps) => {
  const { selectedProjectId, setSelectedProjectId } = useSelectedProject();
  useProjectAutoSelect(projectEndpoint, !showProjectPicker);
  const [transactions, setTransactions] = useState<TransactionRecord[]>([]);
  const [warehouses, setWarehouses] = useState<Warehouse[]>([]);
  const [items, setItems] = useState<Item[]>([]);
  const [typeFilter, setTypeFilter] = useState<'all' | TransactionType>('all');
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState('');

  const [formOpen, setFormOpen] = useState(false);
  const [form, setForm] = useState<TxFormState>(emptyForm);
  const [saving, setSaving] = useState(false);

  const resourceBase = scopeMode === 'path' ? `${apiBasePrefix}/${selectedProjectId}` : apiBasePrefix;

  const loadTransactions = () => {
    if (selectedProjectId === 'all') {
      setTransactions([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    const query =
      scopeMode === 'query'
        ? buildQuery({ project_id: selectedProjectId, type: typeFilter === 'all' ? undefined : typeFilter })
        : buildQuery({ type: typeFilter === 'all' ? undefined : typeFilter });
    api
      .get<{ data: TransactionRecord[] }>(`${resourceBase}/transactions${query}`)
      .then((res) => setTransactions(res.data))
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat transaksi'))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadTransactions();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId, typeFilter, apiBasePrefix, scopeMode]);

  useEffect(() => {
    if (selectedProjectId === 'all' || !canCreate) {
      setWarehouses([]);
      setItems([]);
      return;
    }
    const query = scopeMode === 'query' ? buildQuery({ project_id: selectedProjectId }) : '';
    api
      .get<{ data: Warehouse[] }>(`${resourceBase}/warehouses${query}`)
      .then((res) => setWarehouses(res.data))
      .catch(() => setWarehouses([]));
    api
      .get<{ data: Item[] }>(`${resourceBase}/items${query}`)
      .then((res) => setItems(res.data))
      .catch(() => setItems([]));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId, apiBasePrefix, scopeMode, canCreate]);

  const openCreateForm = () => {
    setForm(emptyForm);
    setFormOpen(true);
  };

  const handleSubmit = async () => {
    setSaving(true);
    setErrorMsg('');
    try {
      await api.post(`${resourceBase}/transactions`, {
        project_id: Number(selectedProjectId),
        type: form.type,
        source_warehouse_id: form.source_warehouse_id ? Number(form.source_warehouse_id) : undefined,
        dest_warehouse_id: form.dest_warehouse_id ? Number(form.dest_warehouse_id) : undefined,
        note: form.note,
        items: [{ item_id: Number(form.item_id), quantity: Number(form.quantity) }],
      });
      setFormOpen(false);
      loadTransactions();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal membuat transaksi');
    } finally {
      setSaving(false);
    }
  };

  const totalMasuk = transactions.filter((t) => t.type === 'in').length;
  const totalKeluar = transactions.filter((t) => t.type === 'out').length;
  const totalTransfer = transactions.filter((t) => t.type === 'transfer').length;

  return (
    <div className="dashboard-content">
      <PageHeader
        title="Transaksi"
        subtitle="Riwayat pergerakan stok masuk, keluar, dan transfer antar gudang"
        actions={
          canCreate ? (
            <button className="btn-primary" onClick={openCreateForm} disabled={selectedProjectId === 'all'}>
              <Plus size={18} />
              Tambah Transaksi
            </button>
          ) : undefined
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
        <p style={{ color: 'var(--text-muted)' }}>Pilih project terlebih dahulu untuk melihat transaksi.</p>
      ) : (
        <>
          <div className="summary-cards" style={{ gridTemplateColumns: 'repeat(3, 1fr)' }}>
            <StatCard label="Total Masuk" value={totalMasuk} icon={<ArrowDownCircle size={18} />} />
            <StatCard label="Total Keluar" value={totalKeluar} icon={<ArrowUpCircle size={18} />} />
            <StatCard label="Total Transfer" value={totalTransfer} icon={<ArrowRightLeft size={18} />} />
          </div>

          <div className="filter-group">
            {(['all', 'in', 'out', 'transfer'] as const).map((t) => (
              <button
                key={t}
                className={`filter-btn ${typeFilter === t ? 'active' : ''}`}
                onClick={() => setTypeFilter(t)}
              >
                {t === 'all' ? 'Semua' : typeLabel[t]}
              </button>
            ))}
          </div>

          <TableCard title="Riwayat Transaksi" count={transactions.length}>
            <thead>
              <tr>
                <th>TIPE</th>
                <th>ITEM</th>
                <th>QTY</th>
                <th>DARI</th>
                <th>KE</th>
                <th>CATATAN</th>
                <th>TANGGAL</th>
              </tr>
            </thead>
            <tbody>
              {!loading &&
                transactions.map((t) => (
                  <tr key={t.id}>
                    <td>
                      <Badge label={typeLabel[t.type]} variant={typeVariant[t.type]} />
                    </td>
                    <td>{t.items.map((i) => i.item?.name).filter(Boolean).join(', ') || '-'}</td>
                    <td>{t.items.map((i) => i.quantity).join(', ')}</td>
                    <td style={{ color: 'var(--text-muted)' }}>{t.source_warehouse?.name || '-'}</td>
                    <td style={{ color: 'var(--text-muted)' }}>{t.dest_warehouse?.name || '-'}</td>
                    <td style={{ color: 'var(--text-muted)' }}>{t.note || '-'}</td>
                    <td style={{ color: 'var(--text-muted)' }}>
                      {new Date(t.created_at).toLocaleString('id-ID')}
                    </td>
                  </tr>
                ))}
            </tbody>
          </TableCard>
        </>
      )}

      {canCreate && (
        <Modal
          isOpen={formOpen}
          onClose={() => setFormOpen(false)}
          title="Tambah Transaksi"
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
            <label>Tipe Transaksi</label>
            <select
              className="form-select"
              value={form.type}
              onChange={(e) => setForm({ ...form, type: e.target.value as TransactionType })}
            >
              <option value="in">Masuk</option>
              <option value="out">Keluar</option>
              <option value="transfer">Transfer</option>
            </select>
          </div>

          {(form.type === 'out' || form.type === 'transfer') && (
            <div className="form-group">
              <label>Gudang Asal</label>
              <select
                className="form-select"
                value={form.source_warehouse_id}
                onChange={(e) => setForm({ ...form, source_warehouse_id: e.target.value })}
              >
                <option value="">Pilih gudang</option>
                {warehouses.map((w) => (
                  <option key={w.id} value={w.id}>
                    {w.name}
                  </option>
                ))}
              </select>
            </div>
          )}

          {(form.type === 'in' || form.type === 'transfer') && (
            <div className="form-group">
              <label>Gudang Tujuan</label>
              <select
                className="form-select"
                value={form.dest_warehouse_id}
                onChange={(e) => setForm({ ...form, dest_warehouse_id: e.target.value })}
              >
                <option value="">Pilih gudang</option>
                {warehouses.map((w) => (
                  <option key={w.id} value={w.id}>
                    {w.name}
                  </option>
                ))}
              </select>
            </div>
          )}

          <div className="form-group">
            <label>Item</label>
            <select className="form-select" value={form.item_id} onChange={(e) => setForm({ ...form, item_id: e.target.value })}>
              <option value="">Pilih item</option>
              {items.map((i) => (
                <option key={i.id} value={i.id}>
                  {i.name} ({i.unit})
                </option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label>Jumlah</label>
            <input
              type="number"
              className="form-input"
              value={form.quantity}
              onChange={(e) => setForm({ ...form, quantity: e.target.value })}
            />
          </div>

          <div className="form-group">
            <label>Catatan</label>
            <textarea
              className="form-textarea"
              rows={2}
              value={form.note}
              onChange={(e) => setForm({ ...form, note: e.target.value })}
            />
          </div>
        </Modal>
      )}
    </div>
  );
};

export default TransaksiPage;
