// Runtime Iconify resolution, registered once at app start.
//
// pr/ui otherwise imports every glyph offline (see icons/issues.tsx,
// icons/tags.tsx, and the Ui* set from clicky-ui). That convention holds for
// everything the app itself draws. This provider exists for one case the
// convention cannot cover: a tag definition authored by a user may name any
// Iconify glyph, and the bundle cannot contain every glyph that might be typed.
//
// The trade-off is explicit: a name we bundle never touches the network, and
// only an uncurated one is fetched from api.iconify.design. An unreachable name
// renders nothing rather than a broken placeholder, and the tag editor shows a
// live preview so a bad name is caught while typing rather than on every later
// read.
import { Icon as IconifyRuntimeIcon } from '@iconify/react';
import { setFallbackIconProvider, type FallbackIconProps } from '@flanksource/clicky-ui/data';

import { setTagFallbackIcon } from './tags';

function RuntimeIcon({ name, className, size, alt }: FallbackIconProps) {
  if (!name) return null;
  return (
    <IconifyRuntimeIcon
      icon={name}
      className={className}
      {...(size != null ? { width: size, height: size } : {})}
      {...(alt ? { 'aria-label': alt, role: 'img' } : { 'aria-hidden': true })}
    />
  );
}

/**
 * registerIconifyFallback wires runtime icon-name resolution into both
 * clicky-ui (whose Icon renders a dashed placeholder for unresolved string
 * names) and this app's tag glyphs. Call once, before the first render.
 */
export function registerIconifyFallback(): void {
  setFallbackIconProvider(RuntimeIcon);
  setTagFallbackIcon(({ name, className }) => <RuntimeIcon name={name} className={className} />);
}
