export type FileEntry = {
  id: string;
  path: string;
  kind: "document" | "image";
};
export type Document = FileEntry & { source: string; revision: string };

/** A non-2xx API answer, keeping the status and raw body for callers that branch on them. */
export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly body: string,
  ) {
    super(`${status}: ${body}`);
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
