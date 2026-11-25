import type { ReactNode } from "react";
import { Button, DropdownMenu } from "@flanksource/clicky-ui";
import { UiGearSix } from "@flanksource/clicky-ui/icons";

export type AssignAllAction = { label: string; disabled?: boolean; onSelect: () => void };

export function PanelSettingsMenu({ settingsContent, assignAll }: { settingsContent: ReactNode; assignAll: AssignAllAction }) {
  return (
    <DropdownMenu
      role="dialog"
      align="right"
      menuLabel="Settings"
      className="shrink-0"
      menuClassName="w-96 max-w-[calc(100vw-16px)] max-h-[80dvh] overflow-auto"
      trigger={
        <Button size="icon" variant="ghost" className="size-7 shrink-0" aria-label="Settings" title="Agent settings and thread assignment">
          <UiGearSix className="size-4" aria-hidden />
        </Button>
      }
    >
      {close => <>
        {settingsContent}
        <section aria-label="Assign all open threads" className="border-t border-border p-3">
          <Button
            size="sm"
            variant="ghost"
            className="h-auto w-full justify-start whitespace-normal py-2 text-left"
            disabled={assignAll.disabled}
            onClick={() => { close(); assignAll.onSelect(); }}
          >
            {assignAll.label}
          </Button>
        </section>
      </>}
    </DropdownMenu>
  );
}
