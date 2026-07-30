import { useEffect, useState } from 'react';
import { Folder, CheckCircle, UserCog, Users, Box } from 'lucide-react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { api, ApiError } from '../../lib/api';
import type { SummaryData, TransactionReportRow } from '../../types';
import PageHeader from '../../components/PageHeader';
import StatCard from '../../components/StatCard';
import '../Dashboard.css';

const RingkasanPage = () => {
  const [summary, setSummary] = useState<SummaryData | null>(null);
  const [movement, setMovement] = useState<TransactionReportRow[]>([]);
  const [errorMsg, setErrorMsg] = useState('');

  useEffect(() => {
    api
      .get<SummaryData>('/super-admin/summary')
      .then(setSummary)
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat ringkasan'));

    api
      .get<{ data: TransactionReportRow[] }>('/super-admin/reports/transactions')
      .then((res) => setMovement(res.data))
      .catch(() => setMovement([]));
  }, []);

  const chartData = Object.values(
    movement.reduce<Record<string, { period: string; masuk: number; keluar: number }>>((acc, row) => {
      acc[row.period] = acc[row.period] || { period: row.period, masuk: 0, keluar: 0 };
      if (row.type === 'in') acc[row.period].masuk += row.total_qty;
      if (row.type === 'out') acc[row.period].keluar += row.total_qty;
      return acc;
    }, {}),
  );

  return (
    <div className="dashboard-content">
      <PageHeader title="Ringkasan" subtitle="Gambaran umum seluruh project, gudang, dan tim" />

      {errorMsg && (
        <div style={{ color: 'var(--danger)', marginBottom: '16px', fontSize: '14px' }}>{errorMsg}</div>
      )}

      <div className="summary-cards">
        <StatCard label="Total Project" value={summary?.total_projects ?? '-'} icon={<Folder size={18} />} />
        <StatCard label="Project Aktif" value={summary?.active_projects ?? '-'} icon={<CheckCircle size={18} />} />
        <StatCard label="Total Gudang" value={summary?.total_warehouses ?? '-'} icon={<Box size={18} />} />
        <StatCard label="Admin Bertugas" value={summary?.total_admins ?? '-'} icon={<UserCog size={18} />} />
        <StatCard label="Total Member" value={summary?.total_members ?? '-'} icon={<Users size={18} />} />
      </div>

      <div className="card full-width chart-card">
        <div className="chart-header">
          <div>
            <div className="chart-title">Arus Pergerakan Stok</div>
            <div className="chart-subtitle">Total barang masuk vs keluar per bulan, seluruh project</div>
          </div>
        </div>
        <div className="chart-container" style={{ height: '300px', marginTop: '20px' }}>
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={chartData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--border-color)" />
              <XAxis dataKey="period" axisLine={false} tickLine={false} stroke="var(--text-muted)" />
              <YAxis axisLine={false} tickLine={false} stroke="var(--text-muted)" />
              <Tooltip
                contentStyle={{ backgroundColor: 'var(--surface-color)', border: 'none', borderRadius: '8px' }}
              />
              <Bar dataKey="masuk" fill="var(--primary)" radius={[4, 4, 0, 0]} name="Barang Masuk" />
              <Bar dataKey="keluar" fill="var(--danger)" radius={[4, 4, 0, 0]} name="Barang Keluar" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  );
};

export default RingkasanPage;
