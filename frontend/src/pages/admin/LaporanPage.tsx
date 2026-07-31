import SharedLaporanPage from '../super-admin/LaporanPage';

const LaporanPage = () => (
  <SharedLaporanPage
    apiBasePrefix="/admin/projects"
    scopeMode="path"
    projectEndpoint="/admin/projects"
    includeAllOption={false}
  />
);

export default LaporanPage;
