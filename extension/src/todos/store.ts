import type { QueuedTodo } from "./model";

export class TodoStore {
  private readonly database: Promise<IDBDatabase>;

  constructor(factory: IDBFactory = indexedDB) {
    this.database = new Promise((resolve, reject) => {
      const request = factory.open("gavel-todo-capture", 1);
      request.onupgradeneeded = () => {
        request.result.createObjectStore("queue", { keyPath: "id" });
        request.result.createObjectStore("cache");
      };
      request.onsuccess = () => resolve(request.result);
      request.onerror = () => reject(request.error);
      request.onblocked = () => reject(new Error("Todo storage upgrade is blocked by another extension page"));
    });
  }

  private async transaction<T>(store: "queue" | "cache", mode: IDBTransactionMode, work: (store: IDBObjectStore) => IDBRequest<T>): Promise<T> {
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
  }

  async queue(server?: string): Promise<QueuedTodo[]> {
    const records = await this.transaction("queue", "readonly", store => store.getAll()) as QueuedTodo[];
    return records.filter(record => !server || record.server === server).sort((a, b) => a.createdAt - b.createdAt);
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
