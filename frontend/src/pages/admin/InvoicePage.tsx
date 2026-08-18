import SharedInvoicePage from '../super-admin/InvoicePage';

const InvoicePage = () => (
  <SharedInvoicePage
    apiBasePrefix="/admin/projects"
    scopeMode="path"
    projectEndpoint="/admin/projects"
    includeAllOption={false}
    showProjectPicker={false}
  />
);

export default InvoicePage;
