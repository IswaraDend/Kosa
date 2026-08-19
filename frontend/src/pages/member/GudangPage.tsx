import { useMemberPermissions } from '../../hooks/useMemberPermissions';
import SharedGudangPage from '../super-admin/GudangPage';
import { MEMBER_SCOPE } from './scope';

const GudangPage = () => {
  const permissions = useMemberPermissions(true);

  return <SharedGudangPage {...MEMBER_SCOPE} permissions={permissions} />;
};

export default GudangPage;
