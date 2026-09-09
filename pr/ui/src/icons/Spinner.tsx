import type { IconProps } from '@flanksource/clicky-ui/icons';
import { UiLoader } from '@flanksource/clicky-ui/icons';

// Spinner is the app's single busy indicator: the clicky-ui loader glyph with
// `animate-spin` baked in, so callers (including config tables and status
// mappers that hand back a component reference) get the spin without having to
// remember to add the class themselves.
export function Spinner({ className, ...props }: IconProps) {
  return <UiLoader className={['animate-spin', className].filter(Boolean).join(' ')} {...props} />;
}
