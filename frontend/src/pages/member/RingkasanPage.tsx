import { useEffect, useState } from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { api, ApiError } from '../../lib/api';
import type { Project, TransactionReportRow } from '../../types';
import { useSelectedProject } from '../../hooks/useSelectedProject';
import { useMemberPermissions } from '../../hooks/useMemberPermissions';
import PageHeader from '../../components/PageHeader';
import ProjectPicker from '../../components/ProjectPicker';
import '../Dashboard.css';

const RingkasanPage = () => {
  const { selectedProjectId, setSelectedProjectId } = useSelectedProject();
  const permissions = useMemberPermissions(true);
  const [project, setProject] = useState<Project | null>(null);
  const [movement, setMovement] = useState<TransactionReportRow[]>([]);
  const [errorMsg, setErrorMsg] = useState('');

  useEffect(() => {
    if (selectedProjectId === 'all' || !selectedProjectId) return;
    api
      .get<Project>(`/member/projects/${selectedProjectId}`)
      .then(setProject)
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat data project'));
  }, [selectedProjectId]);

  useEffect(() => {
    if (selectedProjectId === 'all' || !selectedProjectId || !permissions.includes('report.view')) {
      setMovement([]);
      return;
    }
    api
      .get<{ data: TransactionReportRow[] }>(`/member/projects/${selectedProjectId}/reports/transactions`)
      .then((res) => setMovement(res.data))
      .catch(() => setMovement([]));
  }, [selectedProjectId, permissions]);

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
      <PageHeader title="Ringkasan" subtitle="Anda melihat data ini karena Admin telah memberi Anda akses" />

      <div className="form-group" style={{ maxWidth: 280 }}>
        <label>Project</label>
        <ProjectPicker
          value={selectedProjectId}
          onChange={setSelectedProjectId}
          includeAllOption={false}
          endpoint="/member/projects"
        />
      </div>

      {errorMsg && (
        <div style={{ color: 'var(--danger)', marginBottom: '16px', fontSize: '14px' }}>{errorMsg}</div>
      )}

      {project && (
        <div className="welcome-banner" style={{ background: 'linear-gradient(135deg, #2980b9 0%, #3498db 100%)' }}>
          <h2>{project.name}</h2>
          <p>{project.description || 'Tidak ada deskripsi.'}</p>
        </div>
      )}

      {permissions.includes('report.view') ? (
        <div className="card full-width chart-card">
          <div className="chart-header">
            <div>
              <div className="chart-title">Arus Pergerakan Stok (Read Only)</div>
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
                <Bar dataKey="masuk" fill="var(--primary)" radius={[4, 4, 0, 0]} name="Barang Masuk" />
                <Bar dataKey="keluar" fill="var(--danger)" radius={[4, 4, 0, 0]} name="Barang Keluar" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      ) : (
        <p style={{ color: 'var(--text-muted)' }}>
          Anda belum memiliki akses untuk melihat laporan. Hubungi Admin project Anda.
        </p>
      )}
    </div>
  );
};

export default RingkasanPage;
