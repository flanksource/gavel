// The platform dashboard's layout: a sidebar naming each section, and its subsections indented,
// beside the scrolling sections; a click scrolls to one, and the one being read stays highlighted.
import { useRef, useState, type ReactNode } from 'react';

// PlatformSection is one section; its subsections are parts of content, which marks each with the
// element id sectionAnchor(subsection.id).
export interface PlatformSection {
  id: string;
  label: string;
  content: ReactNode;
  subsections?: { id: string; label: string }[];
}

// readingOffset is how far below the scrolling area's top a section's top may be and still be
// the one being read.
const readingOffset = 16;

// sectionAnchor is the element id a section or subsection is scrolled to by.
export function sectionAnchor(id: string): string {
  return `platform-${id}`;
}

// PlatformSections shows header above the sections, and a sidebar entry for each section.
export function PlatformSections({ sections, header }: { sections: PlatformSection[]; header?: ReactNode }) {
  // entries are the sidebar's entries in page order: each section, then its subsections.
  const entries = sections.flatMap(section => [
    { id: section.id, label: section.label, depth: 0 },
    ...(section.subsections ?? []).map(subsection => ({ ...subsection, depth: 1 })),
  ]);
  const [reading, setReading] = useState('');
  // active is the entry being read, or the first while none is, such as before any scrolling.
  const active = entries.some(entry => entry.id === reading) ? reading : entries[0]?.id;
  const area = useRef<HTMLDivElement>(null);
  // chosen is the section last clicked, until the reader scrolls by hand: at the bottom a short
  // section can't scroll up to the top, so its position alone wouldn't say it's the one read.
  const chosen = useRef<string | null>(null);

  const show = (id: string) => {
    chosen.current = id;
    setReading(id);
    document.getElementById(sectionAnchor(id))?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  };

  const onScroll = () => {
    const root = area.current;
    if (!root || entries.length === 0) return;
    const atBottom = root.scrollHeight > root.clientHeight && root.scrollTop + root.clientHeight >= root.scrollHeight - 2;
    if (atBottom) {
      setReading(chosen.current ?? entries[entries.length - 1].id);
      return;
    }
    const top = root.getBoundingClientRect().top;
    let passed = entries[0].id;
    for (const entry of entries) {
      const element = document.getElementById(sectionAnchor(entry.id));
      if (element && element.getBoundingClientRect().top - top <= readingOffset) passed = entry.id;
    }
    setReading(passed);
  };

  const byHand = () => {
    chosen.current = null;
  };

  return (
    <div className="flex h-full min-h-0 text-sm">
      <nav aria-label="Platform sections" className="hidden w-48 shrink-0 flex-col gap-0.5 overflow-y-auto border-r border-border p-2 md:flex">
        {entries.map(entry => (
          <button
            key={entry.id}
            type="button"
            data-depth={entry.depth}
            aria-current={entry.id === active}
            onClick={() => show(entry.id)}
            className={`rounded py-1.5 pr-2 text-left hover:bg-muted ${entry.depth ? 'pl-6 text-xs' : 'pl-2'} ${entry.id === active ? 'bg-primary/10 font-medium' : ''}`}
          >
            {entry.label}
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
          <div key={section.id} id={sectionAnchor(section.id)} className="scroll-mt-4">
            {section.content}
          </div>
        ))}
      </div>
    </div>
  );
}
