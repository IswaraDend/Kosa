import { useMemberPermissions } from '../../hooks/useMemberPermissions';
import SharedCustomerPage from '../super-admin/CustomerPage';
import { MEMBER_SCOPE } from './scope';

const CustomerPage = () => {
  const permissions = useMemberPermissions(true);

  return <SharedCustomerPage {...MEMBER_SCOPE} permissions={permissions} />;
};

export default CustomerPage;
