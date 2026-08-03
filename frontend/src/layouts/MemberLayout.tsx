import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { LayoutDashboard, Box, BarChart2, FileText, Radio, LogOut } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';
import { useMemberPermissions } from '../hooks/useMemberPermissions';
import './MemberLayout.css';

const MemberLayout = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuth();
  const memberPermissions = useMemberPermissions(true);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const allMenus = [
    { path: '/member', label: 'Dashboard', icon: <LayoutDashboard size={17} />, requires: null as string | null },
    { path: '/member/gudang', label: 'Gudang', icon: <Box size={17} />, requires: 'warehouse.view' },
    { path: '/member/transaksi', label: 'Transaksi', icon: <BarChart2 size={17} />, requires: 'transaction.view' },
    { path: '/member/laporan', label: 'Laporan', icon: <FileText size={17} />, requires: 'report.view' },
  ];

  const menus = allMenus.filter((m) => m.requires === null || memberPermissions.includes(m.requires));

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
            <div className="member-brand-mark">
              <Radio size={16} />
            </div>
            <div>
              <div className="member-brand-name">Stockpulse</div>
              <span className="member-brand-tag">Member Deck</span>
            </div>
          </div>

          <nav className="member-nav">
            {menus.map((menu) => (
              <button
                key={menu.path}
                className={`member-nav-item ${location.pathname === menu.path ? 'active' : ''}`}
                onClick={() => navigate(menu.path)}
              >
                {menu.icon}
                <span>{menu.label}</span>
              </button>
            ))}
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
