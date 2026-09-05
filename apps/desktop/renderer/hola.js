"use strict";
var hola = (() => {
  var __defProp = Object.defineProperty;
  var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
  var __getOwnPropNames = Object.getOwnPropertyNames;
  var __hasOwnProp = Object.prototype.hasOwnProperty;
  var __export = (target, all) => {
    for (var name in all)
      __defProp(target, name, { get: all[name], enumerable: true });
  };
  var __copyProps = (to, from, except, desc) => {
    if (from && typeof from === "object" || typeof from === "function") {
      for (let key of __getOwnPropNames(from))
        if (!__hasOwnProp.call(to, key) && key !== except)
          __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
    }
    return to;
  };
  var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

  // src/index.ts
  var index_exports = {};
  __export(index_exports, {
    GreetingSync: () => GreetingSync,
    apiRoot: () => apiRoot,
    defaultApiUrl: () => defaultApiUrl,
    fetchHello: () => fetchHello,
    isNewer: () => isNewer,
    memoryKvStore: () => memoryKvStore,
    pingHealth: () => pingHealth,
    putHello: () => putHello,
    statusLabel: () => statusLabel,
    webKvStore: () => webKvStore,
    wrapKvStore: () => wrapKvStore
  });

  // src/lww.ts
  function isNewer(incoming, current) {
    const incomingAt = Date.parse(incoming.updatedAt);
    const currentAt = Date.parse(current.updatedAt);
    if (incomingAt !== currentAt) {
      return incomingAt > currentAt;
    }
    if (incoming.updatedBy !== current.updatedBy) {
      return incoming.updatedBy > current.updatedBy;
    }
    return incoming.version > current.version;
  }

  // src/store.ts
  function memoryKvStore(initial = {}) {
    const map = new Map(Object.entries(initial));
    return {
      async getItem(key) {
        return map.get(key) ?? null;
      },
      async setItem(key, value) {
        map.set(key, value);
      }
    };
  }
  function webKvStore(prefix = "hola:") {
    return {
      async getItem(key) {
        if (typeof localStorage === "undefined") return null;
        return localStorage.getItem(prefix + key);
      },
      async setItem(key, value) {
        localStorage.setItem(prefix + key, value);
      }
    };
  }
  function wrapKvStore(storage, prefix = "hola:") {
    return {
      async getItem(key) {
        return storage.getItem(prefix + key);
      },
      async setItem(key, value) {
        await storage.setItem(prefix + key, value);
      }
    };
  }

  // src/client.ts
  function defaultApiUrl() {
    return typeof process !== "undefined" && process.env.EXPO_PUBLIC_API_URL || typeof process !== "undefined" && process.env.NEXT_PUBLIC_API_URL || typeof process !== "undefined" && process.env.API_URL || "http://localhost:8080";
  }
  function apiRoot(apiBaseUrl) {
    return apiBaseUrl.replace(/\/$/, "");
  }
  async function fetchHello(apiBaseUrl = defaultApiUrl()) {
    const url = `${apiRoot(apiBaseUrl)}/api/hello`;
    const res = await fetch(url);
    if (!res.ok) {
      throw new Error(`API ${res.status} al pedir ${url}`);
    }
    return await res.json();
  }
  async function putHello(record, apiBaseUrl = defaultApiUrl()) {
    const url = `${apiRoot(apiBaseUrl)}/api/hello`;
    const res = await fetch(url, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(record)
    });
    if (!res.ok && res.status !== 409) {
      throw new Error(`API ${res.status} al guardar ${url}`);
    }
    return {
      record: await res.json(),
      accepted: res.status === 200
    };
  }
  async function pingHealth(apiBaseUrl = defaultApiUrl()) {
    const ctrl = new AbortController();
    const timer = setTimeout(() => ctrl.abort(), 2500);
    try {
      const res = await fetch(`${apiRoot(apiBaseUrl)}/health`, { signal: ctrl.signal });
      return res.ok;
    } catch {
      return false;
    } finally {
      clearTimeout(timer);
    }
  }

  // src/sync.ts
  var RECORD_KEY = "greeting";
  var DEVICE_KEY = "deviceId";
  var GreetingSync = class {
    apiUrl;
    store;
    pollMs;
    listeners = /* @__PURE__ */ new Set();
    timer = null;
    onlineUnsub = null;
    inflight = null;
    online = false;
    error = null;
    syncing = false;
    greeting = null;
    deviceId = "";
    constructor(opts) {
      this.apiUrl = opts.apiUrl;
      this.store = opts.store;
      this.pollMs = opts.pollMs ?? 4e3;
    }
    subscribe(listener) {
      this.listeners.add(listener);
      listener(this.snapshot());
      return () => {
        this.listeners.delete(listener);
      };
    }
    snapshot() {
      return {
        greeting: this.greeting,
        online: this.online,
        status: this.status(),
        error: this.error
      };
    }
    async start() {
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
    stop() {
      if (this.timer) clearInterval(this.timer);
      this.timer = null;
      this.onlineUnsub?.();
      this.onlineUnsub = null;
    }
    async setName(name) {
      const trimmed = name.trim();
      if (!trimmed) return;
      if (!this.deviceId) {
        this.deviceId = await this.ensureDeviceId();
      }
      const next = {
        name: trimmed,
        updatedAt: (/* @__PURE__ */ new Date()).toISOString(),
        updatedBy: this.deviceId,
        version: (this.greeting?.version ?? 0) + 1,
        dirty: true
      };
      this.greeting = next;
      await this.writeLocal(next);
      this.error = null;
      this.emit();
      await this.reconcile();
    }
    async reconcile() {
      if (this.inflight) return this.inflight;
      this.inflight = this.reconcileOnce().finally(() => {
        this.inflight = null;
      });
      return this.inflight;
    }
    async reconcileOnce() {
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
        const local = this.greeting ?? await this.readLocal();
        if (!local) {
          const saved2 = { ...remote, dirty: false };
          this.greeting = saved2;
          await this.writeLocal(saved2);
          this.error = null;
          return;
        }
        if (local.dirty && isNewer(local, remote)) {
          const { record } = await putHello(
            {
              name: local.name,
              updatedAt: local.updatedAt,
              updatedBy: local.updatedBy
            },
            this.apiUrl
          );
          const saved2 = { ...record, dirty: false };
          this.greeting = saved2;
          await this.writeLocal(saved2);
          this.error = null;
          return;
        }
        const saved = { ...remote, dirty: false };
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
    status() {
      if (this.syncing) return "syncing";
      if (this.greeting?.dirty) return this.online ? "pending" : "offline";
      if (!this.online) return "offline";
      if (this.error) return "error";
      return "synced";
    }
    emit() {
      const snap = this.snapshot();
      for (const listener of this.listeners) listener(snap);
    }
    async ensureDeviceId() {
      const existing = await this.store.getItem(DEVICE_KEY);
      if (existing) return existing;
      const id = globalThis.crypto?.randomUUID?.() ?? `dev-${Date.now()}-${Math.random().toString(16).slice(2)}`;
      await this.store.setItem(DEVICE_KEY, id);
      return id;
    }
    async readLocal() {
      const raw = await this.store.getItem(RECORD_KEY);
      if (!raw) return null;
      try {
        return JSON.parse(raw);
      } catch {
        return null;
      }
    }
    async writeLocal(record) {
      await this.store.setItem(RECORD_KEY, JSON.stringify(record));
    }
  };
  function watchConnectivity(onChange) {
    if (typeof window === "undefined" || !window.addEventListener) {
      return () => {
      };
    }
    const handler = () => onChange();
    window.addEventListener("online", handler);
    window.addEventListener("offline", handler);
    return () => {
      window.removeEventListener("online", handler);
      window.removeEventListener("offline", handler);
    };
  }

  // src/index.ts
  function statusLabel(status) {
    switch (status) {
      case "offline":
        return "sin conexi\xF3n \xB7 cambios locales";
      case "syncing":
        return "reconciliando\u2026";
      case "pending":
        return "pendiente de subir";
      case "synced":
        return "sincronizado con postgres";
      case "error":
        return "error de sync";
    }
  }
  return __toCommonJS(index_exports);
})();
