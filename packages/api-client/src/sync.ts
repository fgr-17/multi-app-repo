import { fetchHello, pingHealth, putHello } from "./client";
import { isNewer } from "./lww";
import type {
  GreetingSnapshot,
  KvStore,
  LocalGreeting,
  SyncStatus,
} from "./types";

const RECORD_KEY = "greeting";
const DEVICE_KEY = "deviceId";

export type GreetingSyncOptions = {
  apiUrl: string;
  store: KvStore;
  pollMs?: number;
};

export class GreetingSync {
  private readonly apiUrl: string;
  private readonly store: KvStore;
  private readonly pollMs: number;
  private listeners = new Set<(snap: GreetingSnapshot) => void>();
  private timer: ReturnType<typeof setInterval> | null = null;
  private onlineUnsub: (() => void) | null = null;
  private inflight: Promise<void> | null = null;
  private online = false;
  private error: string | null = null;
  private syncing = false;
  private greeting: LocalGreeting | null = null;
  deviceId = "";

  constructor(opts: GreetingSyncOptions) {
    this.apiUrl = opts.apiUrl;
    this.store = opts.store;
    this.pollMs = opts.pollMs ?? 4000;
  }

  subscribe(listener: (snap: GreetingSnapshot) => void): () => void {
    this.listeners.add(listener);
    listener(this.snapshot());
    return () => {
      this.listeners.delete(listener);
    };
  }

  snapshot(): GreetingSnapshot {
    return {
      greeting: this.greeting,
      online: this.online,
      status: this.status(),
      error: this.error,
    };
  }

  async start(): Promise<void> {
    this.deviceId = await this.ensureDeviceId();
    this.greeting = await this.readLocal();
    this.emit();
    this.onlineUnsub = watchConnectivity(() => {
      void this.reconcile();
    });
    await this.reconcile();
    this.timer = setInterval(() => {
      void this.reconcile();
    }, this.pollMs);
  }

  stop(): void {
    if (this.timer) clearInterval(this.timer);
    this.timer = null;
    this.onlineUnsub?.();
    this.onlineUnsub = null;
  }

  async setName(name: string): Promise<void> {
    const trimmed = name.trim();
    if (!trimmed) return;
    if (!this.deviceId) {
      this.deviceId = await this.ensureDeviceId();
    }
    const next: LocalGreeting = {
      name: trimmed,
      updatedAt: new Date().toISOString(),
      updatedBy: this.deviceId,
      version: (this.greeting?.version ?? 0) + 1,
      dirty: true,
    };
    this.greeting = next;
    await this.writeLocal(next);
    this.error = null;
    this.emit();
    await this.reconcile();
  }

  async reconcile(): Promise<void> {
    if (this.inflight) return this.inflight;
    this.inflight = this.reconcileOnce().finally(() => {
      this.inflight = null;
    });
    return this.inflight;
  }

  private async reconcileOnce(): Promise<void> {
    this.syncing = true;
    this.emit();
    const reachable = await pingHealth(this.apiUrl);
    this.online = reachable;
    if (!reachable) {
      this.syncing = false;
      this.emit();
      return;
    }

    try {
      const remote = await fetchHello(this.apiUrl);
      const local = this.greeting ?? (await this.readLocal());

      if (!local) {
        const saved: LocalGreeting = { ...remote, dirty: false };
        this.greeting = saved;
        await this.writeLocal(saved);
        this.error = null;
        return;
      }

      if (local.dirty && isNewer(local, remote)) {
        const { record } = await putHello(
          {
            name: local.name,
            updatedAt: local.updatedAt,
            updatedBy: local.updatedBy,
          },
          this.apiUrl,
        );
        // 200 = our write landed. 409 = another device won LWW; adopt server.
        const saved: LocalGreeting = { ...record, dirty: false };
        this.greeting = saved;
        await this.writeLocal(saved);
        this.error = null;
        return;
      }

      const saved: LocalGreeting = { ...remote, dirty: false };
      this.greeting = saved;
      await this.writeLocal(saved);
      this.error = null;
    } catch (err) {
      this.online = false;
      this.error = err instanceof Error ? err.message : String(err);
    } finally {
      this.syncing = false;
      this.emit();
    }
  }

  private status(): SyncStatus {
    if (this.syncing) return "syncing";
    if (this.greeting?.dirty) return this.online ? "pending" : "offline";
    if (!this.online) return "offline";
    if (this.error) return "error";
    return "synced";
  }

  private emit(): void {
    const snap = this.snapshot();
    for (const listener of this.listeners) listener(snap);
  }

  private async ensureDeviceId(): Promise<string> {
    const existing = await this.store.getItem(DEVICE_KEY);
    if (existing) return existing;
    const id =
      globalThis.crypto?.randomUUID?.() ??
      `dev-${Date.now()}-${Math.random().toString(16).slice(2)}`;
    await this.store.setItem(DEVICE_KEY, id);
    return id;
  }

  private async readLocal(): Promise<LocalGreeting | null> {
    const raw = await this.store.getItem(RECORD_KEY);
    if (!raw) return null;
    try {
      return JSON.parse(raw) as LocalGreeting;
    } catch {
      return null;
    }
  }

  private async writeLocal(record: LocalGreeting): Promise<void> {
    await this.store.setItem(RECORD_KEY, JSON.stringify(record));
  }
}

function watchConnectivity(onChange: () => void): () => void {
  if (typeof window === "undefined" || !window.addEventListener) {
    return () => {};
  }
  const handler = () => onChange();
  window.addEventListener("online", handler);
  window.addEventListener("offline", handler);
  return () => {
    window.removeEventListener("online", handler);
    window.removeEventListener("offline", handler);
  };
}
