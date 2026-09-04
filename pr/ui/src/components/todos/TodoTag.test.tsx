import { describe, expect, it } from 'vitest';
import { render, screen } from '@testing-library/react';
import type { TodoTagDef } from '../../types';
import { TodoTag, TodoTagRow } from './TodoTag';
import { buildTagIndex } from './tagResolve';

const defs: TodoTagDef[] = [
  { name: 'bug', color: 'red', icon: 'debug', iconify: 'ion:bug', scope: 'builtin', description: 'Broken' },
  { name: 'area', color: 'teal', scope: 'global' },
];
const index = buildTagIndex(defs);

describe('TodoTag', () => {
  it('renders the label with its definition colour', () => {
    const { container } = render(<TodoTag tag={index.resolve('bug')} />);
    expect(screen.getByText('bug')).toBeTruthy();
    expect(container.querySelector('[data-tag="bug"]')?.className).toContain('bg-red-100');
  });

  it('shows only the value when the key is suppressed', () => {
    render(<TodoTag tag={index.resolve('area/ui')} showKey={false} />);
    expect(screen.getByText('ui')).toBeTruthy();
    expect(screen.queryByText('area/ui')).toBeNull();
  });

  it('shows the full token by default', () => {
    render(<TodoTag tag={index.resolve('area/ui')} />);
    expect(screen.getByText('area/ui')).toBeTruthy();
  });

  it('colours an undefined tag rather than leaving it bare', () => {
    const { container } = render(<TodoTag tag={index.resolve('never-defined')} />);
    const className = container.querySelector('[data-tag="never-defined"]')?.className ?? '';
    expect(className).toMatch(/bg-\w+-100/);
  });

  it('renders no text in glyph-only mode', () => {
    render(<TodoTag tag={index.resolve('bug')} glyphOnly />);
    expect(screen.queryByText('bug')).toBeNull();
  });

  // A chip inside a list row must not be interactive: clicky's Badge defaults
  // clickToCopy on, which turns the chip into a button that swallows the row's
  // own click. This chip is a plain span for exactly that reason.
  it('is not an interactive element', () => {
    const { container } = render(<TodoTag tag={index.resolve('bug')} />);
    expect(container.querySelector('button')).toBeNull();
    expect(container.querySelector('a')).toBeNull();
  });

  it('exposes the description as a tooltip', () => {
    const { container } = render(<TodoTag tag={index.resolve('bug')} />);
    expect(container.querySelector('[data-tag="bug"]')?.getAttribute('title')).toContain('Broken');
  });
});

describe('TodoTagRow', () => {
  it('renders nothing when there are no labels', () => {
    const { container } = render(<TodoTagRow labels={[]} index={index} />);
    expect(container.firstChild).toBeNull();
  });

  it('caps visible chips and counts the rest', () => {
    render(<TodoTagRow labels={['a', 'b', 'c', 'd']} index={index} max={2} />);
    expect(screen.getByText('a')).toBeTruthy();
    expect(screen.getByText('b')).toBeTruthy();
    expect(screen.queryByText('c')).toBeNull();
    expect(screen.getByText('+2')).toBeTruthy();
  });

  it('names the hidden labels on the overflow chip', () => {
    const { container } = render(<TodoTagRow labels={['a', 'b', 'c']} index={index} max={1} />);
    expect(container.querySelector('[title="b, c"]')).toBeTruthy();
  });

  it('shows no overflow chip when everything fits', () => {
    render(<TodoTagRow labels={['a', 'b']} index={index} max={5} />);
    expect(screen.queryByText(/^\+/)).toBeNull();
  });
});
