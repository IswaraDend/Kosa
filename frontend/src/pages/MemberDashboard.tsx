import React from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import './Dashboard.css';

const data = [
  { name: 'Jan', masuk: 4000, keluar: 2400 },
  { name: 'Feb', masuk: 3000, keluar: 1398 },
  { name: 'Mar', masuk: 2000, keluar: 9800 },
  { name: 'Apr', masuk: 2780, keluar: 3908 },
  { name: 'Mei', masuk: 1890, keluar: 4800 },
  { name: 'Jun', masuk: 2390, keluar: 3800 },
];

const MemberDashboard = () => {
  return (
    <div className="dashboard-content">
      <div className="welcome-banner" style={{ background: 'linear-gradient(135deg, #2980b9 0%, #3498db 100%)' }}>
        <h2>Member Dashboard</h2>
        <p>Anda melihat data ini karena Admin Proyek telah memberi Anda izin akses.</p>
      </div>

      <div className="dashboard-grid">
        <div className="card full-width">
          <h3>Arus Pergerakan Stok (Read Only)</h3>
          <div className="chart-container" style={{ height: '300px', marginTop: '20px' }}>
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={data} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#eee" />
                <XAxis dataKey="name" axisLine={false} tickLine={false} />
                <YAxis axisLine={false} tickLine={false} />
                <Tooltip cursor={{fill: '#f5f5f5'}} />
                <Bar dataKey="masuk" fill="var(--primary)" radius={[4, 4, 0, 0]} name="Barang Masuk" />
                <Bar dataKey="keluar" fill="#e74c3c" radius={[4, 4, 0, 0]} name="Barang Keluar" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>
    </div>
  );
};

export default MemberDashboard;
