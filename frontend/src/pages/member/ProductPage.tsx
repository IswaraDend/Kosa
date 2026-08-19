import { useMemberPermissions } from '../../hooks/useMemberPermissions';
import { useProjectModules } from '../../hooks/useProjectModules';
import SharedProductPage from '../super-admin/ProductPage';
import { MEMBER_SCOPE } from './scope';

const ProductPage = () => {
  const permissions = useMemberPermissions(true);
  const modules = useProjectModules('/member/projects', true);

  return <SharedProductPage {...MEMBER_SCOPE} permissions={permissions}
      enabledModules={modules} />;
};

export default ProductPage;
