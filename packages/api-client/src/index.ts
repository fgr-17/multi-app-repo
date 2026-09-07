import type { SyncStatus } from "./types";

export type {
  GreetingRecord,
  GreetingSnapshot,
  KvStore,
  LocalGreeting,
  SyncStatus,
  User,
} from "./types";
export { isNewer } from "./lww";
export { memoryKvStore, webKvStore, wrapKvStore } from "./store";
export {
  apiRoot,
  defaultApiUrl,
  fetchHello,
  fetchUsers,
  pingHealth,
  putHello,
} from "./client";
export { GreetingSync } from "./sync";
export type { GreetingSyncOptions } from "./sync";

export function statusLabel(status: SyncStatus): string {
  switch (status) {
    case "offline":
      return "sin conexión · cambios locales";
    case "syncing":
      return "reconciliando…";
    case "pending":
      return "pendiente de subir";
    case "synced":
      return "sincronizado con postgres";
    case "error":
      return "error de sync";
  }
}
