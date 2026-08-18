import { useEffect, useState } from 'react';
import { Plus, PackagePlus, Trash2, Pencil, ListTree, UploadCloud } from 'lucide-react';
import { api, ApiError, buildQuery } from '../../lib/api';
import { formatCurrency } from '../../lib/format';
import type { Item, Product, ProductRecipeLine } from '../../types';
import { useSelectedProject } from '../../hooks/useSelectedProject';
import { useProjectAutoSelect } from '../../hooks/useProjectAutoSelect';
import PageHeader from '../../components/PageHeader';
import StatCard from '../../components/StatCard';
import TableCard from '../../components/TableCard';
import Modal from '../../components/Modal';
import ConfirmDialog from '../../components/ConfirmDialog';
import ProjectPicker from '../../components/ProjectPicker';
import ImportModal from '../../components/ImportModal';
import '../Dashboard.css';

interface ProductFormState {
  sku: string;
  name: string;
  unit: string;
  default_price: string;
}

const emptyForm: ProductFormState = { sku: '', name: '', unit: '', default_price: '' };

interface ProductPageProps {
  apiBasePrefix?: string;
  scopeMode?: 'query' | 'path';
  projectEndpoint?: string;
  includeAllOption?: boolean;
  showProjectPicker?: boolean;
}

