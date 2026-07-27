import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { LayoutDashboard, Users, Box, Lock, BarChart2, FileText, Radio, Folder } from 'lucide-react';
import './DashboardLayout.css';

const DashboardLayout = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const userStr = localStorage.getItem('user');
  const user = userStr ? JSON.parse(userStr) : null;

  let menus = [];

  if (user?.is_super_admin) {
    menus = [
      { path: '/super-admin/ringkasan', label: 'Ringkasan', icon: <LayoutDashboard size={18} /> },
      { path: '/super-admin', label: 'Kelola Project', icon: <Folder size={18} /> },
      { path: '/super-admin/users', label: 'Admin & Member', icon: <Users size={18} /> },
      { path: '/super-admin/gudang', label: 'Gudang', icon: <Box size={18} /> },
      { path: '/super-admin/permission', label: 'Permission', icon: <Lock size={18} /> },
      { path: '/super-admin/transaksi', label: 'Transaksi', icon: <BarChart2 size={18} /> },
      { path: '/super-admin/laporan', label: 'Laporan', icon: <FileText size={18} /> },
    ];
  } else if (location.pathname.startsWith('/admin')) {
    menus = [
      { path: '/admin', label: 'My Project', icon: <LayoutDashboard size={18} /> },
      { path: '/admin/members', label: 'Members', icon: <Users size={18} /> },
    ];
  } else {
    menus = [
      { path: '/member', label: 'Ringkasan', icon: <LayoutDashboard size={18} /> },
    ];
  }

  return (
    <div className="dashboard-container">
      <aside className="sidebar">
        <div className="sidebar-header">
          <div className="logo-icon">
            <Radio size={20} />
          </div>
          <h2>Stockpulse</h2>
        </div>
        
        <nav className="sidebar-nav">
          {menus.map((menu) => (
            <button
              key={menu.path}
              className={`nav-item ${location.pathname === menu.path ? 'active' : ''}`}
              onClick={() => navigate(menu.path)}
            >
              {menu.icon}
              <span>{menu.label}</span>
            </button>
          ))}
        </nav>

        <div className="sidebar-footer">
          <div className="footer-label">Login sebagai</div>
          <div className="footer-role">{user?.is_super_admin ? 'SUPER ADMIN' : 'ADMIN'}</div>
        </div>
      </aside>

      <main className="main-content">
        <div className="content-area">
          <Outlet />
        </div>
      </main>
    </div>
  );
};

export default DashboardLayout;
