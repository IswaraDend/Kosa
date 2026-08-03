import { useEffect } from 'react';
import { api } from '../lib/api';
import type { Project } from '../types';
import { useSelectedProject } from './useSelectedProject';

/**
 * Silently fetches the caller's project list and selects the first one if
 * the currently-selected project isn't in that list (or none is selected
 * yet) — no dropdown UI. Used by Admin/Member pages, which are always
 * scoped to a single project, so there's nothing for the user to pick.
 */
export function useProjectAutoSelect(endpoint: string, enabled: boolean) {
  const { selectedProjectId, setSelectedProjectId } = useSelectedProject();

  useEffect(() => {
    if (!enabled) return;
    api
      .get<{ data: Project[] }>(endpoint)
      .then((res) => {
        const stillValid = res.data.some((p) => String(p.id) === selectedProjectId);
        if (!stillValid && res.data[0]) {
          setSelectedProjectId(String(res.data[0].id));
        }
      })
      .catch(() => {});
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [endpoint, enabled]);
}
