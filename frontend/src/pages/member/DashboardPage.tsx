import { useEffect, useState } from 'react';
import { Box, ArrowDownCircle, ArrowUpCircle, ArrowRight } from 'lucide-react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { api, ApiError } from '../../lib/api';
import type { Project, TransactionRecord, TransactionReportRow, Warehouse } from '../../types';
import { useSelectedProject } from '../../hooks/useSelectedProject';
import { useProjectAutoSelect } from '../../hooks/useProjectAutoSelect';
import { useMemberPermissions } from '../../hooks/useMemberPermissions';
import { useAuth } from '../../hooks/useAuth';
import StatCard from '../../components/StatCard';
import TableCard from '../../components/TableCard';
import '../Dashboard.css';

const DashboardPage = () => {
  const { selectedProjectId } = useSelectedProject();
  useProjectAutoSelect('/member/projects', true);
  const { user } = useAuth();
  const permissions = useMemberPermissions(true);

  const [project, setProject] = useState<Project | null>(null);
  const [warehouses, setWarehouses] = useState<Warehouse[]>([]);
  const [transactions, setTransactions] = useState<TransactionRecord[]>([]);
  const [movement, setMovement] = useState<TransactionReportRow[]>([]);
  const [errorMsg, setErrorMsg] = useState('');

  const canViewWarehouse = permissions.includes('warehouse.view');
  const canViewTransaction = permissions.includes('transaction.view');
  const canViewReport = permissions.includes('report.view');
  const hasAnyAccess = canViewWarehouse || canViewTransaction || canViewReport;

  useEffect(() => {
    if (selectedProjectId === 'all' || !selectedProjectId) return;
    api
      .get<Project>(`/member/projects/${selectedProjectId}`)
      .then(setProject)
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat data project'));
  }, [selectedProjectId]);

  useEffect(() => {
    if (selectedProjectId === 'all' || !selectedProjectId) return;

    if (canViewWarehouse) {
      api
        .get<{ data: Warehouse[] }>(`/member/projects/${selectedProjectId}/warehouses`)
        .then((res) => setWarehouses(res.data))
        .catch(() => setWarehouses([]));
    } else {
      setWarehouses([]);
    }

    if (canViewTransaction) {
      api
        .get<{ data: TransactionRecord[] }>(`/member/projects/${selectedProjectId}/transactions`)
        .then((res) => setTransactions(res.data))
        .catch(() => setTransactions([]));
    } else {
      setTransactions([]);
    }

    if (canViewReport) {
      api
        .get<{ data: TransactionReportRow[] }>(`/member/projects/${selectedProjectId}/reports/transactions`)
        .then((res) => setMovement(res.data))
        .catch(() => setMovement([]));
    } else {
      setMovement([]);
    }
  }, [selectedProjectId, canViewWarehouse, canViewTransaction, canViewReport]);

  const chartData = Object.values(
    movement.reduce<Record<string, { period: string; masuk: number; keluar: number }>>((acc, row) => {
      acc[row.period] = acc[row.period] || { period: row.period, masuk: 0, keluar: 0 };
      if (row.type === 'in') acc[row.period].masuk += row.total_qty;
      if (row.type === 'out') acc[row.period].keluar += row.total_qty;
      return acc;
    }, {}),
  );

  const lineQty = (t: TransactionRecord) => t.items.reduce((sum, i) => sum + i.quantity, 0);
  const totalMasuk = transactions.filter((t) => t.type === 'in').reduce((sum, t) => sum + lineQty(t), 0);
  const totalKeluar = transactions.filter((t) => t.type === 'out').reduce((sum, t) => sum + lineQty(t), 0);
  const recentActivity = transactions.slice(0, 5);

  const firstName = (user?.name || '').split(' ')[0];

  return (
    <div className="dashboard-content">
      <div className="page-header">
        <div className="page-title">
          <div className="member-greeting-eyebrow">
            <span className="member-pulse" /> Live &middot; data terkini
          </div>
          <h1>Halo{firstName ? `, ${firstName}` : ''} 👋</h1>
          <p>
            Ini pergerakan stok di {project ? project.name : 'project kamu'} yang bisa kamu pantau hari ini.
          </p>
        </div>
      </div>

      {errorMsg && (
        <div style={{ color: 'var(--danger)', marginBottom: '16px', fontSize: '14px' }}>{errorMsg}</div>
      )}

      {!hasAnyAccess ? (
        <p style={{ color: 'var(--text-muted)' }}>
          Anda belum memiliki akses ke data operasional. Hubungi Admin project Anda.
        </p>
      ) : (
        <>
          <div className="summary-cards">
            {canViewWarehouse && (
              <StatCard label="Total Gudang" value={warehouses.length} icon={<Box size={18} />} />
            )}
            {canViewTransaction && (
              <>
                <StatCard label="Barang Masuk" value={totalMasuk} icon={<ArrowDownCircle size={18} />} />
                <StatCard label="Barang Keluar" value={totalKeluar} icon={<ArrowUpCircle size={18} />} />
              </>
            )}
          </div>

          {canViewReport && (
            <div className="card full-width chart-card" style={{ marginBottom: 24 }}>
              <div className="chart-header">
                <div>
                  <div className="chart-title">Arus Stok</div>
                  <div className="chart-subtitle">Masuk vs keluar per bulan</div>
                </div>
                <div className="chart-legend">
                  <span className="chart-legend-dot" style={{ background: 'var(--success)' }} /> Masuk
                  <span className="chart-legend-dot" style={{ background: 'var(--primary)', marginLeft: 12 }} /> Keluar
                </div>
              </div>
              <div className="chart-container" style={{ height: '260px', marginTop: '20px' }}>
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={chartData} margin={{ top: 5, right: 20, left: 10, bottom: 5 }}>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--border-color)" />
                    <XAxis dataKey="period" axisLine={false} tickLine={false} stroke="var(--text-muted)" />
                    <YAxis axisLine={false} tickLine={false} stroke="var(--text-muted)" />
                    <Tooltip contentStyle={{ backgroundColor: 'var(--surface-color-light)', border: 'none', borderRadius: '8px' }} />
                    <Legend />
                    <Bar dataKey="masuk" fill="var(--success)" radius={[4, 4, 0, 0]} name="Barang Masuk" />
                    <Bar dataKey="keluar" fill="var(--primary)" radius={[4, 4, 0, 0]} name="Barang Keluar" />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>
          )}

          <div className="dashboard-grid">
            {canViewTransaction && (
              <TableCard title="Aktivitas Terbaru" count={recentActivity.length}>
                <tbody>
                  {recentActivity.length === 0 && (
                    <tr>
                      <td style={{ color: 'var(--text-muted)' }}>Belum ada transaksi.</td>
                    </tr>
                  )}
                  {recentActivity.map((t) => (
                    <tr key={t.id} style={{ borderBottom: 'none' }}>
                      <td colSpan={1} style={{ padding: 0, border: 'none' }}>
                        <div className="member-feed-row">
                          <div className={`member-feed-icon ${t.type === 'in' ? 'in' : 'out'}`}>
                            {t.type === 'in' ? <ArrowDownCircle size={15} /> : <ArrowUpCircle size={15} />}
                          </div>
                          <div>
                            <div className="member-feed-title">
                              {t.items.map((i) => i.item?.name).filter(Boolean).join(', ') || 'Item'}{' '}
                              {t.type === 'in' ? 'masuk' : t.type === 'out' ? 'keluar' : 'transfer'}
                              {t.source_warehouse || t.dest_warehouse
                                ? ` — ${(t.dest_warehouse || t.source_warehouse)?.name}`
                                : ''}
                            </div>
                            <div className="member-feed-meta">{new Date(t.created_at).toLocaleString('id-ID')}</div>
                          </div>
                          <div className={`member-feed-qty ${t.type === 'in' ? 'in' : 'out'}`}>
                            {t.type === 'in' ? '+' : t.type === 'out' ? '−' : ''}
                            {lineQty(t)}
                          </div>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </TableCard>
            )}

            {canViewWarehouse && (
              <TableCard title="Gudang" count={warehouses.length}>
                <thead>
                  <tr>
                    <th>NAMA</th>
                    <th>KODE</th>
                    <th>ALAMAT</th>
                  </tr>
                </thead>
                <tbody>
                  {warehouses.map((w) => (
                    <tr key={w.id}>
                      <td>{w.name}</td>
                      <td style={{ color: 'var(--text-muted)' }}>{w.code}</td>
                      <td style={{ color: 'var(--text-muted)' }}>{w.address}</td>
                    </tr>
                  ))}
                </tbody>
              </TableCard>
            )}
          </div>

          <div style={{ fontSize: 11.5, color: 'var(--text-muted)', display: 'flex', alignItems: 'center', gap: 6 }}>
            <ArrowRight size={13} style={{ opacity: 0.6 }} />
            Menu yang tidak terlihat berarti akses belum diberikan oleh Admin project ini.
          </div>
        </>
      )}
    </div>
  );
};

export default DashboardPage;
