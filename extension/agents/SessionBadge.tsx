import { Badge } from "@flanksource/clicky-ui";

/** A session's name as a compact monospace pill, truncated to its container; the full name is its title. */
export function SessionBadge({ title }: { title: string }) {
  return (
    <Badge tone="neutral" size="xs" maxWidth="100%" clickToCopy={false} className="rounded-md bg-primary/10 font-mono text-[11px] text-primary">
      {title}
    </Badge>
  );
}
