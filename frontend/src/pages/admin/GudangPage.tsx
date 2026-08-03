import SharedGudangPage from '../super-admin/GudangPage';

const GudangPage = () => (
  <SharedGudangPage
    apiBasePrefix="/admin/projects"
    scopeMode="path"
    projectEndpoint="/admin/projects"
    includeAllOption={false}
    showProjectPicker={false}
  />
);

export default GudangPage;
