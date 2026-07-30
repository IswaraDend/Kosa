import { createContext, useState, type ReactNode } from 'react';

interface ProjectContextValue {
  selectedProjectId: string;
  setSelectedProjectId: (id: string) => void;
}

// eslint-disable-next-line react-refresh/only-export-components
export const ProjectContext = createContext<ProjectContextValue | null>(null);

export function ProjectProvider({ children }: { children: ReactNode }) {
  const [selectedProjectId, setSelectedProjectIdState] = useState<string>(
    () => localStorage.getItem('selectedProjectId') || 'all',
  );

  const setSelectedProjectId = (id: string) => {
    localStorage.setItem('selectedProjectId', id);
    setSelectedProjectIdState(id);
  };

  return (
    <ProjectContext.Provider value={{ selectedProjectId, setSelectedProjectId }}>
      {children}
    </ProjectContext.Provider>
  );
}
