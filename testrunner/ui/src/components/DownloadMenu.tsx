import { useState, useEffect, useRef } from 'preact/hooks';
import { downloadCurrentView } from '../export';
import type { RouteState } from '../routes';
import { buildExportRoute } from '../routes';

type Variant = 'neutral' | 'primary';

interface Props {
  routeState: RouteState;
  align?: 'left' | 'right';
  variant?: Variant;
  title?: string;
}

const triggerClass: Record<Variant, string> = {
  neutral: 'text-xs px-2 py-1 rounded border border-gray-300 text-gray-600 hover:bg-gray-200 transition-colors flex items-center gap-1',
  primary: 'text-xs px-2 py-1 rounded bg-blue-600 text-white hover:bg-blue-700 transition-colors flex items-center gap-1',
};

export function DownloadMenu({ routeState, align = 'right', variant = 'neutral', title }: Props) {
  const [open, setOpen] = useState(false);
  const wrapperRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const onMouseDown = (e: MouseEvent) => {
      if (!wrapperRef.current) return;
      if (!wrapperRef.current.contains(e.target as Node)) setOpen(false);
    };
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setOpen(false);
    };
    document.addEventListener('mousedown', onMouseDown);
    document.addEventListener('keydown', onKeyDown);
    return () => {
      document.removeEventListener('mousedown', onMouseDown);
      document.removeEventListener('keydown', onKeyDown);
    };
  }, [open]);

  const select = (format: 'json' | 'md') => {
    downloadCurrentView(routeState, format);
    setOpen(false);
  };

  const jsonURL = buildExportRoute(routeState, 'json');
  const mdURL = buildExportRoute(routeState, 'md');

  return (
    <div class="relative" ref={wrapperRef}>
      <button
        class={triggerClass[variant]}
        onClick={() => setOpen(v => !v)}
        title={title || 'Download as JSON or Markdown'}
        aria-haspopup="menu"
        aria-expanded={open}
      >
        <iconify-icon icon="codicon:cloud-download" />
        <span>Download</span>
        <iconify-icon icon="codicon:chevron-down" />
      </button>
      {open && (
        <div
          class={`absolute mt-1 ${align === 'right' ? 'right-0' : 'left-0'} bg-white border border-gray-200 rounded shadow-md z-20 min-w-32 py-1`}
          role="menu"
        >
          <button
            class="w-full text-left text-xs px-3 py-1.5 hover:bg-gray-100 flex items-center gap-2 text-gray-700"
            onClick={() => select('json')}
            title={jsonURL}
            role="menuitem"
          >
            <iconify-icon icon="codicon:json" />
            JSON
          </button>
          <button
            class="w-full text-left text-xs px-3 py-1.5 hover:bg-gray-100 flex items-center gap-2 text-gray-700"
            onClick={() => select('md')}
            title={mdURL}
            role="menuitem"
          >
            <iconify-icon icon="codicon:markdown" />
            Markdown
          </button>
        </div>
      )}
    </div>
  );
}
