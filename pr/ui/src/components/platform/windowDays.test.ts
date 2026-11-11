// Tests the days of a dashboard window: each UTC day from days days before now through now's day.
import { expect, it } from 'vitest';
import { windowDays } from './windowDays';

it('lists each UTC day of the window, oldest first, through now\'s day', () => {
  expect(windowDays(2, new Date('2026-11-10T23:30:00-02:00'))).toEqual(['2026-11-09', '2026-11-10', '2026-11-11']);
});
