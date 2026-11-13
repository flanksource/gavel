import type { ProjectFileStatus } from './ProjectStatusView';

export function ProjectLinkedDepsBadge({ file }: { file: ProjectFileStatus }) {
  if (!file.linkedDependencies?.length) return null;
  const details = file.linkedDependencies.map(dep => `${dep.name} → ${dep.target}`).join('\n');
  return <span
    className="shrink-0 rounded bg-amber-500/10 px-1.5 py-0.5 text-[10px] font-medium text-amber-700 dark:text-amber-400"
    title={`Cannot commit dependencies linked outside this repository.\n${details}`}
  >Linked dependencies</span>;
}
