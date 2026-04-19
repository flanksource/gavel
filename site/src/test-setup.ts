import "@testing-library/jest-dom/vitest";

// Node 22's experimental localStorage (enabled when --experimental-webstorage or the
// `storage` built-in is loaded) shadows jsdom's implementation but is read-only. Replace
// window.localStorage with a simple in-memory Storage before tests touch it.
function createMemoryStorage(): Storage {
  const data = new Map<string, string>();
  return {
    get length() {
      return data.size;
    },
    key(i) {
      return Array.from(data.keys())[i] ?? null;
    },
    getItem(k) {
      return data.has(k) ? data.get(k)! : null;
    },
    setItem(k, v) {
      data.set(k, String(v));
    },
    removeItem(k) {
      data.delete(k);
    },
    clear() {
      data.clear();
    },
  } satisfies Storage;
}

Object.defineProperty(window, "localStorage", {
  configurable: true,
  value: createMemoryStorage(),
});
