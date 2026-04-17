import { paletteClass } from '../utils';

interface Props {
  src?: string;
  alt: string;
  size?: number;
  rounded?: 'full' | 'md';
  title?: string;
  href?: string;
  // colorKey selects the palette color for the initial-fallback. Defaults to
  // alt so repeated rendering of the same identity stays consistent. Pass the
  // full "owner/name" for repo avatars so two repos named "cli" under
  // different orgs render with different colors.
  colorKey?: string;
  onError?: (e: Event) => void;
}

export function Avatar({ src, alt, size = 20, rounded = 'full', title, href, colorKey, onError }: Props) {
  const shape = rounded === 'full' ? 'rounded-full' : 'rounded';
  const key = colorKey ?? alt;
  const baseClass = `inline-block shrink-0 ${shape}`;

  const content = src ? (
    <img
      src={src}
      alt={alt}
      title={title ?? alt}
      width={size}
      height={size}
      class={`${baseClass} bg-gray-100`}
      loading="lazy"
      onError={onError}
    />
  ) : (
    <span
      class={`${baseClass} ${paletteClass(key)} inline-flex items-center justify-center font-semibold`}
      style={{ width: size, height: size, fontSize: Math.max(9, Math.floor(size * 0.5)) }}
      title={title ?? alt}
    >
      {(alt.replace(/^@/, '').charAt(0) || '?').toUpperCase()}
    </span>
  );

  if (href) {
    return (
      <a
        href={href}
        target="_blank"
        rel="noopener"
        class="inline-flex shrink-0"
        onClick={(e) => e.stopPropagation()}
      >
        {content}
      </a>
    );
  }
  return content;
}
