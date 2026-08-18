import { useEffect, useState } from 'react';
import { Folder, Box, Receipt, TrendingUp, Wallet } from 'lucide-react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { api, ApiError } from '../../lib/api';
import { formatCurrency } from '../../lib/format';
import type {
  InvoiceRecord,
  SalesSummaryRow,
  StockSummaryRow,
  SummaryData,
  TopProductRow,
  TransactionReportRow,
} from '../../types';
import PageHeader from '../../components/PageHeader';
import StatCard from '../../components/StatCard';
import TableCard from '../../components/TableCard';
import Badge from '../../components/Badge';
import '../Dashboard.css';

const statusLabel: Record<string, string> = { unpaid: 'Belum Lunas', paid: 'Lunas', cancelled: 'Dibatalkan' };
const statusVariant: Record<string, string> = { unpaid: 'nonaktif', paid: 'aktif', cancelled: 'keluar' };

const DashboardPage = () => {
  const [summary, setSummary] = useState<SummaryData | null>(null);
  const [sales, setSales] = useState<SalesSummaryRow[]>([]);
  const [movement, setMovement] = useState<TransactionReportRow[]>([]);
  const [topProducts, setTopProducts] = useState<TopProductRow[]>([]);
  const [lowStock, setLowStock] = useState<StockSummaryRow[]>([]);
  const [recentInvoices, setRecentInvoices] = useState<InvoiceRecord[]>([]);
  const [errorMsg, setErrorMsg] = useState('');

  useEffect(() => {
    api
      .get<SummaryData>('/super-admin/summary')
      .then(setSummary)
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat ringkasan'));

    api
      .get<{ data: SalesSummaryRow[] }>('/super-admin/reports/sales-summary')
      .then((res) => setSales(res.data))
      .catch(() => setSales([]));

    api
      .get<{ data: TransactionReportRow[] }>('/super-admin/reports/transactions')
      .then((res) => setMovement(res.data))
      .catch(() => setMovement([]));

    api
      .get<{ data: TopProductRow[] }>('/super-admin/reports/top-products?limit=5')
      .then((res) => setTopProducts(res.data))
      .catch(() => setTopProducts([]));

    api
      .get<{ data: StockSummaryRow[] }>('/super-admin/reports/stock-summary')
      .then((res) => setLowStock([...res.data].sort((a, b) => a.quantity - b.quantity).slice(0, 5)))
      .catch(() => setLowStock([]));

    api
      .get<{ data: InvoiceRecord[] }>('/super-admin/invoices?limit=5')
      .then((res) => setRecentInvoices(res.data))
      .catch(() => setRecentInvoices([]));
  }, []);

  const movementChartData = Object.values(
    movement.reduce<Record<string, { period: string; masuk: number; keluar: number }>>((acc, row) => {
      acc[row.period] = acc[row.period] || { period: row.period, masuk: 0, keluar: 0 };
      if (row.type === 'in') acc[row.period].masuk += row.total_qty;
      if (row.type === 'out') acc[row.period].keluar += row.total_qty;
      return acc;
    }, {}),
  );

  return (
    <div className="dashboard-content">
      <PageHeader title="Dashboard" subtitle="Ringkasan bisnis lintas semua project" />

      {errorMsg && (
        <div style={{ color: 'var(--danger)', marginBottom: '16px', fontSize: '14px' }}>{errorMsg}</div>
      )}

      <div className="summary-cards">
        <StatCard label="Total Omzet" value={formatCurrency(summary?.total_revenue ?? 0)} icon={<Wallet size={18} />} />
        <StatCard label="Total Margin" value={formatCurrency(summary?.total_margin ?? 0)} icon={<TrendingUp size={18} />} />
        <StatCard label="Total Invoice" value={summary?.total_invoices ?? '-'} icon={<Receipt size={18} />} />
        <StatCard label="Project Aktif" value={summary?.active_projects ?? '-'} icon={<Folder size={18} />} />
        <StatCard label="Total Gudang" value={summary?.total_warehouses ?? '-'} icon={<Box size={18} />} />
      </div>

      <div className="dashboard-grid">
        <div className="card chart-card">
          <div className="chart-header">
            <div>
              <div className="chart-title">Tren Omzet &amp; Margin</div>
              <div className="chart-subtitle">Penjualan produk per bulan, seluruh project</div>
            </div>
          </div>
          <div className="chart-container" style={{ height: '260px', marginTop: '20px' }}>
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={sales} margin={{ top: 5, right: 20, left: 10, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--border-color)" />
                <XAxis dataKey="period" axisLine={false} tickLine={false} stroke="var(--text-muted)" />
                <YAxis axisLine={false} tickLine={false} stroke="var(--text-muted)" />
                <Tooltip
                  formatter={(value) => formatCurrency(Number(value))}
                  contentStyle={{ backgroundColor: 'var(--surface-color)', border: 'none', borderRadius: '8px' }}
                />
                <Legend />
                <Bar dataKey="subtotal" fill="var(--primary)" radius={[4, 4, 0, 0]} name="Omzet" />
                <Bar dataKey="total_hpp" fill="var(--danger)" radius={[4, 4, 0, 0]} name="HPP" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div className="card chart-card">
          <div className="chart-header">
            <div>
              <div className="chart-title">Arus Pergerakan Stok</div>
              <div className="chart-subtitle">Barang masuk vs keluar per bulan</div>
            </div>
          </div>
          <div className="chart-container" style={{ height: '260px', marginTop: '20px' }}>
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={movementChartData} margin={{ top: 5, right: 20, left: 10, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--border-color)" />
                <XAxis dataKey="period" axisLine={false} tickLine={false} stroke="var(--text-muted)" />
                <YAxis axisLine={false} tickLine={false} stroke="var(--text-muted)" />
                <Tooltip contentStyle={{ backgroundColor: 'var(--surface-color)', border: 'none', borderRadius: '8px' }} />
                <Bar dataKey="masuk" fill="var(--success)" radius={[4, 4, 0, 0]} name="Masuk" />
                <Bar dataKey="keluar" fill="var(--danger)" radius={[4, 4, 0, 0]} name="Keluar" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      <div className="dashboard-grid">
        <TableCard title="Produk Terlaris" count={topProducts.length}>
          <thead>
            <tr>
              <th>PRODUK</th>
              <th>QTY TERJUAL</th>
              <th>OMZET</th>
            </tr>
          </thead>
          <tbody>
            {topProducts.length === 0 && (
              <tr>
                <td colSpan={3} style={{ color: 'var(--text-muted)' }}>
                  Belum ada penjualan.
                </td>
              </tr>
            )}
            {topProducts.map((p) => (
              <tr key={p.product_id}>
                <td>{p.product_name}</td>
                <td>{p.total_qty_sold}</td>
                <td>{formatCurrency(p.total_revenue)}</td>
              </tr>
            ))}
          </tbody>
        </TableCard>

        <TableCard title="Stok Terendah" count={lowStock.length}>
          <thead>
            <tr>
              <th>ITEM</th>
              <th>GUDANG</th>
              <th>SISA QTY</th>
            </tr>
          </thead>
          <tbody>
            {lowStock.length === 0 && (
              <tr>
                <td colSpan={3} style={{ color: 'var(--text-muted)' }}>
                  Belum ada data stok.
                </td>
              </tr>
            )}
            {lowStock.map((s) => (
              <tr key={`${s.item_id}-${s.warehouse_id}`}>
                <td>{s.item_name}</td>
                <td style={{ color: 'var(--text-muted)' }}>{s.warehouse_name}</td>
                <td>
                  {s.quantity} {s.unit}
                </td>
              </tr>
            ))}
          </tbody>
        </TableCard>
      </div>

      <TableCard title="Invoice Terbaru" count={recentInvoices.length}>
        <thead>
          <tr>
            <th>NO. INVOICE</th>
            <th>CUSTOMER</th>
            <th>SUBTOTAL</th>
            <th>STATUS</th>
            <th>TANGGAL</th>
          </tr>
        </thead>
        <tbody>
          {recentInvoices.length === 0 && (
            <tr>
              <td colSpan={5} style={{ color: 'var(--text-muted)' }}>
                Belum ada invoice.
              </td>
            </tr>
          )}
          {recentInvoices.map((inv) => (
            <tr key={inv.id}>
              <td>{inv.invoice_number}</td>
              <td>{inv.customer?.name ?? '-'}</td>
              <td>{formatCurrency(inv.subtotal)}</td>
              <td>
                <Badge label={statusLabel[inv.status]} variant={statusVariant[inv.status]} />
              </td>
              <td style={{ color: 'var(--text-muted)' }}>{new Date(inv.created_at).toLocaleDateString('id-ID')}</td>
            </tr>
          ))}
        </tbody>
      </TableCard>
    </div>
  );
};

export default DashboardPage;
