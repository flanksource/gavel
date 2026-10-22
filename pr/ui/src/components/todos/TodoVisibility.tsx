// The visibility of a hosted org's todo: private to its creator, or public to every member.
// Shown only on a hosted dashboard; a local one shows every todo.
import { Button, DropdownMenu, Select } from '@flanksource/clicky-ui/components';
import { UiCheck, UiChevronDown } from '@flanksource/clicky-ui/icons';
import type { TodoVisibility } from '../../types';
import { inputClass } from './format';

export const todoVisibilities: { value: TodoVisibility; label: string }[] = [
  { value: 'private', label: 'Private: only you' },
  { value: 'public', label: 'Public: everyone in the organization' },
];

export function TodoVisibilitySelect({ value, onChange, disabled }: {
  value: TodoVisibility;
  onChange: (value: TodoVisibility) => void;
  disabled?: boolean;
}) {
  return (
    <Select
      value={value}
      onChange={e => onChange(e.currentTarget.value as TodoVisibility)}
      className={inputClass}
      aria-label="Visibility"
      disabled={disabled}
    >
      {todoVisibilities.map(option => <option key={option.value} value={option.value}>{option.label}</option>)}
    </Select>
  );
}

// VisibilityMenu shows a todo's visibility in the detail's field row and changes it.
export function VisibilityMenu({ value, disabled, onSelect }: {
  value: TodoVisibility;
  disabled?: boolean;
  onSelect: (value: TodoVisibility) => void;
}) {
  return (
    <DropdownMenu
      align="left"
      menuLabel="Update todo visibility"
      menuClassName="w-72"
      trigger={
        <Button
          variant="ghost"
          type="button"
          disabled={disabled}
          className="inline-flex h-7 items-center gap-1.5 rounded-md border border-border bg-muted/20 px-2 text-[11px] font-medium text-muted-foreground hover:bg-muted disabled:opacity-50"
          title="Update visibility"
          aria-label="Update todo visibility"
        >
          <span>Visibility</span>
          <span className="capitalize text-foreground">{value}</span>
          <UiChevronDown className="text-[11px] opacity-70" />
        </Button>
      }
    >
      {close => (
        <div className="p-1 text-xs">
          {todoVisibilities.map(option => (
            <Button
              key={option.value}
              variant="ghost"
              type="button"
              disabled={disabled}
              onClick={() => {
                close();
                if (option.value !== value) onSelect(option.value);
              }}
              className="flex h-auto w-full items-center justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted"
            >
              <span className="min-w-0 flex-1 text-foreground">{option.label}</span>
              {option.value === value && <UiCheck className="text-xs text-primary" />}
            </Button>
          ))}
        </div>
      )}
    </DropdownMenu>
  );
}
