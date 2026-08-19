import { useMemberPermissions } from '../../hooks/useMemberPermissions';
import SharedItemPage from '../super-admin/ItemPage';
import { MEMBER_SCOPE } from './scope';

const ItemPage = () => {
  const permissions = useMemberPermissions(true);

  return <SharedItemPage {...MEMBER_SCOPE} permissions={permissions} />;
};

export default ItemPage;
