import type { CSSProperties } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { LayoutDashboard, Box, BarChart2, FileText, LogOut } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';
import { useMemberPermissions } from '../hooks/useMemberPermissions';
import { useProjectModules } from '../hooks/useProjectModules';
import { LAYERS, groupByLayer, type ModuleCode, type NavLeaf, type SidebarBlock } from '../lib/modules';
import './MemberLayout.css';

const MEMBER_MODULE_NAV: Partial<Record<ModuleCode, NavLeaf & { permission: string }>> = {
  warehouse: { path: '/member/gudang', label: 'Gudang', icon: <Box size={17} />, permission: 'warehouse.view' },
  transaction: { path: '/member/transaksi', label: 'Transaksi', icon: <BarChart2 size={17} />, permission: 'transaction.view' },
  report: { path: '/member/laporan', label: 'Laporan', icon: <FileText size={17} />, permission: 'report.view' },
};

const MemberLayout = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuth();
  const memberPermissions = useMemberPermissions(true);
  const memberModules = useProjectModules('/member/projects', true);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const grantedNav: Partial<Record<ModuleCode, NavLeaf>> = {};
  (Object.keys(MEMBER_MODULE_NAV) as ModuleCode[]).forEach((code) => {
    const entry = MEMBER_MODULE_NAV[code]!;
    if (memberPermissions.includes(entry.permission)) grantedNav[code] = entry;
  });

  const blocks: SidebarBlock[] = [
    { kind: 'item', item: { path: '/member', label: 'Dashboard', icon: <LayoutDashboard size={17} /> } },
    ...groupByLayer(memberModules, grantedNav),
  ];

  const initials = (user?.name || 'M')
    .split(' ')
    .map((s) => s[0])
    .slice(0, 2)
    .join('')
    .toUpperCase();

  return (
    <div className="member-app">
      <div className="member-shell">
        <aside className="member-rail">
          <div className="member-brand">
            <span className="member-brand-mark">कोष</span>
            <div>
              <div className="member-brand-name">Kośa</div>
              <span className="member-brand-tag">Member Deck</span>
            </div>
          </div>

          <nav className="member-nav">
            {blocks.map((block) =>
              block.kind === 'item' ? (
                <button
                  key={block.item.path}
                  className={`member-nav-item ${location.pathname === block.item.path ? 'active' : ''}`}
                  onClick={() => navigate(block.item.path)}
                >
                  {block.item.icon}
                  <span>{block.item.label}</span>
                </button>
              ) : (
                <div className="member-nav-group" key={LAYERS[block.layerIdx].code} style={{ '--band-c': `var(${LAYERS[block.layerIdx].colorVar})` } as CSSProperties}>
                  <div className="member-nav-group-label">
                    {LAYERS[block.layerIdx].name}
                    <span className="member-nav-group-deva">{LAYERS[block.layerIdx].deva}</span>
                  </div>
                  {block.items.map((item) => (
                    <button
                      key={item.path}
                      className={`member-nav-item ${location.pathname === item.path ? 'active' : ''}`}
                      onClick={() => navigate(item.path)}
                    >
                      <span className="member-nav-dot" />
                      {item.icon}
                      <span>{item.label}</span>
                    </button>
                  ))}
                </div>
              ),
            )}
          </nav>

          <div className="member-rail-foot">
            <div className="member-avatar">{initials}</div>
            <div>
              <div className="member-rail-foot-name">{user?.name}</div>
              <div className="member-rail-foot-role">Member</div>
            </div>
          </div>
          <button className="member-logout" onClick={handleLogout}>
            <LogOut size={16} />
            Keluar
          </button>
        </aside>

        <main className="member-main">
          <Outlet />
        </main>
      </div>
    </div>
  );
};

export default MemberLayout;
