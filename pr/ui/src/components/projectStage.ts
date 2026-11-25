import type { ProjectFileStatus } from './ProjectStatusView';

export type StageState = 'all' | 'partial' | 'none';

export interface ProjectEditor {
  id: string;
  label: string;
}

// An ignore rule request: an exact path, a directory (`path/`), or every file
// with the extension (`*.ext`). directory and extension are mutually exclusive.
export interface IgnoreRequest {
  path: string;
  directory?: boolean;
  extension?: string;
}

export const CONFLICT_REASON = 'Resolve conflicts before staging';
const LINKED_DEPS_REASON = 'Cannot stage dependencies linked outside this repository';

const hasStagedContent = (file: ProjectFileStatus) => file.state === 'staged' || file.state === 'both';

// stageStateOf summarises the index state of a set of files: every file fully
// staged, some content staged, or none.
export function stageStateOf(files: ProjectFileStatus[]): StageState {
  if (files.length > 0 && files.every(file => file.state === 'staged')) return 'all';
  return files.some(hasStagedContent) ? 'partial' : 'none';
}

// stageBlockReason is why a file cannot be staged, or null when it can.
// Unstaging is never blocked by the file itself, only by a queued commit.
export function stageBlockReason({ file, lockedGroup }: { file: ProjectFileStatus; lockedGroup: number | undefined }): string | null {
  if (lockedGroup !== undefined) return `Queued in commit group ${lockedGroup}`;
  if (file.state === 'conflict') return CONFLICT_REASON;
  if (file.linkedDependencies?.length && !hasStagedContent(file)) return LINKED_DEPS_REASON;
  return null;
}

// stageRequestPaths are the paths a stage or unstage click sends: unstaging
// takes the staged files; staging takes every file that is not fully staged and
// not blocked, so a folder with one conflict still stages the rest.
export function stageRequestPaths({ files, locked, unstage }: { files: ProjectFileStatus[]; locked: Map<string, number>; unstage: boolean }): string[] {
  return files
    .filter(file => unstage
      ? hasStagedContent(file) && !locked.has(file.path)
      : file.state !== 'staged' && stageBlockReason({ file, lockedGroup: locked.get(file.path) }) === null)
    .map(file => file.path);
}

// stagedCommitPaths are the files a commit takes: everything with staged
// content that no queued commit group already owns.
export function stagedCommitPaths(files: ProjectFileStatus[], locked: Map<string, number>): string[] {
  return files.filter(file => hasStagedContent(file) && !locked.has(file.path)).map(file => file.path).sort();
}
