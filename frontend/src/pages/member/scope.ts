// Every member page is a thin wrapper around the corresponding super-admin
// page, differing only in how it is scoped: path-based under
// /member/projects/:projectId, with no project picker because a member always
// works inside exactly one project.
export const MEMBER_SCOPE = {
  apiBasePrefix: '/member/projects',
  scopeMode: 'path',
  projectEndpoint: '/member/projects',
  includeAllOption: false,
  showProjectPicker: false,
} as const;
