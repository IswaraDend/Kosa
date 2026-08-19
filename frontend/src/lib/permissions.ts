/**
 * Builds the per-action predicate a shared page uses to decide which controls
 * to render.
 *
 * `undefined` means "no per-action gating", which is the correct default for
 * the two roles that are not permission-scoped: Super Admin bypasses
 * permissions entirely, and an Admin's reach is already bounded by the
 * project's enabled modules (see middleware.RequireModule). Only the member
 * wrappers pass a real list of granted codes.
 *
 * This mirrors the backend exactly — every code checked here is the same
 * string a member route names in middleware.RequirePermission, so hiding a
 * button and refusing the request cannot drift apart.
 */
export function makeCan(permissions?: string[]): (code: string) => boolean {
  if (!permissions) return () => true;
  return (code) => permissions.includes(code);
}
