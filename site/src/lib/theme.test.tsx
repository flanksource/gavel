import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useTheme } from "./theme";

const matchMedia = (matches: boolean) =>
  ({
    matches,
    media: "(prefers-color-scheme: dark)",
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    addListener: vi.fn(),
    removeListener: vi.fn(),
    dispatchEvent: vi.fn(),
    onchange: null,
  }) as unknown as MediaQueryList;

describe("useTheme", () => {
  beforeEach(() => {
    window.localStorage.removeItem("gavel-theme");
    document.documentElement.classList.remove("dark");
    document.documentElement.removeAttribute("data-theme");
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("falls back to the system preference when no saved theme exists", () => {
    vi.spyOn(window, "matchMedia").mockImplementation(() => matchMedia(true));
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe("dark");
    expect(document.documentElement.classList.contains("dark")).toBe(true);
  });

  it("restores the persisted theme over the system preference", () => {
    window.localStorage.setItem("gavel-theme", "light");
    vi.spyOn(window, "matchMedia").mockImplementation(() => matchMedia(true));
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe("light");
    expect(document.documentElement.classList.contains("dark")).toBe(false);
  });

  it("persists the theme and toggles between light and dark", () => {
    vi.spyOn(window, "matchMedia").mockImplementation(() => matchMedia(false));
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe("light");

    act(() => result.current.toggle());
    expect(result.current.theme).toBe("dark");
    expect(window.localStorage.getItem("gavel-theme")).toBe("dark");
    expect(document.documentElement.classList.contains("dark")).toBe(true);

    act(() => result.current.toggle());
    expect(result.current.theme).toBe("light");
    expect(window.localStorage.getItem("gavel-theme")).toBe("light");
  });
});
