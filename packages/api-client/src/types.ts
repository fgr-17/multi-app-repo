export type GreetingRecord = {
  name: string;
  updatedAt: string;
  updatedBy: string;
  version: number;
};

export type LocalGreeting = GreetingRecord & {
  dirty: boolean;
};

export type SyncStatus = "offline" | "syncing" | "pending" | "synced" | "error";

export type GreetingSnapshot = {
  greeting: LocalGreeting | null;
  online: boolean;
  status: SyncStatus;
  error: string | null;
};

export type KvStore = {
  getItem(key: string): Promise<string | null>;
  setItem(key: string, value: string): Promise<void>;
};
