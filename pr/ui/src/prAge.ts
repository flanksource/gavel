export const PR_AGE_OPTIONS = [
  { value: 'day', label: 'Under 24 hours' },
  { value: 'week', label: '1–7 days' },
  { value: 'month', label: '7–30 days' },
  { value: 'older', label: '30 days or more' },
  { value: 'unknown', label: 'Unknown update time' },
];

export function prAgeBucket(updatedAt: string, now: number): string {
  const updated = Date.parse(updatedAt);
  if (!Number.isFinite(updated) || updatedAt.startsWith('0001-01-01')) return 'unknown';
  const days = (now - updated) / 86_400_000;
  if (days < 1) return 'day';
  if (days < 7) return 'week';
  if (days < 30) return 'month';
  return 'older';
}
