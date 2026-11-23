import type { QueuedTodo } from "./model";
import type { QueuedComment } from "../comment-sync";

export class TodoStore {
  private readonly database: Promise<IDBDatabase>;

  constructor(factory: IDBFactory = indexedDB, private readonly queueChanged?: () => void) {
    this.database = new Promise((resolve, reject) => {
      const request = factory.open("gavel-todo-capture", 2);
      request.onupgradeneeded = event => {
        if (event.oldVersion < 1) {
          request.result.createObjectStore("queue", { keyPath: "id" });
          request.result.createObjectStore("cache");
        }
        request.result.createObjectStore("comments", { keyPath: "id" }).createIndex("page", ["server", "project", "pageKey"]);
      };
      request.onsuccess = () => {
        request.result.onversionchange = () => request.result.close();
        resolve(request.result);
      };
      request.onerror = () => reject(request.error);
      request.onblocked = () => reject(new Error("Todo storage upgrade is blocked by another extension page"));
    });
  }

  private async transaction<T>(store: "queue" | "cache" | "comments", mode: IDBTransactionMode, work: (store: IDBObjectStore) => IDBRequest<T>): Promise<T> {
    const database = await this.database;
    return new Promise((resolve, reject) => {
      const transaction = database.transaction(store, mode);
      const request = work(transaction.objectStore(store));
      transaction.oncomplete = () => resolve(request.result);
      transaction.onerror = () => reject(transaction.error ?? request.error);
      transaction.onabort = () => reject(transaction.error ?? new Error("Todo storage transaction aborted"));
    });
  }

  async put(todo: QueuedTodo): Promise<void> {
    await this.transaction("queue", "readwrite", store => store.put(todo));
    this.queueChanged?.();
  }

  async queue(server?: string): Promise<QueuedTodo[]> {
    const records = await this.transaction("queue", "readonly", store => store.getAll()) as QueuedTodo[];
    return records.filter(record => !server || record.server === server).sort((a, b) => a.createdAt - b.createdAt);
  }

  async putComment(comment: QueuedComment): Promise<void> {
    await this.transaction("comments", "readwrite", store => store.put(comment));
    this.queueChanged?.();
  }

  async removeComment(id: string): Promise<void> {
    await this.transaction("comments", "readwrite", store => store.delete(id));
    this.queueChanged?.();
  }

  async commentQueue(page?: Pick<QueuedComment, "server" | "project" | "pageKey">): Promise<QueuedComment[]> {
    const rows = await this.transaction("comments", "readonly", store => page
      ? store.index("page").getAll([page.server, page.project, page.pageKey])
      : store.getAll()) as QueuedComment[];
    return rows.sort((left, right) => left.savedAt - right.savedAt);
  }

  async cache<T>(key: string): Promise<T | undefined> {
    return this.transaction("cache", "readonly", store => store.get(key));
  }

  async remember<T>(key: string, value: T): Promise<void> {
    await this.transaction("cache", "readwrite", store => store.put(value, key));
  }
}

export function cacheKey(server: string, resource: string, dir = ""): string {
  return JSON.stringify([server, resource, dir]);
}
