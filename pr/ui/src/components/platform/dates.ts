// How the platform dashboard writes dates: day first, as 31/12/2026, moments in the reader's time
// zone with a 24-hour time, and calendar days (YYYY-MM-DD) as they are.

const two = (n: number) => String(n).padStart(2, '0');

// formatDate is the day a moment falls on, as 31/12/2026.
export function formatDate(at: string | Date): string {
  const date = new Date(at);
  return `${two(date.getDate())}/${two(date.getMonth() + 1)}/${date.getFullYear()}`;
}

// formatDateTime is a moment as 31/12/2026 14:05.
export function formatDateTime(at: string | Date): string {
  const date = new Date(at);
  return `${formatDate(date)} ${two(date.getHours())}:${two(date.getMinutes())}`;
}

// formatDay is a calendar day, YYYY-MM-DD, as 31/12/2026.
export function formatDay(day: string): string {
  const [year, month, date] = day.split('-');
  return `${date}/${month}/${year}`;
}

// formatDayShort is a calendar day without its year, as 31/12, for chart axes.
export function formatDayShort(day: string): string {
  return formatDay(day).slice(0, 5);
}
