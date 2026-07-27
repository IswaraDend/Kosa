import React from 'react';
import './Dashboard.css';

const AdminDashboard = () => {
  return (
    <div className="dashboard-content">
      <div className="welcome-banner" style={{ background: 'linear-gradient(135deg, #27ae60 0%, #2ecc71 100%)' }}>
        <h2>Admin Project Area</h2>
        <p>Kelola proyek spesifik, tambahkan member, dan atur hak akses data (MEMBER_PERMISSIONS).</p>
      </div>

      <div className="quick-actions" style={{ display: 'flex', gap: '16px', marginBottom: '24px' }}>
        <button className="btn-primary" style={{ padding: '10px 20px', borderRadius: '8px', cursor: 'pointer', background: '#27ae60', color: 'white', border: 'none' }}>+ Add New Member</button>
      </div>

      <div className="dashboard-grid">
        <div className="card">
          <h3>Member List</h3>
          <p>Daftar pengguna yang memiliki <code>USER_ROLES</code> di proyek ini.</p>
          <div style={{ padding: '20px', background: '#f5f5f5', borderRadius: '8px', marginTop: '16px', color: '#555' }}>
            [ Mockup: List of Members ... ]
          </div>
        </div>

        <div className="card">
          <h3>Access Control</h3>
          <p>Atur <code>MEMBER_PERMISSIONS</code> untuk membatasi data mana yang bisa dilihat oleh Member.</p>
          <div style={{ padding: '20px', background: '#f5f5f5', borderRadius: '8px', marginTop: '16px', color: '#555' }}>
            [ Mockup: Checkbox Hak Akses ... ]
          </div>
        </div>
      </div>
    </div>
  );
};

export default AdminDashboard;
