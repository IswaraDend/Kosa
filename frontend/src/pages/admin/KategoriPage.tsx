import SharedKategoriPage from '../super-admin/KategoriPage';

const KategoriPage = () => (
  <SharedKategoriPage
    apiBasePrefix="/admin/projects"
    scopeMode="path"
    projectEndpoint="/admin/projects"
    includeAllOption={false}
    showProjectPicker={false}
  />
);

export default KategoriPage;
