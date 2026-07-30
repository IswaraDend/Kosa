import { useContext } from 'react';
import { ProjectContext } from '../context/ProjectContext';

export function useSelectedProject() {
  const ctx = useContext(ProjectContext);
  if (!ctx) {
    throw new Error('useSelectedProject harus dipakai di dalam ProjectProvider');
  }
  return ctx;
}
