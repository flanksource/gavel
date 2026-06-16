import type { ReactNode } from 'react';
import { DropdownMenu } from '@flanksource/clicky-ui/components';
import { Version } from '@flanksource/clicky-ui/data';

// Backend build metadata injected by the Go server (see pr/ui/handler.go),
// frontend metadata baked in by Vite (see pr/ui/vite.config.ts).
const backend = typeof window !== 'undefined' ? window.__GAVEL__ : undefined;
const frontend = { version: __GAVEL_UI_VERSION__, commit: __GAVEL_UI_COMMIT__ };

function VersionRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex items-baseline justify-between gap-4">
      <span className="whitespace-nowrap text-muted-foreground">{label}</span>
      <span className="min-w-0 text-right font-mono text-foreground">{children}</span>
    </div>
  );
}

// SettingsButton is the top-right gear menu. For now it surfaces the running
// versions — gavel backend, the embedded frontend bundle, and the clicky-ui
// component library — so build drift between the binary and its UI is visible.
export function SettingsButton() {
  const trigger = (
    <button
      type="button"
      title="Versions & settings"
      aria-label="Settings and version info"
      className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
    >
      <iconify-icon icon="codicon:settings-gear" className="text-base" />
    </button>
  );

  return (
    <DropdownMenu trigger={trigger} align="right" menuLabel="Versions" menuClassName="w-72">
      {() => (
        <div className="p-3 text-xs space-y-2.5">
          <div className="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
            Versions
          </div>
          <VersionRow label="clicky-ui">
            <Version date={false} />
          </VersionRow>
          <VersionRow label="gavel backend">
            {backend?.version || 'unknown'}
            {backend?.commit && backend.commit !== 'unknown' && (
              <span className="text-muted-foreground"> · {backend.commit}</span>
            )}
          </VersionRow>
          <VersionRow label="gavel frontend">
            {frontend.version}
            {frontend.commit !== 'unknown' && (
              <span className="text-muted-foreground"> · {frontend.commit}</span>
            )}
          </VersionRow>
        </div>
      )}
    </DropdownMenu>
  );
}
