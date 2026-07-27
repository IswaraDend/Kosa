import { 
  Box, 
  TrendingUp, 
  TrendingDown, 
  AlertTriangle, 
  LayoutGrid, 
  ArrowRightLeft 
} from 'lucide-react';
import { 
  AreaChart, 
  Area, 
  XAxis, 
  Tooltip, 
  ResponsiveContainer 
} from 'recharts';
import './Dashboard.css';

interface DashboardProps {
  role: 'admin' | 'super-admin';
}

const data = [
  { time: '04:00', volume: 10 },
  { time: '08:00', volume: 20 },
  { time: '12:00', volume: 55 },
  { time: '16:00', volume: 40 },
  { time: '20:00', volume: 45 },
  { time: '24:00', volume: 25 },
];

const criticalStocks = [
  { name: 'Semen Portland 40kg', loc: 'Gudang Jaksel • MAT-0012', current: 8, max: 25, color: '#f1c40f' },
  { name: 'Kabel NYM 2x1.5', loc: 'Proyek Bandung • ELC-0044', current: 14, max: 30, color: '#a76cf7' },
  { name: 'Cat Tembok Putih 5L', loc: 'Cabang Surabaya • FIN-0031', current: 3, max: 20, color: '#f1c40f' },
  { name: 'Besi Beton 10mm', loc: 'Gudang Jaksel • MAT-0077', current: 21, max: 50, color: '#a76cf7' },
];

const Dashboard = ({ role }: DashboardProps) => {
  return (
    <div className="dashboard-content">
      {/* Super Admin Notice */}
      {role === 'super-admin' && (
        <div style={{ padding: '12px 16px', background: 'rgba(167, 108, 247, 0.1)', border: '1px solid var(--brand-primary)', borderRadius: '8px', color: 'var(--brand-primary)', fontSize: '14px', display: 'flex', alignItems: 'center', gap: '8px' }}>
          <AlertTriangle size={16} />
          Anda login sebagai <strong>Super Admin</strong>. Menampilkan data akumulasi dari seluruh project.
        </div>
      )}

      {/* Metrics Row */}
      <div className="stats-grid">
        <div className="card stat-card">
          <div className="stat-header">
            Nilai Stok Total
            <Box size={16} />
          </div>
          <div>
            <div className="stat-value">Rp {role === 'super-admin' ? '842,3jt' : '210,5jt'}</div>
            <div className="stat-trend up">
              <TrendingUp size={14} /> +4.2% minggu ini
            </div>
          </div>
        </div>

        <div className="card stat-card">
          <div className="stat-header">
            Item Stok Kritis
            <AlertTriangle size={16} color="var(--warning)" />
          </div>
          <div>
            <div className="stat-value">{role === 'super-admin' ? '17' : '4'}</div>
            <div className="stat-trend neutral">
              <TrendingDown size={14} /> +3 minggu ini
            </div>
          </div>
        </div>

        <div className="card stat-card">
          <div className="stat-header">
            Project Aktif
            <LayoutGrid size={16} />
          </div>
          <div>
            <div className="stat-value">{role === 'super-admin' ? '6' : '1'}</div>
            <div className="stat-trend up">
              <TrendingUp size={14} /> +1 minggu ini
            </div>
          </div>
        </div>

        <div className="card stat-card">
          <div className="stat-header">
            Transaksi Hari Ini
            <ArrowRightLeft size={16} />
          </div>
          <div>
            <div className="stat-value">{role === 'super-admin' ? '128' : '32'}</div>
            <div className="stat-trend down">
              <TrendingDown size={14} /> -6% minggu ini
            </div>
          </div>
        </div>
      </div>

      {/* Charts Row */}
      <div className="charts-grid">
        {/* Area Chart */}
        <div className="card chart-card">
          <div className="chart-header">
            <div>
              <div className="chart-title">Arus Pergerakan Stok</div>
              <div className="chart-subtitle">Unit keluar/masuk — 24 jam terakhir, {role === 'super-admin' ? 'semua gudang' : 'gudang cabang'}</div>
            </div>
            <div className="chart-legend">
              <div className="chart-legend-dot"></div>
              Volume transaksi
            </div>
          </div>
          
          <div style={{ flex: 1, width: '100%', minHeight: '220px' }}>
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={data} margin={{ top: 10, right: 0, left: -20, bottom: 0 }}>
                <defs>
                  <linearGradient id="colorVolume" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#e5b370" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="#e5b370" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <XAxis dataKey="time" stroke="var(--text-muted)" fontSize={12} tickLine={false} axisLine={false} />
                <Tooltip 
                  contentStyle={{ backgroundColor: 'var(--bg-input)', border: 'none', borderRadius: '8px', color: 'var(--text-primary)' }}
                  itemStyle={{ color: 'var(--brand-secondary)' }}
                />
                <Area type="monotone" dataKey="volume" stroke="#e5b370" strokeWidth={2} fillOpacity={1} fill="url(#colorVolume)" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Critical Stocks List */}
        <div className="card chart-card">
          <div className="chart-header" style={{ marginBottom: '16px' }}>
            <div className="chart-title" style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <AlertTriangle size={16} color="var(--warning)" />
              Stok Kritis
            </div>
          </div>
          
          <div className="critical-list">
            {criticalStocks.map((stock, i) => (
              <div className="critical-item" key={i}>
                <div>
                  <div className="critical-item-header">
                    <span>{stock.name}</span>
                    <span className="critical-item-stock">{stock.current}/{stock.max}</span>
                  </div>
                  <div className="critical-item-subtitle">{stock.loc}</div>
                </div>
                <div className="progress-bar-bg">
                  <div 
                    className="progress-bar-fill" 
                    style={{ 
                      width: `${(stock.current / stock.max) * 100}%`,
                      backgroundColor: stock.color
                    }}
                  ></div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
