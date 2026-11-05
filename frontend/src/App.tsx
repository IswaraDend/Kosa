import type { ReactNode } from 'react';
import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import Login from './pages/Login';
import DashboardLayout from './layouts/DashboardLayout';
import MemberLayout from './layouts/MemberLayout';
import RequireAuth from './components/RequireAuth';
import { useAuth } from './hooks/useAuth';
import { isActiveModule } from './lib/modules';
import DashboardPage from './pages/super-admin/DashboardPage';
import KelolaProjectPage from './pages/super-admin/KelolaProjectPage';
import UsersPage from './pages/super-admin/UsersPage';
import GudangPage from './pages/super-admin/GudangPage';
import ItemPage from './pages/super-admin/ItemPage';
import ProductPage from './pages/super-admin/ProductPage';
import ProduksiPage from './pages/super-admin/ProduksiPage';
import CustomerPage from './pages/super-admin/CustomerPage';
import InvoicePage from './pages/super-admin/InvoicePage';
import PermissionPage from './pages/super-admin/PermissionPage';
import ApiKeyPage from './pages/super-admin/ApiKeyPage';
import TransaksiPage from './pages/super-admin/TransaksiPage';
import LaporanPage from './pages/super-admin/LaporanPage';
import AdminRingkasanPage from './pages/admin/RingkasanPage';
import AdminMembersPage from './pages/admin/MembersPage';
import AdminPermissionPage from './pages/admin/PermissionPage';
import AdminApiKeyPage from './pages/admin/ApiKeyPage';
import AdminGudangPage from './pages/admin/GudangPage';
import AdminItemPage from './pages/admin/ItemPage';
import AdminProductPage from './pages/admin/ProductPage';
import AdminProduksiPage from './pages/admin/ProduksiPage';
import AdminCustomerPage from './pages/admin/CustomerPage';
import AdminInvoicePage from './pages/admin/InvoicePage';
import AdminTransaksiPage from './pages/admin/TransaksiPage';
import AdminLaporanPage from './pages/admin/LaporanPage';
import MemberDashboardPage from './pages/member/DashboardPage';
import MemberGudangPage from './pages/member/GudangPage';
import MemberItemPage from './pages/member/ItemPage';
import MemberProductPage from './pages/member/ProductPage';
import MemberProduksiPage from './pages/member/ProduksiPage';
import MemberCustomerPage from './pages/member/CustomerPage';
import MemberInvoicePage from './pages/member/InvoicePage';
import MemberTransaksiPage from './pages/member/TransaksiPage';
import MemberLaporanPage from './pages/member/LaporanPage';

// A module this build does not expose gets no route at all, so its pages are
// unreachable by typing the URL as well as by the (already filtered) sidebar.
// React Router skips non-element children, which is what makes the inline
// conditional below legal. The components stay imported and the backend stays
// untouched — flipping ACTIVE_MODULES in lib/modules.ts restores the screens.
const ifActive = (module: string, route: ReactNode) => (isActiveModule(module) ? route : null);

const HomeRedirect = () => {
  const { isAuthenticated, isSuperAdmin, isAdmin, isMember } = useAuth();
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  if (isSuperAdmin) return <Navigate to="/super-admin" replace />;
  if (isAdmin) return <Navigate to="/admin" replace />;
  if (isMember) return <Navigate to="/member" replace />;
  return <Navigate to="/login" replace />;
};

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
              <RequireAuth requireRole="super-admin">
                <Outlet />
              </RequireAuth>
            }
          >
            <Route index element={<DashboardPage />} />
            <Route path="projects" element={<KelolaProjectPage />} />
            <Route path="users" element={<UsersPage />} />
            {ifActive('warehouse', <Route path="gudang" element={<GudangPage />} />)}
            {ifActive('item', <Route path="item" element={<ItemPage />} />)}
            <Route path="produk" element={<ProductPage />} />
            {ifActive('production', <Route path="produksi" element={<ProduksiPage />} />)}
            <Route path="pelanggan" element={<CustomerPage />} />
            <Route path="invoice" element={<InvoicePage />} />
            <Route path="permission" element={<PermissionPage />} />
            {/* Not wrapped in ifActive: secret keys are project
                infrastructure, not one of the 8 business modules. */}
            <Route path="secret-key" element={<ApiKeyPage />} />
            {ifActive('transaction', <Route path="transaksi" element={<TransaksiPage />} />)}
            {ifActive('report', <Route path="laporan" element={<LaporanPage />} />)}
          </Route>

          <Route
            path="admin"
            element={
              <RequireAuth requireRole="admin">
                <Outlet />
              </RequireAuth>
            }
          >
            <Route index element={<AdminRingkasanPage />} />
            <Route path="members" element={<AdminMembersPage />} />
            <Route path="permission" element={<AdminPermissionPage />} />
            <Route path="secret-key" element={<AdminApiKeyPage />} />
            {ifActive('warehouse', <Route path="gudang" element={<AdminGudangPage />} />)}
            {ifActive('item', <Route path="item" element={<AdminItemPage />} />)}
            <Route path="produk" element={<AdminProductPage />} />
            {ifActive('production', <Route path="produksi" element={<AdminProduksiPage />} />)}
            <Route path="pelanggan" element={<AdminCustomerPage />} />
            <Route path="invoice" element={<AdminInvoicePage />} />
            {ifActive('transaction', <Route path="transaksi" element={<AdminTransaksiPage />} />)}
            {ifActive('report', <Route path="laporan" element={<AdminLaporanPage />} />)}
          </Route>

          <Route index element={<HomeRedirect />} />
        </Route>

        <Route
          path="/member"
          element={
            <RequireAuth requireRole="member">
              <MemberLayout />
            </RequireAuth>
          }
        >
          <Route index element={<MemberDashboardPage />} />
          {ifActive('warehouse', <Route path="gudang" element={<MemberGudangPage />} />)}
          {ifActive('item', <Route path="item" element={<MemberItemPage />} />)}
          <Route path="produk" element={<MemberProductPage />} />
          {ifActive('production', <Route path="produksi" element={<MemberProduksiPage />} />)}
          {ifActive('transaction', <Route path="transaksi" element={<MemberTransaksiPage />} />)}
          <Route path="pelanggan" element={<MemberCustomerPage />} />
          <Route path="invoice" element={<MemberInvoicePage />} />
          {ifActive('report', <Route path="laporan" element={<MemberLaporanPage />} />)}
        </Route>

        {/* Not /login: a bookmark to a page this build no longer routes (say
            /super-admin/laporan) would otherwise bounce a signed-in user to
            the login screen, which reads as a dropped session. HomeRedirect
            sends them to their own dashboard, and only signs out the ones who
            really are not authenticated. */}
        <Route path="*" element={<HomeRedirect />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
