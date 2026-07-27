import React from 'react';
import { Search, Bell, Plus, Folder, CheckCircle, Users, UserCog, MoreVertical } from 'lucide-react';
import './Dashboard.css';

const SuperAdminDashboard = () => {
  const projects = [
    { name: 'Gudang Jaksel', code: 'PRJ-001', admin: 'Budi Santoso', member: 8, gudang: 2, status: 'Aktif', created: '12 Jan 2026' },
    { name: 'Proyek Bandung', code: 'PRJ-002', admin: 'Rina Amelia', member: 5, gudang: 1, status: 'Aktif', created: '03 Feb 2026' },
    { name: 'Cabang Surabaya', code: 'PRJ-003', admin: 'Andi Pratama', member: 6, gudang: 3, status: 'Aktif', created: '20 Mar 2026' },
    { name: 'Gudang Bekasi', code: 'PRJ-004', admin: '— Belum ditentukan', member: 0, gudang: 1, status: 'Nonaktif', created: '02 Jun 2026' },
  ];

  return (
    <div className="dashboard-content">
      
      {/* Header Area */}
      <div className="page-header">
        <div className="page-title">
          <h1>Kelola Project</h1>
          <p>Buat project baru dan tentukan admin penanggung jawabnya</p>
        </div>
        
        <div className="header-actions">
          <div className="search-box">
            <Search size={16} />
            <input type="text" placeholder="Cari project..." />
          </div>
          <button className="btn-icon">
            <Bell size={18} />
          </button>
          <button className="btn-primary">
            <Plus size={18} />
            Tambah Project
          </button>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="summary-cards">
        <div className="stat-card">
          <div className="stat-header">
            <span>Total Project</span>
            <Folder size={18} className="stat-icon" />
          </div>
          <div className="stat-value">4</div>
        </div>

        <div className="stat-card">
          <div className="stat-header">
            <span>Project Aktif</span>
            <CheckCircle size={18} className="stat-icon" />
          </div>
          <div className="stat-value">3</div>
        </div>

        <div className="stat-card">
          <div className="stat-header">
            <span>Admin Bertugas</span>
            <UserCog size={18} className="stat-icon" />
          </div>
          <div className="stat-value">3</div>
        </div>

        <div className="stat-card">
          <div className="stat-header">
            <span>Total Member</span>
            <Users size={18} className="stat-icon" />
          </div>
          <div className="stat-value">19</div>
        </div>
      </div>

      {/* Table Area */}
      <div className="table-card">
        <div className="table-header">
          <h3>Daftar Project (4)</h3>
        </div>
        <table>
          <thead>
            <tr>
              <th>NAMA PROJECT</th>
              <th>KODE</th>
              <th>ADMIN</th>
              <th>MEMBER</th>
              <th>GUDANG</th>
              <th>STATUS</th>
              <th>DIBUAT</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {projects.map((project, index) => (
              <tr key={index}>
                <td>{project.name}</td>
                <td style={{ color: 'var(--text-muted)' }}>{project.code}</td>
                <td>{project.admin}</td>
                <td>{project.member}</td>
                <td>{project.gudang}</td>
                <td>
                  <span className={`badge ${project.status.toLowerCase()}`}>
                    {project.status}
                  </span>
                </td>
                <td style={{ color: 'var(--text-muted)' }}>{project.created}</td>
                <td style={{ textAlign: 'right' }}>
                  <button style={{ background: 'transparent', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
                    <MoreVertical size={16} />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

    </div>
  );
};

export default SuperAdminDashboard;
