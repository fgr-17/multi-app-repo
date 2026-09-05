"use client";

import { fetchHello } from "@hola/api-client";
import { useEffect, useState } from "react";

const apiUrl = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export default function Home() {
  const [name, setName] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchHello(apiUrl)
      .then((data) => setName(data.name))
      .catch((err: Error) => setError(err.message));
  }, []);

  return (
    <main className="flex min-h-screen flex-col justify-center px-12 py-16 sm:px-20">
      <p className="mb-3 font-sans text-xs uppercase tracking-[0.14em] text-muted">
        Web · Next.js
      </p>
      <h1 className="font-serif text-6xl leading-none sm:text-8xl">
        {name ? `Hola ${name}` : "Hola …"}
      </h1>
      <p className="mt-5 max-w-lg font-sans text-sm text-muted">
        {error
          ? `no pude hablar con el API: ${error}`
          : name
            ? `nombre servido por ${apiUrl}/api/hello`
            : "pidiendo el nombre al API de Go…"}
      </p>
    </main>
  );
}
