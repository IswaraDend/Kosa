import { useEffect, useState } from 'react';
import { Plus, Factory } from 'lucide-react';
import { api, ApiError, buildQuery , type PaginatedResponse } from '../../lib/api';
import { formatCurrency } from '../../lib/format';
import type { Product, ProductRecipeLine, ProductionRecord, Warehouse } from '../../types';
import { useSelectedProject } from '../../hooks/useSelectedProject';
import { useProjectAutoSelect } from '../../hooks/useProjectAutoSelect';
import { usePagination, metaFrom, PER_PAGE } from '../../hooks/usePagination';
import PageHeader from '../../components/PageHeader';
import StatCard from '../../components/StatCard';
import TableCard from '../../components/TableCard';
import Pagination from '../../components/Pagination';
import Modal from '../../components/Modal';
import ProjectPicker from '../../components/ProjectPicker';
import { makeCan } from '../../lib/permissions';
import '../Dashboard.css';

interface ProduksiFormState {
  warehouse_id: string;
  product_id: string;
  quantity: string;
  note: string;
}

const emptyForm: ProduksiFormState = { warehouse_id: '', product_id: '', quantity: '', note: '' };

interface ProduksiPageProps {
  apiBasePrefix?: string;
  scopeMode?: 'query' | 'path';
  projectEndpoint?: string;
  includeAllOption?: boolean;
  showProjectPicker?: boolean;
  /** Granted permission codes. Undefined = no per-action gating (Super Admin / Admin). */
  permissions?: string[];
}

