import SharedLaporanPage from '../super-admin/LaporanPage';

const LaporanPage = () => (
  <SharedLaporanPage
    apiBasePrefix="/member/projects"
    scopeMode="path"
    projectEndpoint="/member/projects"
    includeAllOption={false}
    showProjectPicker={false}
  />
);

export default LaporanPage;
