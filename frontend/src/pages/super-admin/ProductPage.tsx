import { useEffect, useState } from 'react';
import { Plus, PackagePlus, Trash2, Pencil, ListTree, UploadCloud, Boxes } from 'lucide-react';
import { api, ApiError, buildQuery , type PaginatedResponse } from '../../lib/api';
import { formatCurrency } from '../../lib/format';
import type { Item, Product, ProductCategory, ProductRecipeLine, ProductStockRow } from '../../types';
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
import ImportModal from '../../components/ImportModal';
import { makeCan } from '../../lib/permissions';
import '../Dashboard.css';

interface ProductFormState {
  sku: string;
  name: string;
  unit: string;
  default_price: string;
  // Held as a string because that is what a <select> gives back; '' is the
  // "Tanpa kategori" option, which the API accepts as 0.
  category_id: string;
}

const emptyForm: ProductFormState = { sku: '', name: '', unit: '', default_price: '', category_id: '' };

interface ProductPageProps {
  apiBasePrefix?: string;
  scopeMode?: 'query' | 'path';
  projectEndpoint?: string;
  includeAllOption?: boolean;
  showProjectPicker?: boolean;
  /**
   * Enabled modules of the current project. Leave undefined for Super Admin,
   * which is deliberately not module-gated and always sees the full page.
   */
  enabledModules?: string[];
  /** Granted permission codes. Undefined = no per-action gating (Super Admin / Admin). */
  permissions?: string[];
}

