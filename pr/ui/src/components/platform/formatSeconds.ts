// Shows a length of time on the platform dashboard, such as sandbox or Captain run time.

// formatSeconds shows seconds as seconds, whole minutes or hours.
export function formatSeconds(seconds: number): string {
  if (seconds < 60) return `${Math.round(seconds)}s`;
  if (seconds < 3600) return `${Math.round(seconds / 60)}m`;
  return `${(seconds / 3600).toFixed(1)}h`;
}
