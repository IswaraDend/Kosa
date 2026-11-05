import { useMemberPermissions } from '../../hooks/useMemberPermissions';
import SharedKategoriPage from '../super-admin/KategoriPage';
import { MEMBER_SCOPE } from './scope';

const KategoriPage = () => {
  const permissions = useMemberPermissions(true);

  return <SharedKategoriPage {...MEMBER_SCOPE} permissions={permissions} />;
};

export default KategoriPage;
