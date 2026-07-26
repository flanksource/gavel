// ProjectsPlaceholder stands in for a workspace list that has nothing to render.
//
// The three cases are deliberately distinct: /api/projects still in flight,
// /api/projects failed, and genuinely no projects configured. Collapsing the
// first two into the empty copy is what made the menubar assert "No projects
// configured" while it was only waiting for a response.
export function ProjectsPlaceholder({ loaded, error, emptyText, className }: {
  loaded: boolean;
  error?: string;
  emptyText: string;
  className: string;
}) {
  if (error) {
    return <div role="alert" className={`${className} text-destructive`}>{error}</div>;
  }
  if (!loaded) {
    return <div className={className} aria-busy="true">Loading projects…</div>;
  }
  return <div className={className}>{emptyText}</div>;
}
