import type { ComponentType } from 'react';
import { useTheme } from '@flanksource/clicky-ui/hooks';
import { Button } from '@flanksource/clicky-ui/components';
import type { IconProps } from '@flanksource/clicky-ui/icons';
import { UiDesktop, UiMoon, UiSun } from '@flanksource/clicky-ui/icons';

// Cycles light -> dark -> system. The icon reflects the *resolved* theme so
// the user sees what's currently rendered; the title shows the explicit mode.
const NEXT: Record<string, 'light' | 'dark' | 'system'> = {
  light: 'dark',
  dark: 'system',
  system: 'light',
};

const ICON: Record<string, ComponentType<IconProps>> = {
  light: UiDesktop,
  dark: UiDesktop,
  system: UiDesktop,
};

export function ThemeToggle() {
  const { theme, resolvedTheme, setTheme } = useTheme();
  const Icon = theme === 'system' ? ICON.system : resolvedTheme === 'dark' ? UiMoon : UiSun;
  return (
    <Button
      variant="ghost"
      size="icon"
      type="button"
      onClick={() => setTheme(NEXT[theme] ?? 'light')}
      title={`Theme: ${theme} — click for ${NEXT[theme]}`}
      aria-label={`Switch theme (currently ${theme})`}
      className="h-8 w-8 text-muted-foreground hover:bg-accent hover:text-foreground"
    >
      <Icon className="text-base" />
    </Button>
  );
}