const ProductPage = ({
  apiBasePrefix = '/super-admin',
  scopeMode = 'query',
  projectEndpoint = '/super-admin/projects',
  includeAllOption = true,
  showProjectPicker = true,
}: ProductPageProps) => {
  const { selectedProjectId, setSelectedProjectId } = useSelectedProject();
  useProjectAutoSelect(projectEndpoint, !showProjectPicker);
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState('');

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Product | null>(null);
  const [form, setForm] = useState<ProductFormState>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState<Product | null>(null);
  const [importOpen, setImportOpen] = useState(false);

  const [recipeProduct, setRecipeProduct] = useState<Product | null>(null);
  const [recipe, setRecipe] = useState<ProductRecipeLine[]>([]);
  const [items, setItems] = useState<Item[]>([]);
  const [recipeForm, setRecipeForm] = useState({ item_id: '', quantity_per_unit: '' });
  const [recipeError, setRecipeError] = useState('');

  const resourceBase = scopeMode === 'path' ? `${apiBasePrefix}/${selectedProjectId}` : apiBasePrefix;
  const listQuery = scopeMode === 'query' ? buildQuery({ project_id: selectedProjectId }) : '';

  const loadProducts = () => {
    if (selectedProjectId === 'all') {
      setProducts([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    api
      .get<{ data: Product[] }>(`${resourceBase}/products${listQuery}`)
      .then((res) => setProducts(res.data))
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat data produk'))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadProducts();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId, apiBasePrefix, scopeMode]);

  const openCreateForm = () => {
    setEditing(null);
    setForm(emptyForm);
    setFormOpen(true);
  };

  const openEditForm = (product: Product) => {
    setEditing(product);
    setForm({
      sku: product.sku,
      name: product.name,
      unit: product.unit,
      default_price: product.default_price ? String(product.default_price) : '',
    });
    setFormOpen(true);
  };

  const handleSubmit = async () => {
    setSaving(true);
    setErrorMsg('');
    try {
      const payload = { ...form, default_price: form.default_price ? Number(form.default_price) : 0 };
      if (editing) {
        await api.put(`${resourceBase}/products/${editing.id}`, payload);
      } else {
        await api.post(`${resourceBase}/products`, { ...payload, project_id: Number(selectedProjectId) });
      }
      setFormOpen(false);
      loadProducts();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal menyimpan produk');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!confirmDelete) return;
    try {
      await api.delete(`${resourceBase}/products/${confirmDelete.id}`);
      setConfirmDelete(null);
      loadProducts();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal menghapus produk');
      setConfirmDelete(null);
    }
  };

  const loadRecipe = (productId: number) => {
    api
      .get<{ data: ProductRecipeLine[] }>(`${resourceBase}/products/${productId}/recipe`)
      .then((res) => setRecipe(res.data))
      .catch(() => setRecipe([]));
  };

  const openRecipeModal = (product: Product) => {
    setRecipeProduct(product);
    setRecipeForm({ item_id: '', quantity_per_unit: '' });
    setRecipeError('');
    loadRecipe(product.id);
    api
      .get<{ data: Item[] }>(`${resourceBase}/items${listQuery}`)
      .then((res) => setItems(res.data))
      .catch(() => setItems([]));
  };

  const handleAddRecipeLine = async () => {
    if (!recipeProduct || !recipeForm.item_id || !recipeForm.quantity_per_unit) return;
    setRecipeError('');
    try {
      await api.post(`${resourceBase}/products/${recipeProduct.id}/recipe`, {
        item_id: Number(recipeForm.item_id),
        quantity_per_unit: Number(recipeForm.quantity_per_unit),
      });
      setRecipeForm({ item_id: '', quantity_per_unit: '' });
      loadRecipe(recipeProduct.id);
    } catch (err) {
      setRecipeError(err instanceof ApiError ? err.message : 'Gagal menambah komponen resep');
    }
  };

  const handleRemoveRecipeLine = async (recipeId: number) => {
    if (!recipeProduct) return;
    try {
      await api.delete(`${resourceBase}/products/${recipeProduct.id}/recipe/${recipeId}`);
      loadRecipe(recipeProduct.id);
    } catch (err) {
      setRecipeError(err instanceof ApiError ? err.message : 'Gagal menghapus komponen resep');
    }
  };

  return (
    <div className="dashboard-content">
      <PageHeader
        title="Produk"
        subtitle="Kelola barang jadi beserta resep (BOM) bahan bakunya"
        actions={
          <>
            <button className="btn-secondary" onClick={() => setImportOpen(true)} disabled={selectedProjectId === 'all'}>
              <UploadCloud size={18} />
              Import Stok
            </button>
            <button className="btn-primary" onClick={openCreateForm} disabled={selectedProjectId === 'all'}>
              <Plus size={18} />
              Tambah Produk
            </button>
          </>
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
        <p style={{ color: 'var(--text-muted)' }}>Pilih project terlebih dahulu untuk melihat daftar produk.</p>
      ) : (
        <>
          <div className="summary-cards" style={{ gridTemplateColumns: 'repeat(1, 1fr)' }}>
            <StatCard label="Total Produk" value={products.length} icon={<PackagePlus size={18} />} />
          </div>

          <TableCard title="Daftar Produk" count={products.length}>
            <thead>
              <tr>
                <th>SKU</th>
                <th>NAMA PRODUK</th>
                <th>SATUAN</th>
                <th>HPP/UNIT</th>
                <th>HARGA JUAL</th>
                <th>DIBUAT</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {!loading &&
                products.map((product) => (
                  <tr key={product.id}>
                    <td style={{ color: 'var(--text-muted)' }}>{product.sku}</td>
                    <td>{product.name}</td>
                    <td style={{ color: 'var(--text-muted)' }}>{product.unit}</td>
                    <td style={{ color: 'var(--text-muted)' }}>
                      {product.average_cost ? formatCurrency(product.average_cost) : '-'}
                    </td>
                    <td style={{ color: 'var(--text-muted)' }}>
                      {product.default_price ? formatCurrency(product.default_price) : '-'}
                    </td>
                    <td style={{ color: 'var(--text-muted)' }}>
                      {new Date(product.created_at).toLocaleDateString('id-ID')}
                    </td>
                    <td style={{ textAlign: 'right', display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
                      <button
                        style={{ background: 'transparent', border: 'none', color: 'var(--primary)', cursor: 'pointer' }}
                        onClick={() => openRecipeModal(product)}
                        title="Kelola Resep"
                      >
                        <ListTree size={16} />
                      </button>
                      <button
                        style={{ background: 'transparent', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}
                        onClick={() => openEditForm(product)}
                      >
                        <Pencil size={16} />
                      </button>
                      <button
                        style={{ background: 'transparent', border: 'none', color: 'var(--danger)', cursor: 'pointer' }}
                        onClick={() => setConfirmDelete(product)}
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
        title={editing ? 'Edit Produk' : 'Tambah Produk'}
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
          <label>Nama Produk</label>
          <input className="form-input" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
        </div>
        <div className="form-group">
          <label>Satuan</label>
          <input
            className="form-input"
            placeholder="pcs, box, paket, dst"
            value={form.unit}
            onChange={(e) => setForm({ ...form, unit: e.target.value })}
          />
        </div>
        <div className="form-group">
          <label>Harga Jual Default (opsional)</label>
          <input
            type="number"
            min="0"
            step="any"
            className="form-input"
            placeholder="Dipakai isi awal form Invoice, tetap bisa diubah"
            value={form.default_price}
            onChange={(e) => setForm({ ...form, default_price: e.target.value })}
          />
        </div>
      </Modal>

      <Modal
        isOpen={!!recipeProduct}
        onClose={() => setRecipeProduct(null)}
        title={`Kelola Resep — ${recipeProduct?.name ?? ''}`}
        footer={
          <button className="btn-secondary" onClick={() => setRecipeProduct(null)}>
            Tutup
          </button>
        }
      >
        {recipeError && (
          <div style={{ color: 'var(--danger)', marginBottom: '12px', fontSize: '13px' }}>{recipeError}</div>
        )}

        {recipe.length === 0 ? (
          <p style={{ color: 'var(--text-muted)', fontSize: '13px' }}>Belum ada komponen resep.</p>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 16 }}>
            {recipe.map((line) => (
              <div
                key={line.id}
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  padding: '8px 12px',
                  background: 'var(--bg-color)',
                  border: '1px solid var(--border-color)',
                  borderRadius: 8,
                }}
              >
                <span style={{ fontSize: 13 }}>
                  {line.item?.name ?? `Item #${line.item_id}`}{' '}
                  <span style={{ color: 'var(--text-muted)' }}>
                    ({line.quantity_per_unit} {line.item?.unit ?? ''} / unit)
                  </span>
                </span>
                <button
                  style={{ background: 'transparent', border: 'none', color: 'var(--danger)', cursor: 'pointer' }}
                  onClick={() => handleRemoveRecipeLine(line.id)}
                >
                  <Trash2 size={15} />
                </button>
              </div>
            ))}
          </div>
        )}

        <div style={{ display: 'flex', gap: 8, alignItems: 'flex-end' }}>
          <div className="form-group" style={{ flex: 1, marginBottom: 0 }}>
            <label>Item Bahan Baku</label>
            <select
              className="form-select"
              value={recipeForm.item_id}
              onChange={(e) => setRecipeForm({ ...recipeForm, item_id: e.target.value })}
            >
              <option value="">Pilih item</option>
              {items.map((i) => (
                <option key={i.id} value={i.id}>
                  {i.name} ({i.unit})
                </option>
              ))}
            </select>
          </div>
          <div className="form-group" style={{ width: 100, marginBottom: 0 }}>
            <label>Qty / unit</label>
            <input
              type="number"
              className="form-input"
              value={recipeForm.quantity_per_unit}
              onChange={(e) => setRecipeForm({ ...recipeForm, quantity_per_unit: e.target.value })}
            />
          </div>
          <button className="btn-primary" onClick={handleAddRecipeLine}>
            <Plus size={16} />
          </button>
        </div>
      </Modal>

      <ConfirmDialog
        isOpen={!!confirmDelete}
        title="Hapus Produk"
        message={`Yakin ingin menghapus produk "${confirmDelete?.name}"?`}
        confirmLabel="Hapus"
        onConfirm={handleDelete}
        onCancel={() => setConfirmDelete(null)}
      />

      <ImportModal
        isOpen={importOpen}
        onClose={() => setImportOpen(false)}
        title="Import Stok Produk"
        endpoint={`${resourceBase}/imports/products/stock`}
        projectId={scopeMode === 'query' ? selectedProjectId : undefined}
        templateHint="Kolom: SKU | Gudang (kode) | Qty | HPP per Unit (opsional). SKU dan Gudang harus sudah terdaftar di project ini. HPP per unit boleh dikosongkan kalau hanya ingin koreksi jumlah stok."
        onSuccess={loadProducts}
      />
    </div>
  );
};

export default ProductPage;
