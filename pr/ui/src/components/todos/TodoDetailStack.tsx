import type { ReactNode } from 'react';

// TodoDetailStack is the single-column master-detail idiom the full-width
// dashboard layout and the menubar both use: a list, and a todo's detail on top
// of it behind a back arrow. It layers the two rather than swapping them,
// because the list's scroll offset lives in the DOM node and dies with it —
// unmounting the list to show a detail meant Back landed at the top of the list
// however far down the reader had been. The table pays that twice: DataTable's
// clientReveal window is component state, so a remount also threw away every
// row past the first batch.
//
// `invisible` rather than `hidden`: display:none destroys the layout box, and
// the scroll offset goes with it when the box is rebuilt — the very thing this
// exists to avoid. visibility:hidden keeps the box (and the offset) and only
// skips painting. `inert` then takes the covered list out of the tab order and
// the accessibility tree, so it is hidden to a screen reader too rather than
// merely unpainted.
export function TodoDetailStack({ list, detail }: { list: ReactNode; detail?: ReactNode }) {
  const covered = !!detail;
  return (
    <div className="relative h-full min-h-0">
      <div className={`absolute inset-0${covered ? ' invisible' : ''}`} inert={covered}>
        {list}
      </div>
      {covered && <div className="absolute inset-0 bg-background">{detail}</div>}
    </div>
  );
}
