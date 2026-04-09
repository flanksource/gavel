interface Props {
  src?: string;
  alt: string;
  size?: number;
  rounded?: 'full' | 'md';
  title?: string;
  href?: string;
}

export function Avatar({ src, alt, size = 20, rounded = 'full', title, href }: Props) {
  const shape = rounded === 'full' ? 'rounded-full' : 'rounded';
  const baseClass = `inline-block shrink-0 bg-gray-100 ${shape}`;

  const content = src ? (
    <img
      src={src}
      alt={alt}
      title={title ?? alt}
      width={size}
      height={size}
      class={baseClass}
      loading="lazy"
    />
  ) : (
    <span
      class={`${baseClass} inline-flex items-center justify-center font-semibold text-gray-500`}
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
