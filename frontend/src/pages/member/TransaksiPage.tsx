import { useMemberPermissions } from '../../hooks/useMemberPermissions';
import SharedTransaksiPage from '../super-admin/TransaksiPage';
import { MEMBER_SCOPE } from './scope';

const TransaksiPage = () => {
  const permissions = useMemberPermissions(true);

  return <SharedTransaksiPage {...MEMBER_SCOPE} permissions={permissions} />;
};

export default TransaksiPage;
