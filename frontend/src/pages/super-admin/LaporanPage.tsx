import { useEffect, useState } from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { api, ApiError, buildQuery } from '../../lib/api';
import type { StockSummaryRow, TransactionReportRow } from '../../types';
import { useSelectedProject } from '../../hooks/useSelectedProject';
import PageHeader from '../../components/PageHeader';
import TableCard from '../../components/TableCard';
import ProjectPicker from '../../components/ProjectPicker';
import '../Dashboard.css';

interface LaporanPageProps {
  apiBasePrefix?: string;
  scopeMode?: 'query' | 'path';
  projectEndpoint?: string;
  includeAllOption?: boolean;
}

const LaporanPage = ({
  apiBasePrefix = '/super-admin',
  scopeMode = 'query',
  projectEndpoint = '/super-admin/projects',
  includeAllOption = true,
}: LaporanPageProps) => {
  const { selectedProjectId, setSelectedProjectId } = useSelectedProject();
  const [stockRows, setStockRows] = useState<StockSummaryRow[]>([]);
  const [movement, setMovement] = useState<TransactionReportRow[]>([]);
  const [errorMsg, setErrorMsg] = useState('');

  const resourceBase = scopeMode === 'path' ? `${apiBasePrefix}/${selectedProjectId}` : apiBasePrefix;

  useEffect(() => {
    if (scopeMode === 'path' && selectedProjectId === 'all') return;

    const query =
      scopeMode === 'query' ? buildQuery({ project_id: selectedProjectId === 'all' ? undefined : selectedProjectId }) : '';

    api
      .get<{ data: StockSummaryRow[] }>(`${resourceBase}/reports/stock-summary${query}`)
      .then((res) => setStockRows(res.data))
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat laporan stok'));

    api
      .get<{ data: TransactionReportRow[] }>(`${resourceBase}/reports/transactions${query}`)
      .then((res) => setMovement(res.data))
      .catch(() => setMovement([]));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId, apiBasePrefix, scopeMode]);

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
      <PageHeader title="Laporan" subtitle="Ringkasan stok dan pergerakan barang" />

      <div className="form-group" style={{ maxWidth: 280 }}>
        <label>Project</label>
        <ProjectPicker
          value={selectedProjectId}
          onChange={setSelectedProjectId}
          includeAllOption={includeAllOption}
          endpoint={projectEndpoint}
        />
      </div>

      {errorMsg && (
        <div style={{ color: 'var(--danger)', marginBottom: '16px', fontSize: '14px' }}>{errorMsg}</div>
      )}

      <div className="card full-width chart-card" style={{ marginBottom: 24 }}>
        <div className="chart-header">
          <div>
            <div className="chart-title">Tren Pergerakan Stok</div>
            <div className="chart-subtitle">Barang masuk vs keluar per bulan</div>
          </div>
        </div>
        <div className="chart-container" style={{ height: '300px', marginTop: '20px' }}>
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={chartData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--border-color)" />
              <XAxis dataKey="period" axisLine={false} tickLine={false} stroke="var(--text-muted)" />
              <YAxis axisLine={false} tickLine={false} stroke="var(--text-muted)" />
              <Tooltip contentStyle={{ backgroundColor: 'var(--surface-color)', border: 'none', borderRadius: '8px' }} />
              <Legend />
              <Bar dataKey="masuk" fill="var(--primary)" radius={[4, 4, 0, 0]} name="Barang Masuk" />
              <Bar dataKey="keluar" fill="var(--danger)" radius={[4, 4, 0, 0]} name="Barang Keluar" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      <TableCard title="Ringkasan Stok per Gudang" count={stockRows.length}>
        <thead>
          <tr>
            <th>ITEM</th>
            <th>GUDANG</th>
            <th>QTY</th>
            <th>UNIT</th>
          </tr>
        </thead>
        <tbody>
          {stockRows.map((row) => (
            <tr key={`${row.item_id}-${row.warehouse_id}`}>
              <td>{row.item_name}</td>
              <td style={{ color: 'var(--text-muted)' }}>{row.warehouse_name}</td>
              <td>{row.quantity}</td>
              <td style={{ color: 'var(--text-muted)' }}>{row.unit}</td>
            </tr>
          ))}
        </tbody>
      </TableCard>
    </div>
  );
};

export default LaporanPage;
