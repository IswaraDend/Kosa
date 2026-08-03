import SharedProduksiPage from '../super-admin/ProduksiPage';

const ProduksiPage = () => (
  <SharedProduksiPage
    apiBasePrefix="/admin/projects"
    scopeMode="path"
    projectEndpoint="/admin/projects"
    includeAllOption={false}
    showProjectPicker={false}
  />
);

export default ProduksiPage;
