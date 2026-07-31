import { useEffect, useState } from 'react';
import { Box } from 'lucide-react';
import { api, ApiError } from '../../lib/api';
import type { Warehouse } from '../../types';
import { useSelectedProject } from '../../hooks/useSelectedProject';
import PageHeader from '../../components/PageHeader';
import StatCard from '../../components/StatCard';
import TableCard from '../../components/TableCard';
import ProjectPicker from '../../components/ProjectPicker';
import '../Dashboard.css';

const GudangPage = () => {
  const { selectedProjectId, setSelectedProjectId } = useSelectedProject();
  const [warehouses, setWarehouses] = useState<Warehouse[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState('');

  useEffect(() => {
    if (selectedProjectId === 'all' || !selectedProjectId) {
      setWarehouses([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    api
      .get<{ data: Warehouse[] }>(`/member/projects/${selectedProjectId}/warehouses`)
      .then((res) => setWarehouses(res.data))
      .catch((err: ApiError) => setErrorMsg(err.message || 'Anda belum memiliki akses gudang di project ini'))
      .finally(() => setLoading(false));
  }, [selectedProjectId]);

  return (
    <div className="dashboard-content">
      <PageHeader title="Gudang" subtitle="Daftar gudang (read only)" />

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

      <div className="summary-cards" style={{ gridTemplateColumns: 'repeat(1, 1fr)' }}>
        <StatCard label="Total Gudang" value={warehouses.length} icon={<Box size={18} />} />
      </div>

      <TableCard title="Daftar Gudang" count={warehouses.length}>
        <thead>
          <tr>
            <th>NAMA GUDANG</th>
            <th>KODE</th>
            <th>ALAMAT</th>
          </tr>
        </thead>
        <tbody>
          {!loading &&
            warehouses.map((w) => (
              <tr key={w.id}>
                <td>{w.name}</td>
                <td style={{ color: 'var(--text-muted)' }}>{w.code}</td>
                <td style={{ color: 'var(--text-muted)' }}>{w.address}</td>
              </tr>
            ))}
        </tbody>
      </TableCard>
    </div>
  );
};

export default GudangPage;
