import { useProjectModules } from '../../hooks/useProjectModules';
import SharedProductPage from '../super-admin/ProductPage';

const ProductPage = () => {
  const modules = useProjectModules('/admin/projects', true);

  return (
    <SharedProductPage
      apiBasePrefix="/admin/projects"
      scopeMode="path"
      projectEndpoint="/admin/projects"
      includeAllOption={false}
      showProjectPicker={false}
      enabledModules={modules}
    />
  );
};

export default ProductPage;
