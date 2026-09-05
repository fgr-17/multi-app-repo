import assert from "node:assert/strict";
import { after, before, test } from "node:test";
import { GreetingSync } from "./sync";
import { memoryKvStore } from "./store";
import type { GreetingRecord } from "./types";

const originalFetch = globalThis.fetch;

function iso(plusMs = 0): string {
  return new Date(Date.parse("2026-09-05T12:00:00.000Z") + plusMs).toISOString();
}

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

test("offline edit stays local and pushes when the API returns", async () => {
  let remote: GreetingRecord = {
    name: "Mundo",
    updatedAt: iso(),
    updatedBy: "seed",
    version: 1,
  };
  let online = false;

  globalThis.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
    if (!online) throw new TypeError("network down");
    const url = String(input);
    if (url.endsWith("/health")) return new Response("ok", { status: 200 });
    if (url.endsWith("/api/hello") && (!init || init.method === "GET")) {
      return jsonResponse(remote);
    }
    if (url.endsWith("/api/hello") && init?.method === "PUT") {
      const body = JSON.parse(String(init.body)) as GreetingRecord;
      remote = {
        name: body.name,
        updatedAt: body.updatedAt,
        updatedBy: body.updatedBy,
        version: remote.version + 1,
      };
      return jsonResponse(remote);
    }
    throw new Error(url);
  };

  const sync = new GreetingSync({
    apiUrl: "http://api.test",
    store: memoryKvStore(),
    pollMs: 60_000,
  });
  await sync.start();
  assert.equal(sync.snapshot().online, false);

  await sync.setName("Ada");
  assert.equal(sync.snapshot().greeting?.name, "Ada");
  assert.equal(sync.snapshot().greeting?.dirty, true);
  assert.equal(remote.name, "Mundo");

  online = true;
  await sync.reconcile();
  assert.equal(sync.snapshot().online, true);
  assert.equal(sync.snapshot().greeting?.name, "Ada");
  assert.equal(sync.snapshot().greeting?.dirty, false);
  assert.equal(remote.name, "Ada");

  sync.stop();
});

test("remote newer write wins over a stale local dirty edit", async () => {
  const remote: GreetingRecord = {
    name: "Grace",
    updatedAt: iso(60_000),
    updatedBy: "web",
    version: 4,
  };

  globalThis.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    if (url.endsWith("/health")) return new Response("ok", { status: 200 });
    if (url.endsWith("/api/hello") && (!init || init.method === "GET")) {
      return jsonResponse(remote);
    }
    throw new Error(`unexpected ${init?.method} ${url}`);
  };

  const sync = new GreetingSync({
    apiUrl: "http://api.test",
    store: memoryKvStore({
      greeting: JSON.stringify({
        name: "Ada",
        updatedAt: iso(),
        updatedBy: "mobile",
        version: 2,
        dirty: true,
      }),
    }),
    pollMs: 60_000,
  });
  await sync.start();
  assert.equal(sync.snapshot().greeting?.name, "Grace");
  assert.equal(sync.snapshot().greeting?.dirty, false);
  sync.stop();
});

before(() => {});
after(() => {
  globalThis.fetch = originalFetch;
});
