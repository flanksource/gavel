import { useEffect, useMemo, useRef, useState } from 'react';
import type { Project } from '../../types';

export function usePRTodoProjectWorkspace({
  open,
  prRepo,
  workspaces,
  onDirChange,
  onProjectsChanged,
}: {
  open: boolean;
  prRepo: string;
  workspaces: Project[];
  onDirChange: (dir: string) => void;
  onProjectsChanged?: () => void;
}) {
  const [addProjectOpen, setAddProjectOpen] = useState(false);
  const [createdProjects, setCreatedProjects] = useState<Project[]>([]);
  const wasOpen = useRef(false);
  const defaults = useMemo(() => ({
    name: prRepo.split('/').pop() || prRepo,
    dir: '',
    repos: [prRepo],
  }), [prRepo]);
  const choices = useMemo(() => {
    const merged = [...workspaces];
    for (const project of createdProjects) {
      if (!merged.some(existing => existing.name === project.name || existing.dir === project.dir)) {
        merged.push(project);
      }
    }
    return merged;
  }, [workspaces, createdProjects]);

  useEffect(() => {
    if (!open) {
      wasOpen.current = false;
      setAddProjectOpen(false);
      setCreatedProjects([]);
      return;
    }
    if (!wasOpen.current) {
      wasOpen.current = true;
      setAddProjectOpen(false);
      setCreatedProjects([]);
    }
  }, [open]);

  function projectSaved(project: Project) {
    setCreatedProjects(current => (
      current.some(existing => existing.name === project.name || existing.dir === project.dir)
        ? current
        : [...current, project]
    ));
    onDirChange(project.dir);
    setAddProjectOpen(false);
    onProjectsChanged?.();
  }

  return {
    addProjectOpen,
    choices,
    defaults,
    repoOptions: defaults.repos,
    openAddProject: () => setAddProjectOpen(true),
    closeAddProject: () => setAddProjectOpen(false),
    projectSaved,
  };
}
