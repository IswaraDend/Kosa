import SharedTransaksiPage from '../super-admin/TransaksiPage';

const TransaksiPage = () => (
  <SharedTransaksiPage
    apiBasePrefix="/admin/projects"
    scopeMode="path"
    projectEndpoint="/admin/projects"
    includeAllOption={false}
    canCreate
  />
);

export default TransaksiPage;
