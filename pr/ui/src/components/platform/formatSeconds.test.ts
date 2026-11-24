// Tests how the platform dashboard shows a length of time: seconds, minutes or hours.
import { expect, it } from 'vitest';
import { formatSeconds } from './formatSeconds';

it('shows seconds, whole minutes, or hours to one decimal', () => {
  expect([formatSeconds(42.4), formatSeconds(150), formatSeconds(5400)]).toEqual(['42s', '3m', '1.5h']);
});
