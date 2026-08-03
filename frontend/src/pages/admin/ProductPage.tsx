import SharedProductPage from '../super-admin/ProductPage';

const ProductPage = () => (
  <SharedProductPage
    apiBasePrefix="/admin/projects"
    scopeMode="path"
    projectEndpoint="/admin/projects"
    includeAllOption={false}
    showProjectPicker={false}
  />
);

export default ProductPage;
