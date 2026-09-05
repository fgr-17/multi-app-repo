"use client";

import { GreetingSync, statusLabel, webKvStore } from "@hola/api-client";
import { FormEvent, useEffect, useMemo, useState } from "react";
import type { GreetingSnapshot } from "@hola/api-client";

const apiUrl = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const empty: GreetingSnapshot = {
  greeting: null,
  online: false,
  status: "offline",
  error: null,
};

export default function Home() {
  const store = useMemo(() => webKvStore("hola-web:"), []);
  const sync = useMemo(() => new GreetingSync({ apiUrl, store }), [store]);
  const [snap, setSnap] = useState<GreetingSnapshot>(empty);
  const [draft, setDraft] = useState("");

  useEffect(() => {
    const unsub = sync.subscribe(setSnap);
    void sync.start();
    return () => {
      unsub();
      sync.stop();
    };
  }, [sync]);

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    await sync.setName(draft || snap.greeting?.name || "");
    setDraft("");
  }

  const name = snap.greeting?.name;

  return (
    <main className="flex min-h-screen flex-col justify-center px-12 py-16 sm:px-20">
      <p className="mb-3 font-sans text-xs uppercase tracking-[0.14em] text-muted">
        Web · Next.js
      </p>
      <h1 className="font-serif text-6xl leading-none sm:text-8xl">
        {name ? `Hola ${name}` : "Hola …"}
      </h1>
      <p className="mt-5 max-w-lg font-sans text-sm text-muted">
        {snap.error ? `${statusLabel(snap.status)}: ${snap.error}` : statusLabel(snap.status)}
      </p>
      <form onSubmit={onSubmit} className="mt-8 flex max-w-md flex-col gap-3 sm:flex-row">
        <input
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          placeholder={name ?? "nombre"}
          className="min-w-0 flex-1 rounded-md border border-[#d9d0c3] bg-white px-3 py-2 font-sans text-base text-foreground outline-none focus:border-[#1c1916]"
        />
        <button
          type="submit"
          className="rounded-md bg-[#1c1916] px-4 py-2 font-sans text-sm text-[#f4efe6]"
        >
          Guardar
        </button>
      </form>
    </main>
  );
}
