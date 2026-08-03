import { useMemberPermissions } from '../../hooks/useMemberPermissions';
import SharedTransaksiPage from '../super-admin/TransaksiPage';

const TransaksiPage = () => {
  const permissions = useMemberPermissions(true);

  return (
    <SharedTransaksiPage
      apiBasePrefix="/member/projects"
      scopeMode="path"
      projectEndpoint="/member/projects"
      includeAllOption={false}
      showProjectPicker={false}
      canCreate={permissions.includes('transaction.create')}
    />
  );
};

export default TransaksiPage;
