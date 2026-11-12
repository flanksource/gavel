// Tests the platform dashboard's section sidebar: one entry per section, a click scrolls to its
// section, and the entry of the section being read stays highlighted while scrolling.
import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { PlatformSections } from './PlatformSections';

const sections = [
  { id: 'tenants', label: 'Tenants', content: <p>tenant rows</p> },
  { id: 'budgets', label: 'Budgets', content: <p>budget rows</p> },
  { id: 'infrastructure', label: 'Infrastructure', content: <p>infrastructure rows</p> },
];

let scrolledTo: string[];

beforeEach(() => {
  scrolledTo = [];
  Element.prototype.scrollIntoView = vi.fn(function (this: Element) { scrolledTo.push(this.id); });
});

afterEach(() => {
  delete (Element.prototype as Partial<Element>).scrollIntoView;
});

function current() {
  return screen.getByRole('navigation', { name: 'Platform sections' }).querySelector('[aria-current="true"]')?.textContent;
}

// placeSections puts each section's top at tops[i] pixels below the scrolling area's top.
function placeSections(tops: number[]) {
  sections.forEach((section, i) => {
    const element = document.getElementById(`platform-${section.id}`)!;
    element.getBoundingClientRect = () => ({ top: tops[i] }) as DOMRect;
  });
}

it('lists every section and highlights the first', () => {
  render(<PlatformSections sections={sections} />);

  const nav = screen.getByRole('navigation', { name: 'Platform sections' });
  expect([...nav.querySelectorAll('button')].map(button => button.textContent)).toEqual(['Tenants', 'Budgets', 'Infrastructure']);
  expect(current()).toBe('Tenants');
  expect(screen.getByText('budget rows')).toBeTruthy();
});

it('scrolls to a section when its entry is clicked, and highlights it', () => {
  render(<PlatformSections sections={sections} />);

  fireEvent.click(screen.getByRole('button', { name: 'Budgets' }));
  expect(scrolledTo).toEqual(['platform-budgets']);
  expect(current()).toBe('Budgets');
});

it('highlights the section being read while scrolling', () => {
  render(<PlatformSections sections={sections} />);
  const area = screen.getByTestId('platform-sections-scroll');

  placeSections([-900, -10, 400]);
  fireEvent.scroll(area);
  expect(current()).toBe('Budgets');

  placeSections([0, 600, 1200]);
  fireEvent.scroll(area);
  expect(current()).toBe('Tenants');
});

it('keeps a clicked section highlighted at the bottom, where it can\'t scroll to the top', () => {
  render(<PlatformSections sections={sections} />);
  const area = screen.getByTestId('platform-sections-scroll');
  Object.defineProperties(area, { scrollTop: { value: 600 }, clientHeight: { value: 400 }, scrollHeight: { value: 1000 } });

  fireEvent.click(screen.getByRole('button', { name: 'Budgets' }));
  placeSections([-900, 50, 300]);
  fireEvent.scroll(area);
  expect(current(), 'Budgets stops short of the top, yet it is what was asked for').toBe('Budgets');

  fireEvent.wheel(area);
  fireEvent.scroll(area);
  expect(current(), 'scrolling by hand to the bottom highlights the last section').toBe('Infrastructure');
});

it('highlights the first section when the sections arrive after the first render', () => {
  const { rerender } = render(<PlatformSections sections={[]} />);
  rerender(<PlatformSections sections={sections} />);

  expect(current()).toBe('Tenants');
});
