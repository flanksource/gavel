// Tests how the platform dashboard writes dates: day first, as 31/12/2026, with a 24-hour time
// for moments, in the reader's time zone, and calendar days as they are, whatever the time zone.
import { expect, it } from 'vitest';
import { formatDate, formatDateTime, formatDay, formatDayShort } from './dates';

it('writes a moment\'s date day first', () => {
  expect(formatDate(new Date(2026, 11, 31, 23, 59))).toBe('31/12/2026');
  expect(formatDate(new Date(2026, 0, 5, 0, 0).toISOString()), 'from an ISO timestamp too').toBe('05/01/2026');
});

it('writes a moment with its 24-hour time', () => {
  expect(formatDateTime(new Date(2026, 11, 31, 14, 5))).toBe('31/12/2026 14:05');
  expect(formatDateTime(new Date(2026, 0, 5, 0, 7).toISOString())).toBe('05/01/2026 00:07');
});

it('writes a calendar day day first, without moving it to the reader\'s time zone', () => {
  expect(formatDay('2026-12-31')).toBe('31/12/2026');
  expect(formatDayShort('2026-01-05'), 'chart axes leave the year out').toBe('05/01');
});
