import { Select } from "@flanksource/clicky-ui";

/** The dashboard's projects; a stored project the dashboard no longer lists stays selectable so it is not silently replaced. */
export function ProjectSelect({ projects, value, onChange, className }: { projects: string[] | undefined; value: string | undefined; onChange: (project: string) => void; className?: string }) {
  const names = projects ?? [];
  return (
    <Select
      aria-label="Project"
      {...(className ? { className } : {})}
      value={value ?? ""}
      options={[
        { value: "", label: projects ? "Choose a project" : "Loading projects…", disabled: true },
        ...names.map(name => ({ value: name, label: name })),
        ...(value && projects && !names.includes(value) ? [{ value, label: `${value} (not on this dashboard)` }] : []),
      ]}
      onChange={event => onChange(event.target.value)}
    />
  );
}
