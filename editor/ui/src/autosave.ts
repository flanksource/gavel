type Options = {
  source: string;
  revision: string;
  write: (source: string, revision: string) => Promise<string>;
  changed: () => void;
};
export function createAutosave(options: Options) {
  let source = options.source,
    saved = source,
    revision = options.revision;
  let status: "saved" | "pending" | "saving" | "error" = "saved",
    error = "";
  let timer: ReturnType<typeof setTimeout> | undefined;
  let saving: Promise<void> | undefined;
  const changed = () => options.changed();
  const flush = (): Promise<void> => {
    clearTimeout(timer);
    if (saving) return saving;
    if (status === "error") return Promise.reject(new Error(error));
    if (source === saved) return Promise.resolve();
    saving = (async () => {
      status = "saving";
      changed();
      try {
        while (source !== saved) {
          const submitted = source;
          revision = await options.write(submitted, revision);
          saved = submitted;
        }
        status = "saved";
        error = "";
      } catch (cause) {
        status = "error";
        error = cause instanceof Error ? cause.message : String(cause);
        throw cause;
      } finally {
        saving = undefined;
        changed();
      }
    })();
    return saving;
  };
  return {
    state: () => ({ source, revision, status, error }),
    edit: (next: string) => {
      if (source === next) return;
      source = next;
      if (status !== "error") {
        status = !saving && source === saved ? "saved" : "pending";
        clearTimeout(timer);
        if (status !== "saved")
          timer = setTimeout(() => {
            void flush().catch(() => {
              /* Error remains visible in state; retry is explicit. */
            });
          }, 600);
      }
      changed();
    },
    flush,
    retry: () => {
      status = "pending";
      error = "";
      return flush();
    },
    dispose: () => clearTimeout(timer),
  };
}
export type Autosave = ReturnType<typeof createAutosave>;
