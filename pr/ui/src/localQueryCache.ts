// A localStorage copy of a query's last good response, used as react-query
// placeholderData so a cold page load renders the last known data at once and
// swaps in the fresh response when it lands. It is only ever a placeholder: the
// query still fetches, and a request failure still surfaces as the query error.

// readLocalCache returns the cached value when it still passes parse. Storage
// can be unavailable (private mode) and an older build may have written a
// different shape; either way there is simply no placeholder.
export function readLocalCache<T>(key: string, parse: (value: unknown) => T): T | undefined {
  try {
    const raw = localStorage.getItem(key);
    return raw === null ? undefined : parse(JSON.parse(raw));
  } catch {
    return undefined;
  }
}

export function writeLocalCache(key: string, value: unknown): void {
  try {
    localStorage.setItem(key, JSON.stringify(value));
  } catch {
    // best-effort: storage unavailable or full — the next load just waits.
  }
}
