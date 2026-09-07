"use client";

import { fetchUsers, GreetingSync, statusLabel, webKvStore } from "@hola/api-client";
import type { GreetingSnapshot, User } from "@hola/api-client";
import { FormEvent, useEffect, useMemo, useState } from "react";

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
  const [users, setUsers] = useState<User[]>([]);
  const [usersError, setUsersError] = useState<string | null>(null);

  useEffect(() => {
    const unsub = sync.subscribe(setSnap);
    void sync.start();
    return () => {
      unsub();
      sync.stop();
    };
  }, [sync]);

  useEffect(() => {
    fetchUsers(apiUrl)
      .then(setUsers)
      .catch((err: Error) => setUsersError(err.message));
  }, []);

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    await sync.setName(draft || snap.greeting?.name || "");
    setDraft("");
  }

  const name = snap.greeting?.name;

  return (
    <main className="flex min-h-screen flex-col justify-center gap-16 px-12 py-16 sm:px-20">
      <section>
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
      </section>

      <section>
        <p className="mb-3 font-sans text-xs uppercase tracking-[0.14em] text-muted">
          Usuarios · Postgres relacional
        </p>
        <p className="mb-6 max-w-xl font-sans text-sm text-muted">
          Identidad en tablas (id, email, nombre). Tocá uno para saludarlo; el
          saludo sigue yendo por eventos + Kafka + Mongo.
        </p>
        {usersError ? (
          <p className="font-sans text-sm text-muted">{usersError}</p>
        ) : (
          <ul className="grid max-w-3xl gap-3 sm:grid-cols-2">
            {users.map((u) => (
              <li key={u.id}>
                <button
                  type="button"
                  onClick={() => void sync.setName(u.givenName)}
                  className="w-full rounded-md border border-[#d9d0c3] bg-white px-4 py-3 text-left"
                >
                  <span className="block font-sans text-base text-foreground">
                    {u.displayName}
                  </span>
                  <span className="mt-1 block font-sans text-xs text-muted">
                    {u.email} · {u.city}, {u.country}
                  </span>
                </button>
              </li>
            ))}
          </ul>
        )}
      </section>
    </main>
  );
}
