import { useTheme } from '@flanksource/clicky-ui/hooks';

// Cycles light -> dark -> system. The icon reflects the *resolved* theme so
// the user sees what's currently rendered; the title shows the explicit mode.
const NEXT: Record<string, 'light' | 'dark' | 'system'> = {
  light: 'dark',
  dark: 'system',
  system: 'light',
};

const ICON: Record<string, string> = {
  light: 'codicon:color-mode',
  dark: 'codicon:color-mode',
  system: 'codicon:device-desktop',
};

export function ThemeToggle() {
  const { theme, resolvedTheme, setTheme } = useTheme();
  const icon = theme === 'system' ? ICON.system : resolvedTheme === 'dark' ? 'codicon:moon' : 'codicon:sun';
  return (
    <button
      type="button"
      onClick={() => setTheme(NEXT[theme] ?? 'light')}
      title={`Theme: ${theme} — click for ${NEXT[theme]}`}
      aria-label={`Switch theme (currently ${theme})`}
      className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
    >
      <iconify-icon icon={icon} className="text-base" />
    </button>
  );
}
