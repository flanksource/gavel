export type FileEntry = {
  id: string;
  path: string;
  kind: "document" | "image";
};
export type Document = FileEntry & { source: string; revision: string };

/** The edited folders and their files, as the folder-changing routes answer. */
export type OpenedFiles = { roots: string[]; files: FileEntry[] };
/** A registered project the editor can switch to (GET /api/edit/projects). */
export type EditProject = { name: string; dir: string };
/** Which files the folders list; an empty extensions list means the built-in set. */
export type ListingConfig = { extensions?: string[]; includeIgnored?: string[] };
/** GET /api/edit/options: the listing in effect, the folder's .gavel.yaml and flag values, and the built-in extensions. */
export type EditListing = ListingConfig & { configured: ListingConfig; flags: ListingConfig; dir: string; defaults: string[] };

/** A non-2xx API answer, keeping the status and raw body for callers that branch on them. */
export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly body: string,
  ) {
    super(`${status}: ${body}`);
    if (status === 403 && this.json<{ code?: string }>()?.code === "workspace_access_denied") {
      this.message = "Workspace access is required for this folder. Your draft is preserved; restore access before saving.";
    }
  }

  /** The body as JSON, or undefined when it is not JSON (plain http.Error text). */
  json<T>(): T | undefined {
    try {
      return JSON.parse(this.body) as T;
    } catch {
      return undefined;
    }
  }
}
export async function request<T>(
  path: string,
  body?: unknown,
  method = "POST",
): Promise<T> {
  const response = await fetch(
    path,
    body === undefined
      ? {}
      : {
          method,
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(body),
        },
  );
  if (!response.ok) throw new ApiError(response.status, await response.text());
  return response.json() as Promise<T>;
}
