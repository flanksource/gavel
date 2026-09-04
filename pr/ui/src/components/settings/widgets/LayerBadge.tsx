import type { SettingsLayer } from '../provenance';

// LayerBadge is the small provenance pill shown next to a field label — it names
// the config layer whose value is currently in effect for that field.
export function LayerBadge({ layer }: { layer: SettingsLayer }) {
  return (
    <span
      title={`Value set in the ${layer} layer`}
      className="ml-2 inline-flex shrink-0 items-center rounded-full border border-border bg-muted px-1.5 py-px font-mono text-[10px] font-semibold tracking-wide text-muted-foreground"
    >
      {layer}
    </span>
  );
}
