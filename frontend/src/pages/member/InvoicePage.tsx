import { useMemberPermissions } from '../../hooks/useMemberPermissions';
import SharedInvoicePage from '../super-admin/InvoicePage';
import { MEMBER_SCOPE } from './scope';

const InvoicePage = () => {
  const permissions = useMemberPermissions(true);

  return <SharedInvoicePage {...MEMBER_SCOPE} permissions={permissions} />;
};

export default InvoicePage;
