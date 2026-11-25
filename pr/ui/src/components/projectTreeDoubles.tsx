import type { ReactNode } from 'react';

interface MenuItem {
  label: ReactNode;
  group?: string;
  title?: string;
  disabled?: boolean;
  onSelect: () => void;
}

/* oxlint-disable clicky-ui/prefer-clicky-components --
   Test doubles for clicky-ui's DropdownMenu and ConfirmDialog, shared by the
   specs that replace '@flanksource/clicky-ui/components': the real menu mounts
   floating-ui, which jsdom cannot lay out. Every item renders as a menuitem
   button, so a spec can read and press it without opening the menu. */
export function DropdownMenuDouble({ trigger, items = [], onOpenChange }: { trigger: ReactNode; items?: MenuItem[]; onOpenChange?: (open: boolean) => void }) {
  return (
    <div>
      <span onClickCapture={() => onOpenChange?.(true)}>{trigger}</span>
      <div role="menu">
        {items.map(item => (
          <button key={`${item.group}:${String(item.label)}`} type="button" role="menuitem" data-group={item.group} title={item.title} disabled={item.disabled} onClick={item.onSelect}>
            {item.label}
          </button>
        ))}
      </div>
    </div>
  );
}

export function ConfirmDialogDouble({ open, title, warning, confirmLabel, onConfirm, onCancel }: {
  open: boolean;
  title: ReactNode;
  warning?: ReactNode;
  confirmLabel?: string;
  onConfirm: () => void;
  onCancel: () => void;
}) {
  if (!open) return null;
  return (
    <section role="alertdialog" aria-label={String(title)}>
      <p>{warning}</p>
      <button type="button" onClick={onConfirm}>{confirmLabel}</button>
      <button type="button" onClick={onCancel}>Cancel</button>
    </section>
  );
}
/* oxlint-enable clicky-ui/prefer-clicky-components */
