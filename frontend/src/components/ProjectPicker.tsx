import { useEffect, useState } from 'react';
import { api } from '../lib/api';
import type { Project } from '../types';

interface ProjectPickerProps {
  value: string;
  onChange: (value: string) => void;
  includeAllOption?: boolean;
}

const ProjectPicker = ({ value, onChange, includeAllOption }: ProjectPickerProps) => {
  const [projects, setProjects] = useState<Project[]>([]);

  useEffect(() => {
    api
      .get<{ data: Project[] }>('/super-admin/projects')
      .then((res) => setProjects(res.data))
      .catch(() => setProjects([]));
  }, []);

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
