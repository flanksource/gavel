import type { IconProps } from '@flanksource/clicky-ui/icons';

// VercelIcon is the Vercel wordmark triangle. clicky-ui has no Vercel glyph, so
// this app-local component keeps the "components, not strings" rule while
// matching the Ui* prop shape (size/title/className + spread SVG props).
export function VercelIcon({ size = '1em', className, title, ...props }: Omit<IconProps, 'ref'>) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width={size}
      height={size}
      viewBox="0 0 24 24"
      role={title ? 'img' : 'presentation'}
      aria-label={title}
      aria-hidden={title ? undefined : true}
      className={className}
      {...props}
    >
      <path fill="currentColor" d="M12 3 23 21H1L12 3Z" />
    </svg>
  );
}
