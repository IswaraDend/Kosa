import { useEffect, useState } from 'react';
import { Box, Package, Users, BarChart2 } from 'lucide-react';
import { api, ApiError } from '../../lib/api';
import { isActiveModule } from '../../lib/modules';
import { useSelectedProject } from '../../hooks/useSelectedProject';
import { useProjectAutoSelect } from '../../hooks/useProjectAutoSelect';
import PageHeader from '../../components/PageHeader';
import StatCard from '../../components/StatCard';
import '../Dashboard.css';

interface ProjectSummary {
  total_warehouses: number;
  total_items: number;
  total_members: number;
  total_transactions: number;
}

const RingkasanPage = () => {
  const { selectedProjectId } = useSelectedProject();
  useProjectAutoSelect('/admin/projects', true);
  const [summary, setSummary] = useState<ProjectSummary | null>(null);
  const [errorMsg, setErrorMsg] = useState('');

  useEffect(() => {
    if (selectedProjectId === 'all' || !selectedProjectId) {
      setSummary(null);
      return;
    }
    api
      .get<ProjectSummary>(`/admin/projects/${selectedProjectId}/summary`)
      .then(setSummary)
      .catch((err: ApiError) => setErrorMsg(err.message || 'Gagal memuat ringkasan'));
  }, [selectedProjectId]);

  return (
    <div className="dashboard-content">
      <PageHeader title="Ringkasan" subtitle="Gambaran umum project yang Anda kelola" />

      {errorMsg && (
        <div style={{ color: 'var(--danger)', marginBottom: '16px', fontSize: '14px' }}>{errorMsg}</div>
      )}

      {/* The summary endpoint still returns all four counts; the two that
          belong to modules this build does not expose are simply not shown,
          so the page cannot advertise a feature with no screen behind it. */}
      <div className="summary-cards">
        {isActiveModule('warehouse') && (
          <StatCard label="Total Gudang" value={summary?.total_warehouses ?? '-'} icon={<Box size={18} />} />
        )}
        {isActiveModule('item') && (
          <StatCard label="Total Item" value={summary?.total_items ?? '-'} icon={<Package size={18} />} />
        )}
        <StatCard label="Total Member" value={summary?.total_members ?? '-'} icon={<Users size={18} />} />
        {isActiveModule('transaction') && (
          <StatCard label="Total Transaksi" value={summary?.total_transactions ?? '-'} icon={<BarChart2 size={18} />} />
        )}
      </div>
    </div>
  );
};

export default RingkasanPage;
