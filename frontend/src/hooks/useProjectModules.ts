import { useEffect, useState } from 'react';
import { api } from '../lib/api';
import type { Project } from '../types';
import { useSelectedProject } from './useSelectedProject';

/**
 * Fetches the enabled-modules list for the caller's current project (Admin
 * or Member — both are always scoped to a single project). Returns null
 * while loading/unresolved so callers can treat "unknown" distinctly from
 * "loaded but empty" if needed; menu filtering below just treats both as
 * "show nothing yet", same as useMemberPermissions does for permissions.
 */
export function useProjectModules(basePath: '/admin/projects' | '/member/projects', enabled: boolean) {
  const { selectedProjectId } = useSelectedProject();
  const [modules, setModules] = useState<string[]>([]);

  useEffect(() => {
    if (!enabled || !selectedProjectId || selectedProjectId === 'all') return;

    let cancelled = false;
    api
      .get<Project>(`${basePath}/${selectedProjectId}`)
      .then((project) => {
        if (!cancelled) setModules(project.modules ?? []);
      })
      .catch(() => {
        if (!cancelled) setModules([]);
      });

    return () => {
      cancelled = true;
    };
  }, [basePath, enabled, selectedProjectId]);

  return modules;
}
