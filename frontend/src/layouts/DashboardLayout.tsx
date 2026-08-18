import type { CSSProperties } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { LayoutDashboard, Users, Box, Package, PackagePlus, Factory, Lock, BarChart2, FileText, Folder, LogOut, Contact, Receipt } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';
import { useMemberPermissions } from '../hooks/useMemberPermissions';
import { useProjectModules } from '../hooks/useProjectModules';
import { LAYERS, groupByLayer, type ModuleCode, type NavLeaf, type SidebarBlock } from '../lib/modules';
import './DashboardLayout.css';

// path/label/icon for every module-bearing page, keyed by module code — reused
// to build each role's grouped sidebar from lib/modules' LAYERS ordering.
const ADMIN_MODULE_NAV: Record<ModuleCode, NavLeaf> = {
  warehouse: { path: '/admin/gudang', label: 'Gudang', icon: <Box size={18} /> },
  item: { path: '/admin/item', label: 'Item', icon: <Package size={18} /> },
  product: { path: '/admin/produk', label: 'Produk', icon: <PackagePlus size={18} /> },
  production: { path: '/admin/produksi', label: 'Produksi', icon: <Factory size={18} /> },
  transaction: { path: '/admin/transaksi', label: 'Transaksi', icon: <BarChart2 size={18} /> },
  customer: { path: '/admin/pelanggan', label: 'Pelanggan', icon: <Contact size={18} /> },
  invoice: { path: '/admin/invoice', label: 'Invoice', icon: <Receipt size={18} /> },
  report: { path: '/admin/laporan', label: 'Laporan', icon: <FileText size={18} /> },
};

const MEMBER_MODULE_NAV: Partial<Record<ModuleCode, NavLeaf & { permission: string }>> = {
  warehouse: { path: '/member/gudang', label: 'Gudang', icon: <Box size={18} />, permission: 'warehouse.view' },
  transaction: { path: '/member/transaksi', label: 'Transaksi', icon: <BarChart2 size={18} />, permission: 'transaction.view' },
  report: { path: '/member/laporan', label: 'Laporan', icon: <FileText size={18} />, permission: 'report.view' },
};

const DashboardLayout = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, isSuperAdmin, isAdmin, isMember, logout } = useAuth();

  const isPureMember = !isSuperAdmin && !isAdmin && isMember;
  const memberPermissions = useMemberPermissions(isPureMember);
  const adminModules = useProjectModules('/admin/projects', isAdmin && !isPureMember);
  const memberModules = useProjectModules('/member/projects', isPureMember);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  let blocks: SidebarBlock[] = [];
  let roleLabel = 'MEMBER';

  if (isSuperAdmin) {
    roleLabel = 'SUPER ADMIN';
    const items: NavLeaf[] = [
      { path: '/super-admin', label: 'Dashboard', icon: <LayoutDashboard size={18} /> },
      { path: '/super-admin/projects', label: 'Kelola Project', icon: <Folder size={18} /> },
      { path: '/super-admin/users', label: 'Admin & Member', icon: <Users size={18} /> },
      { path: '/super-admin/gudang', label: 'Gudang', icon: <Box size={18} /> },
      { path: '/super-admin/item', label: 'Item', icon: <Package size={18} /> },
      { path: '/super-admin/produk', label: 'Produk', icon: <PackagePlus size={18} /> },
      { path: '/super-admin/produksi', label: 'Produksi', icon: <Factory size={18} /> },
      { path: '/super-admin/pelanggan', label: 'Pelanggan', icon: <Contact size={18} /> },
      { path: '/super-admin/invoice', label: 'Invoice', icon: <Receipt size={18} /> },
      { path: '/super-admin/permission', label: 'Permission', icon: <Lock size={18} /> },
      { path: '/super-admin/transaksi', label: 'Transaksi', icon: <BarChart2 size={18} /> },
      { path: '/super-admin/laporan', label: 'Laporan', icon: <FileText size={18} /> },
    ];
    blocks = items.map((item) => ({ kind: 'item', item }));
  } else if (isAdmin && (location.pathname.startsWith('/admin') || !isMember)) {
    roleLabel = 'ADMIN';
    const core: NavLeaf[] = [
      { path: '/admin', label: 'Ringkasan', icon: <LayoutDashboard size={18} /> },
      { path: '/admin/members', label: 'Members', icon: <Users size={18} /> },
      { path: '/admin/permission', label: 'Permission', icon: <Lock size={18} /> },
    ];
    blocks = [...core.map((item) => ({ kind: 'item', item }) as SidebarBlock), ...groupByLayer(adminModules, ADMIN_MODULE_NAV)];
  } else {
    roleLabel = 'MEMBER';
    const core: NavLeaf[] = [{ path: '/member', label: 'Ringkasan', icon: <LayoutDashboard size={18} /> }];
    const grantedNav: Partial<Record<ModuleCode, NavLeaf>> = {};
    (Object.keys(MEMBER_MODULE_NAV) as ModuleCode[]).forEach((code) => {
      const entry = MEMBER_MODULE_NAV[code]!;
      if (memberPermissions.includes(entry.permission)) grantedNav[code] = entry;
    });
    blocks = [...core.map((item) => ({ kind: 'item', item }) as SidebarBlock), ...groupByLayer(memberModules, grantedNav)];
  }

  return (
    <div className="dashboard-container">
      <aside className="sidebar">
        <div className="sidebar-header">
          <span className="brand-mark">कोष</span>
          <h2>Kośa</h2>
        </div>

        <nav className="sidebar-nav">
          {blocks.map((block) =>
            block.kind === 'item' ? (
              <button
                key={block.item.path}
                className={`nav-item ${location.pathname === block.item.path ? 'active' : ''}`}
                onClick={() => navigate(block.item.path)}
              >
                {block.item.icon}
                <span>{block.item.label}</span>
              </button>
            ) : (
              <div className="nav-group" key={LAYERS[block.layerIdx].code} style={{ '--band-c': `var(${LAYERS[block.layerIdx].colorVar})` } as CSSProperties}>
                <div className="nav-group-label">
                  {LAYERS[block.layerIdx].name}
                  <span className="nav-group-deva">{LAYERS[block.layerIdx].deva}</span>
                </div>
                {block.items.map((item) => (
                  <button
                    key={item.path}
                    className={`nav-item ${location.pathname === item.path ? 'active' : ''}`}
                    onClick={() => navigate(item.path)}
                  >
                    <span className="nav-dot" />
                    {item.icon}
                    <span>{item.label}</span>
                  </button>
                ))}
              </div>
            ),
          )}
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
