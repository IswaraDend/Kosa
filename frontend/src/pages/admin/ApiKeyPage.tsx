import SharedApiKeyPage from '../super-admin/ApiKeyPage';

const ApiKeyPage = () => (
  <SharedApiKeyPage
    apiBasePrefix="/admin/projects"
    scopeMode="path"
    projectEndpoint="/admin/projects"
    includeAllOption={false}
    showProjectPicker={false}
    // Reading an existing key back in plaintext is a Super Admin power; the
    // /admin routes have no reveal endpoint at all.
    canReveal={false}
  />
);

export default ApiKeyPage;