const ProductPage = ({
  apiBasePrefix = '/super-admin',
  scopeMode = 'query',
  projectEndpoint = '/super-admin/projects',
  includeAllOption = true,
  showProjectPicker = true,
  enabledModules,
  permissions,
}: ProductPageProps) => {
  const can = makeCan(permissions);
  const canCreate = can('product.create');
  const canUpdate = can('product.update');
  const canDelete = can('product.delete');
  const canImport = can('product.import');
  // Editing the bill of materials is editing the product — same permission
  // the backend requires on POST/DELETE /products/:id/recipe.
  const canManageRecipe = can('product.update');
  const { selectedProjectId, setSelectedProjectId } = useSelectedProject();
  useProjectAutoSelect(projectEndpoint, !showProjectPicker);
  const { page, setPage, meta, setMeta } = usePagination(selectedProjectId);
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState('');

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Product | null>(null);
  const [form, setForm] = useState<ProductFormState>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState<Product | null>(null);
  const [importOpen, setImportOpen] = useState(false);

  const [stockProduct, setStockProduct] = useState<Product | null>(null);
  const [stockRows, setStockRows] = useState<ProductStockRow[]>([]);
  const [stockForm, setStockForm] = useState({ quantity: '', unit_cost: '' });
  const [stockError, setStockError] = useState('');
  const [stockSaving, setStockSaving] = useState(false);

  const [categories, setCategories] = useState<ProductCategory[]>([]);

  const [recipeProduct, setRecipeProduct] = useState<Product | null>(null);
  const [recipe, setRecipe] = useState<ProductRecipeLine[]>([]);
  const [items, setItems] = useState<Item[]>([]);
  const [recipeForm, setRecipeForm] = useState({ item_id: '', quantity_per_unit: '' });
  const [recipeError, setRecipeError] = useState('');

  const resourceBase = scopeMode === 'path' ? `${apiBasePrefix}/${selectedProjectId}` : apiBasePrefix;
  const listQuery = scopeMode === 'query' ? buildQuery({ project_id: selectedProjectId }) : '';
  // Separate from listQuery on purpose: listQuery still scopes the auxiliary
  // dropdown fetches (items, customers) which must stay complete,
  // while only the main table asks for a page.
  const pagedQuery = buildQuery({
    ...(scopeMode === 'query' ? { project_id: selectedProjectId } : {}),
    page,
    per_page: PER_PAGE,
  });


  // On its own, Produk is just a catalogue (SKU/nama/satuan/harga). Stock and
  // HPP no longer hang off another module: ProductStock is one balance per
  // project now rather than one per warehouse, so both doors onto it are open
  // wherever the person holds product.import.
  //
  // The BOM editor is the one part still gated — it picks raw materials, so it
  // needs Item.
  const hasModule = (code: string) => !enabledModules || enabledModules.includes(code);
  const showImport = canImport;
  const showRecipe = canManageRecipe && hasModule('item');
  // Same capability as the Excel import — see the shared product.import
  // permission on the member route — just one row at a time, which is what
  // makes stock possible at all when Produksi is out of scope.
  const showStockEntry = canImport;


  const loadProducts = () => {
    if (selectedProjectId === 'all') {
      setProducts([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    api
      .get<PaginatedResponse<Product>>(`${resourceBase}/products${pagedQuery}`)
      .then((res) => {
        setProducts(res.data);
        setMeta(metaFrom(res));
      })
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat data produk'))
      .finally(() => setLoading(false));
  };

  // Unpaginated on purpose: this list fills a <select>, so a second page would
  // silently hide categories from the person choosing one.
  const loadCategories = () => {
    if (selectedProjectId === 'all') {
      setCategories([]);
      return;
    }
    api
      .get<{ data: ProductCategory[] }>(`${resourceBase}/product-categories${listQuery}`)
      .then((res) => setCategories(res.data))
      .catch(() => setCategories([]));
  };

  useEffect(() => {
    loadProducts();
    loadCategories();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId, apiBasePrefix, scopeMode, page]);

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
      category_id: product.category_id ? String(product.category_id) : '',
    });
    setFormOpen(true);
  };

  const handleSubmit = async () => {
    setSaving(true);
    setErrorMsg('');
    try {
      const payload = {
        ...form,
        default_price: form.default_price ? Number(form.default_price) : 0,
        // 0 rather than null: the API reads both as "Tanpa kategori", and 0
        // survives the round trip through the empty <select> option.
        category_id: form.category_id ? Number(form.category_id) : 0,
      };
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

  const loadStock = (productId: number) => {
    api
      .get<{ data: ProductStockRow[] }>(`${resourceBase}/products/${productId}/stock`)
      .then((res) => setStockRows(res.data))
      .catch(() => setStockRows([]));
  };

  const openStockModal = (product: Product) => {
    setStockProduct(product);
    setStockForm({ quantity: '', unit_cost: '' });
    setStockError('');
    loadStock(product.id);
  };

  const handleAdjustStock = async () => {
    if (!stockProduct || !stockForm.quantity) return;
    setStockSaving(true);
    setStockError('');
    try {
      await api.post(`${resourceBase}/products/${stockProduct.id}/stock`, {
        quantity: Number(stockForm.quantity),
        unit_cost: stockForm.unit_cost ? Number(stockForm.unit_cost) : 0,
      });
      setStockForm({ ...stockForm, quantity: '', unit_cost: '' });
      loadStock(stockProduct.id);
      // A costed addition moves Product.AverageCost, which the table shows.
      loadProducts();
    } catch (err) {
      setStockError(err instanceof ApiError ? err.message : 'Gagal menyimpan stok produk');
    } finally {
      setStockSaving(false);
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
        subtitle={
          showRecipe ? 'Kelola barang jadi beserta resep (BOM) bahan bakunya' : 'Katalog barang jadi'
        }
        actions={
          <>
            {showImport && (
              <button
                className="btn-secondary"
                onClick={() => setImportOpen(true)}
                disabled={selectedProjectId === 'all'}
              >
                <UploadCloud size={18} />
                Import Stok
              </button>
            )}

            {canCreate && (
              <button className="btn-primary" onClick={openCreateForm} disabled={selectedProjectId === 'all'}>
                <Plus size={18} />
                Tambah Produk
              </button>
            )}
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
            <StatCard label="Total Produk" value={meta.total} icon={<PackagePlus size={18} />} />
          </div>

          <TableCard title="Daftar Produk" count={meta.total}>
            <thead>
              <tr>
                <th>SKU</th>
                <th>NAMA PRODUK</th>
                <th>KATEGORI</th>
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
                    <td style={{ color: 'var(--text-muted)' }}>{product.category?.name ?? '-'}</td>
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
                      {showStockEntry && (
                        <button
                          style={{ background: 'transparent', border: 'none', color: 'var(--primary)', cursor: 'pointer' }}
                          onClick={() => openStockModal(product)}
                          title="Atur Stok"
                        >
                          <Boxes size={16} />
                        </button>
                      )}
                      {showRecipe && (
                        <button
                          style={{ background: 'transparent', border: 'none', color: 'var(--primary)', cursor: 'pointer' }}
                          onClick={() => openRecipeModal(product)}
                          title="Kelola Resep"
                        >
                          <ListTree size={16} />
                        </button>
                      )}
                      {canUpdate && (
                        <button
                          style={{ background: 'transparent', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}
                          onClick={() => openEditForm(product)}
                          title="Ubah Produk"
                        >
                          <Pencil size={16} />
                        </button>
                      )}
                      {canDelete && (
                        <button
                          style={{ background: 'transparent', border: 'none', color: 'var(--danger)', cursor: 'pointer' }}
                          onClick={() => setConfirmDelete(product)}
                          title="Hapus Produk"
                        >
                          <Trash2 size={16} />
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
            </tbody>
          </TableCard>

          <Pagination page={page} meta={meta} onChange={setPage} label="produk" />
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
          <label>Kategori (opsional)</label>
          <select
            className="form-select"
            value={form.category_id}
            onChange={(e) => setForm({ ...form, category_id: e.target.value })}
          >
            <option value="">Tanpa kategori</option>
            {categories.map((cat) => (
              <option key={cat.id} value={cat.id}>
                {cat.name}
              </option>
            ))}
          </select>
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
        isOpen={!!stockProduct}
        onClose={() => setStockProduct(null)}
        title={`Atur Stok — ${stockProduct?.name ?? ''}`}
        footer={
          <button className="btn-secondary" onClick={() => setStockProduct(null)}>
            Tutup
          </button>
        }
      >
        <p style={{ fontSize: 13, color: 'var(--text-muted)', marginBottom: 12 }}>
          Stok awal atau koreksi, tanpa lewat Produksi. Qty positif menambah, qty minus mengurangi. HPP per unit
          hanya dipakai saat menambah — nilainya ikut menghitung rata-rata tertimbang HPP produk ini.
        </p>

        {stockError && (
          <div style={{ color: 'var(--danger)', marginBottom: '12px', fontSize: '13px' }}>{stockError}</div>
        )}

        {/* A single balance, not a list: ProductStock is one row per product
            per project now that warehouses are out of the finished-goods
            path, so there is nothing left to break it down by. */}
        {stockRows.length === 0 ? (
          <p style={{ color: 'var(--text-muted)', fontSize: '13px' }}>Belum ada stok tercatat.</p>
        ) : (
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              padding: '8px 12px',
              marginBottom: 16,
              background: 'var(--bg-color)',
              border: '1px solid var(--border-color)',
              borderRadius: 8,
            }}
          >
            <span style={{ fontSize: 13 }}>Stok saat ini</span>
            <span style={{ fontSize: 13, color: 'var(--text-muted)' }}>
              {stockRows[0].quantity} {stockProduct?.unit ?? ''}
            </span>
          </div>
        )}

        <div style={{ display: 'flex', gap: 8 }}>
          <div className="form-group" style={{ flex: 1 }}>
            <label>Qty</label>
            <input
              type="number"
              step="any"
              className="form-input"
              placeholder="mis. 20 atau -5"
              value={stockForm.quantity}
              onChange={(e) => setStockForm({ ...stockForm, quantity: e.target.value })}
            />
          </div>
          <div className="form-group" style={{ flex: 1 }}>
            <label>HPP / unit (opsional)</label>
            <input
              type="number"
              min="0"
              step="any"
              className="form-input"
              placeholder="Kosongkan untuk koreksi jumlah saja"
              value={stockForm.unit_cost}
              onChange={(e) => setStockForm({ ...stockForm, unit_cost: e.target.value })}
              disabled={Number(stockForm.quantity) < 0}
            />
          </div>
        </div>

        <button
          className="btn-primary"
          onClick={handleAdjustStock}
          disabled={stockSaving || !stockForm.quantity}
        >
          {stockSaving ? 'Menyimpan...' : 'Simpan Stok'}
        </button>
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

      {showImport && (
        <ImportModal
          isOpen={importOpen}
          onClose={() => setImportOpen(false)}
          title="Import Stok Produk"
          endpoint={`${resourceBase}/imports/products/stock`}
          projectId={scopeMode === 'query' ? selectedProjectId : undefined}
          templateHint="Kolom: SKU | Qty | HPP per Unit (opsional). SKU harus sudah terdaftar di project ini. HPP per unit boleh dikosongkan kalau hanya ingin koreksi jumlah stok."
          onSuccess={loadProducts}
        />
      )}
    </div>
  );
};

export default ProductPage;
