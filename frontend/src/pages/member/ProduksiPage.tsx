import { useMemberPermissions } from '../../hooks/useMemberPermissions';
import SharedProduksiPage from '../super-admin/ProduksiPage';
import { MEMBER_SCOPE } from './scope';

const ProduksiPage = () => {
  const permissions = useMemberPermissions(true);

  return <SharedProduksiPage {...MEMBER_SCOPE} permissions={permissions} />;
};

export default ProduksiPage;
