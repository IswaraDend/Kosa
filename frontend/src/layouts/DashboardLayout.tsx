import type { CSSProperties } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { LayoutDashboard, Users, Box, Package, PackagePlus, Factory, Lock, BarChart2, FileText, Folder, LogOut, Contact, Receipt, KeyRound, Tags } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';
import { useProjectModules } from '../hooks/useProjectModules';
import { LAYERS, groupByLayer, isActiveModule, type ModuleCode, type NavLeaf, type SidebarBlock } from '../lib/modules';
import './DashboardLayout.css';

// Super Admin's menu is a flat list rather than Pañca Kośa bands, because it
// is not scoped to one project and so has no per-project module set to group
// by. `module` marks the entries that belong to one of the 8 business modules;
// those are filtered by ACTIVE_MODULES, which is what hides a feature the
// build no longer exposes. Entries without it are core screens that are never
// gated (see models/module.go).
const SUPER_ADMIN_NAV: (NavLeaf & { module?: ModuleCode })[] = [
  { path: '/super-admin', label: 'Dashboard', icon: <LayoutDashboard size={18} /> },
  { path: '/super-admin/projects', label: 'Kelola Project', icon: <Folder size={18} /> },
  { path: '/super-admin/users', label: 'Admin & Member', icon: <Users size={18} /> },
  { path: '/super-admin/gudang', label: 'Gudang', icon: <Box size={18} />, module: 'warehouse' },
  { path: '/super-admin/item', label: 'Item', icon: <Package size={18} />, module: 'item' },
  { path: '/super-admin/produk', label: 'Produk', icon: <PackagePlus size={18} />, module: 'product' },
  { path: '/super-admin/kategori', label: 'Kategori', icon: <Tags size={18} />, module: 'product' },
  { path: '/super-admin/produksi', label: 'Produksi', icon: <Factory size={18} />, module: 'production' },
  { path: '/super-admin/pelanggan', label: 'Pelanggan', icon: <Contact size={18} />, module: 'customer' },
  { path: '/super-admin/invoice', label: 'Invoice', icon: <Receipt size={18} />, module: 'invoice' },
  { path: '/super-admin/permission', label: 'Permission', icon: <Lock size={18} /> },
  { path: '/super-admin/secret-key', label: 'Secret Key', icon: <KeyRound size={18} /> },
  { path: '/super-admin/transaksi', label: 'Transaksi', icon: <BarChart2 size={18} />, module: 'transaction' },
  { path: '/super-admin/laporan', label: 'Laporan', icon: <FileText size={18} />, module: 'report' },
];

// path/label/icon for every module-bearing page, keyed by module code — reused
// to build each role's grouped sidebar from lib/modules' LAYERS ordering.
const ADMIN_MODULE_NAV: Record<ModuleCode, NavLeaf | NavLeaf[]> = {
  warehouse: { path: '/admin/gudang', label: 'Gudang', icon: <Box size={18} /> },
  item: { path: '/admin/item', label: 'Item', icon: <Package size={18} /> },
  // Two entries under one module: Kategori belongs to Produk, not to a module
  // of its own, so it sits beside it in the same band.
  product: [
    { path: '/admin/produk', label: 'Produk', icon: <PackagePlus size={18} /> },
    { path: '/admin/kategori', label: 'Kategori', icon: <Tags size={18} /> },
  ],
  production: { path: '/admin/produksi', label: 'Produksi', icon: <Factory size={18} /> },
  transaction: { path: '/admin/transaksi', label: 'Transaksi', icon: <BarChart2 size={18} /> },
  customer: { path: '/admin/pelanggan', label: 'Pelanggan', icon: <Contact size={18} /> },
  invoice: { path: '/admin/invoice', label: 'Invoice', icon: <Receipt size={18} /> },
  report: { path: '/admin/laporan', label: 'Laporan', icon: <FileText size={18} /> },
};

// Serves Super Admin and Admin only. Members have their own shell
// (MemberLayout) on the separate /member route tree; the only way a member
// reaches this component is the "/" index, which HomeRedirect bounces to
// /member immediately — hence the empty fallback nav rather than a second
// copy of the member menu that could drift out of sync with MemberLayout.
const DashboardLayout = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, isSuperAdmin, isAdmin, logout } = useAuth();

  const adminModules = useProjectModules('/admin/projects', isAdmin && !isSuperAdmin);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  let blocks: SidebarBlock[] = [];
  let roleLabel = 'MEMBER';

  if (isSuperAdmin) {
    roleLabel = 'SUPER ADMIN';
    blocks = SUPER_ADMIN_NAV.filter((entry) => !entry.module || isActiveModule(entry.module)).map(
      (item) => ({ kind: 'item', item }),
    );
  } else if (isAdmin) {
    roleLabel = 'ADMIN';
    const core: NavLeaf[] = [
      { path: '/admin', label: 'Ringkasan', icon: <LayoutDashboard size={18} /> },
      { path: '/admin/members', label: 'Members', icon: <Users size={18} /> },
      { path: '/admin/permission', label: 'Permission', icon: <Lock size={18} /> },
      // In the core list, not in a Pañca Kośa band: a secret key is project
      // infrastructure rather than one of the 8 business modules, so it is
      // never module-gated — same as Members and Permission.
      { path: '/admin/secret-key', label: 'Secret Key', icon: <KeyRound size={18} /> },
    ];
    blocks = [...core.map((item) => ({ kind: 'item', item }) as SidebarBlock), ...groupByLayer(adminModules, ADMIN_MODULE_NAV)];
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
