import type { KvStore } from "./types";

export function memoryKvStore(initial: Record<string, string> = {}): KvStore {
  const map = new Map(Object.entries(initial));
  return {
    async getItem(key) {
      return map.get(key) ?? null;
    },
    async setItem(key, value) {
      map.set(key, value);
    },
  };
}

export function webKvStore(prefix = "hola:"): KvStore {
  return {
    async getItem(key) {
      if (typeof localStorage === "undefined") return null;
      return localStorage.getItem(prefix + key);
    },
    async setItem(key, value) {
      localStorage.setItem(prefix + key, value);
    },
  };
}

export function wrapKvStore(
  storage: {
    getItem(key: string): Promise<string | null> | string | null;
    setItem(key: string, value: string): Promise<void> | void;
  },
  prefix = "hola:",
): KvStore {
  return {
    async getItem(key) {
      return storage.getItem(prefix + key);
    },
    async setItem(key, value) {
      await storage.setItem(prefix + key, value);
    },
  };
}
