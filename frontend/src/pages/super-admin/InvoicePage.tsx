import { useEffect, useState } from 'react';
import { Plus, Receipt, Trash2, Printer, Eye } from 'lucide-react';
import { api, ApiError, buildQuery , type PaginatedResponse } from '../../lib/api';
import { formatCurrency } from '../../lib/format';
import type { Customer, InvoiceRecord, InvoiceStatus, Product, Warehouse } from '../../types';
import { useSelectedProject } from '../../hooks/useSelectedProject';
import { useProjectAutoSelect } from '../../hooks/useProjectAutoSelect';
import { usePagination, metaFrom, PER_PAGE } from '../../hooks/usePagination';
import PageHeader from '../../components/PageHeader';
import StatCard from '../../components/StatCard';
import TableCard from '../../components/TableCard';
import Pagination from '../../components/Pagination';
import Badge from '../../components/Badge';
import Modal from '../../components/Modal';
import ProjectPicker from '../../components/ProjectPicker';
import InvoicePrintView from '../../components/InvoicePrintView';
import { makeCan } from '../../lib/permissions';
import '../Dashboard.css';

interface DraftLine {
  product_id: string;
  quantity: string;
  unit_price: string;
}

interface InvoiceFormState {
  customer_id: string;
  warehouse_id: string;
  note: string;
  lines: DraftLine[];
}

const emptyForm: InvoiceFormState = { customer_id: '', warehouse_id: '', note: '', lines: [] };
const emptyDraft: DraftLine = { product_id: '', quantity: '', unit_price: '' };

const statusLabel: Record<InvoiceStatus, string> = { unpaid: 'Belum Lunas', paid: 'Lunas', cancelled: 'Dibatalkan' };
const statusVariant: Record<InvoiceStatus, string> = { unpaid: 'nonaktif', paid: 'aktif', cancelled: 'keluar' };

interface InvoicePageProps {
  apiBasePrefix?: string;
  scopeMode?: 'query' | 'path';
  projectEndpoint?: string;
  includeAllOption?: boolean;
  showProjectPicker?: boolean;
  /** Granted permission codes. Undefined = no per-action gating (Super Admin / Admin). */
  permissions?: string[];
}

