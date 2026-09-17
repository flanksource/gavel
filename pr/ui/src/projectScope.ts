import type { PRItem, Project } from './types';

export function applyProjectScope(projects: Project[], prs: PRItem[], selected: string) {
  if (!selected) return { projects, prs };

  const project = projects.find(candidate => candidate.name === selected);
  if (!project) return { projects: [], prs: [] };

  const repos = new Set(project.repos);
  return {
    projects: [project],
    prs: prs.filter(pr => repos.has(pr.repo)),
  };
}
