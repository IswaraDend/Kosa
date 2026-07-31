import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { LayoutDashboard, Users, Box, Package, Lock, BarChart2, FileText, Radio, Folder, LogOut } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';
import { useMemberPermissions } from '../hooks/useMemberPermissions';
import './DashboardLayout.css';

const DashboardLayout = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, isSuperAdmin, isAdmin, isMember, logout } = useAuth();

  const isPureMember = !isSuperAdmin && !isAdmin && isMember;
  const memberPermissions = useMemberPermissions(isPureMember);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  let menus: { path: string; label: string; icon: React.ReactNode }[] = [];
  let roleLabel = 'MEMBER';

  if (isSuperAdmin) {
    roleLabel = 'SUPER ADMIN';
    menus = [
      { path: '/super-admin/ringkasan', label: 'Ringkasan', icon: <LayoutDashboard size={18} /> },
      { path: '/super-admin', label: 'Kelola Project', icon: <Folder size={18} /> },
      { path: '/super-admin/users', label: 'Admin & Member', icon: <Users size={18} /> },
      { path: '/super-admin/gudang', label: 'Gudang', icon: <Box size={18} /> },
      { path: '/super-admin/item', label: 'Item', icon: <Package size={18} /> },
      { path: '/super-admin/permission', label: 'Permission', icon: <Lock size={18} /> },
      { path: '/super-admin/transaksi', label: 'Transaksi', icon: <BarChart2 size={18} /> },
      { path: '/super-admin/laporan', label: 'Laporan', icon: <FileText size={18} /> },
    ];
  } else if (isAdmin && (location.pathname.startsWith('/admin') || !isMember)) {
    roleLabel = 'ADMIN';
    menus = [
      { path: '/admin', label: 'Ringkasan', icon: <LayoutDashboard size={18} /> },
      { path: '/admin/members', label: 'Members', icon: <Users size={18} /> },
      { path: '/admin/permission', label: 'Permission', icon: <Lock size={18} /> },
      { path: '/admin/gudang', label: 'Gudang', icon: <Box size={18} /> },
      { path: '/admin/item', label: 'Item', icon: <Package size={18} /> },
      { path: '/admin/transaksi', label: 'Transaksi', icon: <BarChart2 size={18} /> },
      { path: '/admin/laporan', label: 'Laporan', icon: <FileText size={18} /> },
    ];
  } else {
    roleLabel = 'MEMBER';
    const allMemberMenus = [
      { path: '/member', label: 'Ringkasan', icon: <LayoutDashboard size={18} />, requires: null as string | null },
      { path: '/member/gudang', label: 'Gudang', icon: <Box size={18} />, requires: 'warehouse.view' },
      { path: '/member/transaksi', label: 'Transaksi', icon: <BarChart2 size={18} />, requires: 'transaction.view' },
      { path: '/member/laporan', label: 'Laporan', icon: <FileText size={18} />, requires: 'report.view' },
    ];
    menus = allMemberMenus
      .filter((m) => m.requires === null || memberPermissions.includes(m.requires))
      .map(({ path, label, icon }) => ({ path, label, icon }));
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
          <div className="footer-role">{user?.name} — {roleLabel}</div>
          <button className="nav-item logout-btn" onClick={handleLogout}>
            <LogOut size={18} />
            <span>Keluar</span>
          </button>
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
