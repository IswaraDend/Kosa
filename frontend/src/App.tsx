import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import Login from './pages/Login';
import DashboardLayout from './layouts/DashboardLayout';
import RequireAuth from './components/RequireAuth';
import AdminDashboard from './pages/AdminDashboard';
import MemberDashboard from './pages/MemberDashboard';
import RingkasanPage from './pages/super-admin/RingkasanPage';
import KelolaProjectPage from './pages/super-admin/KelolaProjectPage';
import UsersPage from './pages/super-admin/UsersPage';
import GudangPage from './pages/super-admin/GudangPage';
import PermissionPage from './pages/super-admin/PermissionPage';
import TransaksiPage from './pages/super-admin/TransaksiPage';
import LaporanPage from './pages/super-admin/LaporanPage';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />

        <Route
          path="/"
          element={
            <RequireAuth>
              <DashboardLayout />
            </RequireAuth>
          }
        >
          <Route
            path="super-admin"
            element={
              <RequireAuth requireSuperAdmin>
                <Outlet />
              </RequireAuth>
            }
          >
            <Route index element={<KelolaProjectPage />} />
            <Route path="ringkasan" element={<RingkasanPage />} />
            <Route path="users" element={<UsersPage />} />
            <Route path="gudang" element={<GudangPage />} />
            <Route path="permission" element={<PermissionPage />} />
            <Route path="transaksi" element={<TransaksiPage />} />
            <Route path="laporan" element={<LaporanPage />} />
          </Route>

          <Route path="admin/*" element={<AdminDashboard />} />
          <Route path="member/*" element={<MemberDashboard />} />
          <Route index element={<Navigate to="/login" replace />} />
        </Route>

        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