const ProduksiPage = ({
  apiBasePrefix = '/super-admin',
  scopeMode = 'query',
  projectEndpoint = '/super-admin/projects',
  includeAllOption = true,
  showProjectPicker = true,
  permissions,
}: ProduksiPageProps) => {
  const can = makeCan(permissions);
  const { selectedProjectId, setSelectedProjectId } = useSelectedProject();
  useProjectAutoSelect(projectEndpoint, !showProjectPicker);
  const { page, setPage, meta, setMeta } = usePagination(selectedProjectId);
  const [productions, setProductions] = useState<ProductionRecord[]>([]);
  const [warehouses, setWarehouses] = useState<Warehouse[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState('');

  const [formOpen, setFormOpen] = useState(false);
  const [form, setForm] = useState<ProduksiFormState>(emptyForm);
  const [recipePreview, setRecipePreview] = useState<ProductRecipeLine[]>([]);
  const [saving, setSaving] = useState(false);

  const resourceBase = scopeMode === 'path' ? `${apiBasePrefix}/${selectedProjectId}` : apiBasePrefix;
  const listQuery = scopeMode === 'query' ? buildQuery({ project_id: selectedProjectId }) : '';
  // Separate from listQuery on purpose: listQuery still scopes the auxiliary
  // dropdown fetches (warehouses, items, customers) which must stay complete,
  // while only the main table asks for a page.
  const pagedQuery = buildQuery({
    ...(scopeMode === 'query' ? { project_id: selectedProjectId } : {}),
    page,
    per_page: PER_PAGE,
  });


  const loadProductions = () => {
    if (selectedProjectId === 'all') {
      setProductions([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    api
      .get<PaginatedResponse<ProductionRecord>>(`${resourceBase}/productions${pagedQuery}`)
      .then((res) => {
        setProductions(res.data);
        setMeta(metaFrom(res));
      })
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat riwayat produksi'))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadProductions();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId, apiBasePrefix, scopeMode, page]);

  useEffect(() => {
    if (selectedProjectId === 'all') {
      setWarehouses([]);
      setProducts([]);
      return;
    }
    api
      .get<{ data: Warehouse[] }>(`${resourceBase}/warehouses${listQuery}`)
      .then((res) => setWarehouses(res.data))
      .catch(() => setWarehouses([]));
    api
      .get<{ data: Product[] }>(`${resourceBase}/products${listQuery}`)
      .then((res) => setProducts(res.data))
      .catch(() => setProducts([]));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId, apiBasePrefix, scopeMode]);

  const openCreateForm = () => {
    setForm(emptyForm);
    setRecipePreview([]);
    setFormOpen(true);
  };

  const handleProductChange = (productId: string) => {
    setForm({ ...form, product_id: productId });
    if (!productId) {
      setRecipePreview([]);
      return;
    }
    api
      .get<{ data: ProductRecipeLine[] }>(`${resourceBase}/products/${productId}/recipe`)
      .then((res) => setRecipePreview(res.data))
      .catch(() => setRecipePreview([]));
  };

  const handleSubmit = async () => {
    setSaving(true);
    setErrorMsg('');
    try {
      await api.post(`${resourceBase}/productions`, {
        project_id: Number(selectedProjectId),
        warehouse_id: Number(form.warehouse_id),
        product_id: Number(form.product_id),
        quantity: Number(form.quantity),
        note: form.note,
      });
      setFormOpen(false);
      loadProductions();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal membuat produksi');
    } finally {
      setSaving(false);
    }
  };

  const quantity = Number(form.quantity) || 0;
  const hppPerUnitPreview = recipePreview.reduce(
    (sum, line) => sum + line.quantity_per_unit * (line.item?.average_cost || 0),
    0,
  );

  return (
    <div className="dashboard-content">
      <PageHeader
        title="Produksi"
        subtitle="Rakit produk dari bahan baku — stok bahan otomatis berkurang sesuai resep"
        actions={
          can('production.create') && (
            <button className="btn-primary" onClick={openCreateForm} disabled={selectedProjectId === 'all'}>
              <Plus size={18} />
              Buat Produksi
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
        <p style={{ color: 'var(--text-muted)' }}>Pilih project terlebih dahulu untuk melihat riwayat produksi.</p>
      ) : (
        <>
          <div className="summary-cards" style={{ gridTemplateColumns: 'repeat(1, 1fr)' }}>
            <StatCard label="Total Produksi" value={meta.total} icon={<Factory size={18} />} />
          </div>

          <TableCard title="Riwayat Produksi" count={meta.total}>
            <thead>
              <tr>
                <th>PRODUK</th>
                <th>GUDANG</th>
                <th>QTY</th>
                <th>HPP TOTAL</th>
                <th>CATATAN</th>
                <th>TANGGAL</th>
              </tr>
            </thead>
            <tbody>
              {!loading &&
                productions.map((p) => (
                  <tr key={p.id}>
                    <td>{p.product?.name ?? `Produk #${p.product_id}`}</td>
                    <td style={{ color: 'var(--text-muted)' }}>{p.warehouse?.name ?? '-'}</td>
                    <td>
                      {p.quantity} {p.product?.unit ?? ''}
                    </td>
                    <td style={{ color: 'var(--text-muted)' }}>{formatCurrency(p.hpp_total)}</td>
                    <td style={{ color: 'var(--text-muted)' }}>{p.note || '-'}</td>
                    <td style={{ color: 'var(--text-muted)' }}>
                      {new Date(p.created_at).toLocaleString('id-ID')}
                    </td>
                  </tr>
                ))}
            </tbody>
          </TableCard>

          <Pagination page={page} meta={meta} onChange={setPage} label="produksi" />
        </>
      )}

      <Modal
        isOpen={formOpen}
        onClose={() => setFormOpen(false)}
        title="Buat Produksi"
        footer={
          <>
            <button className="btn-secondary" onClick={() => setFormOpen(false)}>
              Batal
            </button>
            <button className="btn-primary" onClick={handleSubmit} disabled={saving}>
              {saving ? 'Memproses...' : 'Produksi'}
            </button>
          </>
        }
      >
        <div className="form-group">
          <label>Produk</label>
          <select className="form-select" value={form.product_id} onChange={(e) => handleProductChange(e.target.value)}>
            <option value="">Pilih produk</option>
            {products.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name} ({p.unit})
              </option>
            ))}
          </select>
        </div>

        <div className="form-group">
          <label>Gudang</label>
          <select
            className="form-select"
            value={form.warehouse_id}
            onChange={(e) => setForm({ ...form, warehouse_id: e.target.value })}
          >
            <option value="">Pilih gudang</option>
            {warehouses.map((w) => (
              <option key={w.id} value={w.id}>
                {w.name}
              </option>
            ))}
          </select>
        </div>

        <div className="form-group">
          <label>Jumlah Diproduksi</label>
          <input
            type="number"
            className="form-input"
            value={form.quantity}
            onChange={(e) => setForm({ ...form, quantity: e.target.value })}
          />
        </div>

        {recipePreview.length > 0 && (
          <div className="form-group">
            <label>Bahan Baku yang Akan Dikonsumsi</label>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              {recipePreview.map((line) => (
                <div
                  key={line.id}
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    fontSize: 13,
                    padding: '6px 10px',
                    background: 'var(--bg-color)',
                    border: '1px solid var(--border-color)',
                    borderRadius: 8,
                  }}
                >
                  <span>{line.item?.name ?? `Item #${line.item_id}`}</span>
                  <span style={{ color: 'var(--text-muted)' }}>
                    {line.quantity_per_unit * quantity} {line.item?.unit ?? ''}
                  </span>
                </div>
              ))}
            </div>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                fontSize: 13,
                marginTop: 8,
                paddingTop: 8,
                borderTop: '1px solid var(--border-color)',
              }}
            >
              <span style={{ color: 'var(--text-muted)' }}>Estimasi HPP</span>
              <strong>
                {formatCurrency(hppPerUnitPreview)} / unit &times; {quantity} ={' '}
                {formatCurrency(hppPerUnitPreview * quantity)}
              </strong>
            </div>
          </div>
        )}

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
    </div>
  );
};

export default ProduksiPage;
