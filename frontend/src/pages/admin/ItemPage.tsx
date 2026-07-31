import SharedItemPage from '../super-admin/ItemPage';

const ItemPage = () => (
  <SharedItemPage
    apiBasePrefix="/admin/projects"
    scopeMode="path"
    projectEndpoint="/admin/projects"
    includeAllOption={false}
  />
);

export default ItemPage;