const InvoicePage = ({
  apiBasePrefix = '/super-admin',
  scopeMode = 'query',
  projectEndpoint = '/super-admin/projects',
  includeAllOption = true,
  showProjectPicker = true,
  permissions,
}: InvoicePageProps) => {
  const can = makeCan(permissions);
  const { selectedProjectId, setSelectedProjectId } = useSelectedProject();
  useProjectAutoSelect(projectEndpoint, !showProjectPicker);
  const { page, setPage, meta, setMeta } = usePagination(selectedProjectId);
  const [invoices, setInvoices] = useState<InvoiceRecord[]>([]);
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [warehouses, setWarehouses] = useState<Warehouse[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState('');

  const [formOpen, setFormOpen] = useState(false);
  const [form, setForm] = useState<InvoiceFormState>(emptyForm);
  const [draft, setDraft] = useState<DraftLine>(emptyDraft);
  const [saving, setSaving] = useState(false);

  const [viewInvoice, setViewInvoice] = useState<InvoiceRecord | null>(null);
  const [printInvoice, setPrintInvoice] = useState<InvoiceRecord | null>(null);

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


  const loadInvoices = () => {
    if (selectedProjectId === 'all') {
      setInvoices([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    api
      .get<PaginatedResponse<InvoiceRecord>>(`${resourceBase}/invoices${pagedQuery}`)
      .then((res) => {
        setInvoices(res.data);
        setMeta(metaFrom(res));
      })
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat data invoice'))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadInvoices();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId, apiBasePrefix, scopeMode, page]);

  useEffect(() => {
    if (selectedProjectId === 'all') {
      setCustomers([]);
      setWarehouses([]);
      setProducts([]);
      return;
    }
    api
      .get<{ data: Customer[] }>(`${resourceBase}/customers${listQuery}`)
      .then((res) => setCustomers(res.data))
      .catch(() => setCustomers([]));
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
    setDraft(emptyDraft);
    setFormOpen(true);
  };

  const handleDraftProductChange = (productId: string) => {
    const product = products.find((p) => String(p.id) === productId);
    setDraft({ product_id: productId, quantity: draft.quantity, unit_price: product?.default_price ? String(product.default_price) : draft.unit_price });
  };

  const handleAddLine = () => {
    if (!draft.product_id || !draft.quantity || !draft.unit_price) return;
    setForm({ ...form, lines: [...form.lines, draft] });
    setDraft(emptyDraft);
  };

  const handleRemoveLine = (index: number) => {
    setForm({ ...form, lines: form.lines.filter((_, i) => i !== index) });
  };

  const productName = (id: string) => products.find((p) => String(p.id) === id)?.name ?? `Produk #${id}`;
  const productUnit = (id: string) => products.find((p) => String(p.id) === id)?.unit ?? '';
  const productCost = (id: string) => products.find((p) => String(p.id) === id)?.average_cost ?? 0;

  const draftSubtotal = form.lines.reduce((sum, l) => sum + Number(l.quantity) * Number(l.unit_price), 0);
  const draftMargin = form.lines.reduce(
    (sum, l) => sum + Number(l.quantity) * (Number(l.unit_price) - productCost(l.product_id)),
    0,
  );

  const handleSubmit = async () => {
    if (form.lines.length === 0) {
      setErrorMsg('Tambahkan minimal 1 baris produk');
      return;
    }
    setSaving(true);
    setErrorMsg('');
    try {
      await api.post(`${resourceBase}/invoices`, {
        project_id: Number(selectedProjectId),
        customer_id: Number(form.customer_id),
        warehouse_id: Number(form.warehouse_id),
        note: form.note,
        items: form.lines.map((l) => ({
          product_id: Number(l.product_id),
          quantity: Number(l.quantity),
          unit_price: Number(l.unit_price),
        })),
      });
      setFormOpen(false);
      loadInvoices();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal membuat invoice');
    } finally {
      setSaving(false);
    }
  };

  const handleStatusChange = async (invoice: InvoiceRecord, status: InvoiceStatus) => {
    try {
      await api.patch(`${resourceBase}/invoices/${invoice.id}/status`, { status });
      loadInvoices();
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : 'Gagal mengubah status invoice');
    }
  };

  const totalOmzet = invoices.filter((i) => i.status !== 'cancelled').reduce((sum, i) => sum + i.subtotal, 0);
  const totalMargin = invoices
    .filter((i) => i.status !== 'cancelled')
    .reduce((sum, i) => sum + (i.subtotal - i.total_hpp), 0);

  return (
    <div className="dashboard-content">
      <PageHeader
        title="Invoice"
        subtitle="Jual produk ke customer dan kelola dokumen invoice"
        actions={
          can('invoice.create') && (
            <button className="btn-primary" onClick={openCreateForm} disabled={selectedProjectId === 'all'}>
              <Plus size={18} />
              Buat Invoice
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
        <p style={{ color: 'var(--text-muted)' }}>Pilih project terlebih dahulu untuk melihat daftar invoice.</p>
      ) : (
        <>
          <div className="summary-cards" style={{ gridTemplateColumns: 'repeat(3, 1fr)' }}>
            <StatCard label="Total Invoice" value={meta.total} icon={<Receipt size={18} />} />
            <StatCard label="Omzet (halaman ini)" value={formatCurrency(totalOmzet)} icon={<Receipt size={18} />} />
            <StatCard label="Margin (halaman ini)" value={formatCurrency(totalMargin)} icon={<Receipt size={18} />} />
          </div>

          <TableCard title="Riwayat Invoice" count={meta.total}>
            <thead>
              <tr>
                <th>NO. INVOICE</th>
                <th>CUSTOMER</th>
                <th>GUDANG</th>
                <th>SUBTOTAL</th>
                <th>MARGIN</th>
                <th>STATUS</th>
                <th>TANGGAL</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {!loading &&
                invoices.map((inv) => (
                  <tr key={inv.id}>
                    <td>{inv.invoice_number}</td>
                    <td>{inv.customer?.name ?? '-'}</td>
                    <td style={{ color: 'var(--text-muted)' }}>{inv.warehouse?.name ?? '-'}</td>
                    <td>{formatCurrency(inv.subtotal)}</td>
                    <td style={{ color: 'var(--text-muted)' }}>{formatCurrency(inv.subtotal - inv.total_hpp)}</td>
                    <td>
                      {inv.status === 'cancelled' || !can('invoice.update') ? (
                        <Badge label={statusLabel[inv.status]} variant={statusVariant[inv.status]} />
                      ) : (
                        <select
                          className="form-select"
                          style={{ padding: '4px 8px', fontSize: 12 }}
                          value={inv.status}
                          onChange={(e) => handleStatusChange(inv, e.target.value as InvoiceStatus)}
                        >
                          <option value="unpaid">Belum Lunas</option>
                          <option value="paid">Lunas</option>
                          <option value="cancelled">Batalkan</option>
                        </select>
                      )}
                    </td>
                    <td style={{ color: 'var(--text-muted)' }}>
                      {new Date(inv.created_at).toLocaleDateString('id-ID')}
                    </td>
                    <td style={{ textAlign: 'right', display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
                      <button
                        style={{ background: 'transparent', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}
                        onClick={() => setViewInvoice(inv)}
                        title="Lihat"
                      >
                        <Eye size={16} />
                      </button>
                      <button
                        style={{ background: 'transparent', border: 'none', color: 'var(--primary)', cursor: 'pointer' }}
                        onClick={() => setPrintInvoice(inv)}
                        title="Cetak"
                      >
                        <Printer size={16} />
                      </button>
                    </td>
                  </tr>
                ))}
            </tbody>
          </TableCard>

          <Pagination page={page} meta={meta} onChange={setPage} label="invoice" />
        </>
      )}

      <Modal
        isOpen={formOpen}
        onClose={() => setFormOpen(false)}
        title="Buat Invoice"
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
        {customers.length === 0 ? (
          <p style={{ color: 'var(--text-muted)' }}>Belum ada customer, tambah dulu di halaman Pelanggan.</p>
        ) : (
          <>
            <div className="form-group">
              <label>Customer</label>
              <select
                className="form-select"
                value={form.customer_id}
                onChange={(e) => setForm({ ...form, customer_id: e.target.value })}
              >
                <option value="">Pilih customer</option>
                {customers.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
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

            {form.lines.length > 0 && (
              <div className="form-group">
                <label>Baris Produk</label>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 6, marginBottom: 8 }}>
                  {form.lines.map((line, i) => (
                    <div
                      key={i}
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                        fontSize: 13,
                        padding: '6px 10px',
                        background: 'var(--bg-color)',
                        border: '1px solid var(--border-color)',
                        borderRadius: 8,
                      }}
                    >
                      <span>
                        {productName(line.product_id)} — {line.quantity} {productUnit(line.product_id)} &times;{' '}
                        {formatCurrency(Number(line.unit_price))}
                      </span>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                        <strong>{formatCurrency(Number(line.quantity) * Number(line.unit_price))}</strong>
                        <button
                          style={{ background: 'transparent', border: 'none', color: 'var(--danger)', cursor: 'pointer' }}
                          onClick={() => handleRemoveLine(i)}
                        >
                          <Trash2 size={14} />
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    fontSize: 13,
                    paddingTop: 8,
                    borderTop: '1px solid var(--border-color)',
                  }}
                >
                  <span style={{ color: 'var(--text-muted)' }}>Subtotal / Estimasi Margin</span>
                  <strong>
                    {formatCurrency(draftSubtotal)} / {formatCurrency(draftMargin)}
                  </strong>
                </div>
              </div>
            )}

            <div style={{ display: 'flex', gap: 8, alignItems: 'flex-end' }}>
              <div className="form-group" style={{ flex: 1, marginBottom: 0 }}>
                <label>Produk</label>
                <select
                  className="form-select"
                  value={draft.product_id}
                  onChange={(e) => handleDraftProductChange(e.target.value)}
                >
                  <option value="">Pilih produk</option>
                  {products.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name} ({p.unit})
                    </option>
                  ))}
                </select>
              </div>
              <div className="form-group" style={{ width: 80, marginBottom: 0 }}>
                <label>Qty</label>
                <input
                  type="number"
                  className="form-input"
                  value={draft.quantity}
                  onChange={(e) => setDraft({ ...draft, quantity: e.target.value })}
                />
              </div>
              <div className="form-group" style={{ width: 120, marginBottom: 0 }}>
                <label>Harga Jual</label>
                <input
                  type="number"
                  className="form-input"
                  value={draft.unit_price}
                  onChange={(e) => setDraft({ ...draft, unit_price: e.target.value })}
                />
              </div>
              <button className="btn-primary" onClick={handleAddLine}>
                <Plus size={16} />
              </button>
            </div>

            <div className="form-group" style={{ marginTop: 16 }}>
              <label>Catatan</label>
              <textarea
                className="form-textarea"
                rows={2}
                value={form.note}
                onChange={(e) => setForm({ ...form, note: e.target.value })}
              />
            </div>
          </>
        )}
      </Modal>

      <Modal
        isOpen={!!viewInvoice}
        onClose={() => setViewInvoice(null)}
        title={`Invoice ${viewInvoice?.invoice_number ?? ''}`}
        footer={
          <button className="btn-secondary" onClick={() => setViewInvoice(null)}>
            Tutup
          </button>
        }
      >
        {viewInvoice && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6, fontSize: 13 }}>
            <div>Customer: {viewInvoice.customer?.name ?? '-'}</div>
            <div>Gudang: {viewInvoice.warehouse?.name ?? '-'}</div>
            <div>Status: {statusLabel[viewInvoice.status]}</div>
            <div style={{ marginTop: 8, borderTop: '1px solid var(--border-color)', paddingTop: 8 }}>
              {viewInvoice.items.map((line) => (
                <div key={line.id} style={{ display: 'flex', justifyContent: 'space-between', padding: '4px 0' }}>
                  <span>
                    {line.product?.name ?? `Produk #${line.product_id}`} &times; {line.quantity}
                  </span>
                  <span>{formatCurrency(line.unit_price * line.quantity)}</span>
                </div>
              ))}
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontWeight: 600, marginTop: 8 }}>
              <span>Subtotal</span>
              <span>{formatCurrency(viewInvoice.subtotal)}</span>
            </div>
          </div>
        )}
      </Modal>

      {printInvoice && (
        <InvoicePrintView invoice={printInvoice} onClose={() => setPrintInvoice(null)} />
      )}
    </div>
  );
};

export default InvoicePage;
