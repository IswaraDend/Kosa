import { useEffect, useState } from 'react';
import { api } from '../lib/api';
import type { Project } from '../types';
import { useSelectedProject } from './useSelectedProject';

export function useMemberPermissions(enabled: boolean) {
  const { selectedProjectId, setSelectedProjectId } = useSelectedProject();
  const [permissions, setPermissions] = useState<string[]>([]);

  useEffect(() => {
    if (!enabled) return;

    let cancelled = false;

    async function load() {
      let projectId = selectedProjectId;

      if (projectId === 'all' || !projectId) {
        const res = await api.get<{ data: Project[] }>('/member/projects');
        if (!res.data[0]) return;
        projectId = String(res.data[0].id);
        setSelectedProjectId(projectId);
        return;
      }

      const permRes = await api.get<{ data: string[] }>(`/member/projects/${projectId}/my-permissions`);
      if (!cancelled) setPermissions(permRes.data);
    }

    load().catch(() => {
      if (!cancelled) setPermissions([]);
    });

    return () => {
      cancelled = true;
    };
  }, [enabled, selectedProjectId, setSelectedProjectId]);

  return permissions;
}
