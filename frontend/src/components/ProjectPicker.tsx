import { useEffect, useState } from 'react';
import { api } from '../lib/api';
import type { Project } from '../types';

interface ProjectPickerProps {
  value: string;
  onChange: (value: string) => void;
  includeAllOption?: boolean;
  endpoint?: string;
}

const ProjectPicker = ({ value, onChange, includeAllOption, endpoint = '/super-admin/projects' }: ProjectPickerProps) => {
  const [projects, setProjects] = useState<Project[]>([]);

  useEffect(() => {
    api
      .get<{ data: Project[] }>(endpoint)
      .then((res) => {
        setProjects(res.data);
        const stillValid = (includeAllOption && value === 'all') || res.data.some((p) => String(p.id) === value);
        if (!stillValid && res.data[0]) {
          onChange(String(res.data[0].id));
        }
      })
      .catch(() => setProjects([]));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [endpoint]);

  return (
    <select className="form-select" value={value} onChange={(e) => onChange(e.target.value)}>
      {includeAllOption && <option value="all">Semua Project</option>}
      {projects.map((p) => (
        <option key={p.id} value={p.id}>
          {p.name}
        </option>
      ))}
    </select>
  );
};

export default ProjectPicker;
