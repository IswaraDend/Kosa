import SharedCustomerPage from '../super-admin/CustomerPage';

const CustomerPage = () => (
  <SharedCustomerPage
    apiBasePrefix="/admin/projects"
    scopeMode="path"
    projectEndpoint="/admin/projects"
    includeAllOption={false}
    showProjectPicker={false}
  />
);

export default CustomerPage;
