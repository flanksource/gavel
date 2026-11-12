// The platform dashboard's layout: a sidebar naming each section beside the scrolling sections;
// a click scrolls to a section, and the section being read stays highlighted.
import { useRef, useState, type ReactNode } from 'react';

export interface PlatformSection {
  id: string;
  label: string;
  content: ReactNode;
}

// readingOffset is how far below the scrolling area's top a section's top may be and still be
// the one being read.
const readingOffset = 16;

function anchor(id: string): string {
  return `platform-${id}`;
}

// PlatformSections shows header above the sections, and a sidebar entry for each section.
export function PlatformSections({ sections, header }: { sections: PlatformSection[]; header?: ReactNode }) {
  const [reading, setReading] = useState('');
  // active is the section being read, or the first while none is, such as before any scrolling.
  const active = sections.some(section => section.id === reading) ? reading : sections[0]?.id;
  const area = useRef<HTMLDivElement>(null);
  // chosen is the section last clicked, until the reader scrolls by hand: at the bottom a short
  // section can't scroll up to the top, so its position alone wouldn't say it's the one read.
  const chosen = useRef<string | null>(null);

  const show = (id: string) => {
    chosen.current = id;
    setReading(id);
    document.getElementById(anchor(id))?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  };

  const onScroll = () => {
    const root = area.current;
    if (!root || sections.length === 0) return;
    const atBottom = root.scrollHeight > root.clientHeight && root.scrollTop + root.clientHeight >= root.scrollHeight - 2;
    if (atBottom) {
      setReading(chosen.current ?? sections[sections.length - 1].id);
      return;
    }
    const top = root.getBoundingClientRect().top;
    let passed = sections[0].id;
    for (const section of sections) {
      const element = document.getElementById(anchor(section.id));
      if (element && element.getBoundingClientRect().top - top <= readingOffset) passed = section.id;
    }
    setReading(passed);
  };

  const byHand = () => {
    chosen.current = null;
  };

  return (
    <div className="flex h-full min-h-0 text-sm">
      <nav aria-label="Platform sections" className="hidden w-48 shrink-0 flex-col gap-0.5 overflow-y-auto border-r border-border p-2 md:flex">
        {sections.map(section => (
          <button
            key={section.id}
            type="button"
            aria-current={section.id === active}
            onClick={() => show(section.id)}
            className={`rounded px-2 py-1.5 text-left hover:bg-muted ${section.id === active ? 'bg-primary/10 font-medium' : ''}`}
          >
            {section.label}
          </button>
        ))}
      </nav>
      <div
        ref={area}
        data-testid="platform-sections-scroll"
        onScroll={onScroll}
        onWheel={byHand}
        onTouchStart={byHand}
        onKeyDown={byHand}
        className="flex min-w-0 flex-1 flex-col gap-8 overflow-y-auto p-4"
      >
        {header}
        {sections.map(section => (
          <div key={section.id} id={anchor(section.id)} className="scroll-mt-4">
            {section.content}
          </div>
        ))}
      </div>
    </div>
  );
}
