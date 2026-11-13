// The subset of the Chrome extension API the grabber extension uses.
declare namespace chrome {
  type Listener<T extends unknown[]> = { addListener(callback: (...args: T) => void): void };
  namespace runtime {
    /** `tab` and `frameId` are set when the sender is a content script or an extension frame inside a tab. */
    type MessageSender = { tab?: tabs.Tab; frameId?: number; url?: string };
    function getURL(path: string): string;
    function sendMessage(message: unknown): Promise<unknown>;
    /** Returning `true` keeps the channel open for an asynchronous `sendResponse`. */
    const onMessage: {
      addListener(callback: (message: unknown, sender: MessageSender, sendResponse: (response: unknown) => void) => boolean | void): void;
    };
  }
  namespace tabs {
    type Tab = { id?: number; windowId?: number; url?: string; active?: boolean };
    function query(query: { active?: boolean; currentWindow?: boolean }): Promise<Tab[]>;
    function create(options: { url: string }): Promise<Tab>;
    /** A data URL of the visible area of the window's active tab; needs the host permission for that tab's origin. */
    function captureVisibleTab(windowId: number, options: { format: "png" }): Promise<string>;
    const onRemoved: Listener<[tabId: number]>;
  }
  namespace alarms {
    function create(name: string, info: { periodInMinutes: number }): Promise<void>;
    const onAlarm: Listener<[alarm: { name: string }]>;
  }
  namespace action {
    function setBadgeText(details: { tabId: number; text: string }): Promise<void>;
    function setBadgeBackgroundColor(details: { tabId: number; color: string }): Promise<void>;
    function setTitle(details: { tabId: number; title: string }): Promise<void>;
  }
  namespace scripting {
    type Target = { tabId: number };
    function executeScript(injection: { target: Target; files: string[]; world?: "ISOLATED" | "MAIN" }): Promise<{ result?: unknown }[]>;
    function executeScript<Args extends unknown[], Result>(injection: { target: Target; func: (...args: Args) => Result; args: Args; world?: "ISOLATED" | "MAIN" }): Promise<{ result?: Awaited<Result> }[]>;
  }
  namespace permissions {
    function contains(permissions: { origins: string[] }): Promise<boolean>;
    function request(permissions: { origins: string[] }): Promise<boolean>;
  }
  namespace storage {
    type Area = { get(key: string): Promise<Record<string, unknown>>; set(items: Record<string, unknown>): Promise<void>; remove(key: string): Promise<void> };
    const local: Area;
    const session: Area;
  }
}
