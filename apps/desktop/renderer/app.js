const greetingEl = document.getElementById("greeting");
const statusEl = document.getElementById("status");
const form = document.getElementById("form");
const input = document.getElementById("name");
const apiUrl = (window.holaEnv && window.holaEnv.apiUrl) || "http://localhost:8080";

const sync = new window.hola.GreetingSync({
  apiUrl,
  store: window.hola.webKvStore("hola-desktop:"),
});

sync.subscribe((snap) => {
  const name = snap.greeting && snap.greeting.name;
  greetingEl.textContent = name ? `Hola ${name}` : "Hola …";
  const label = window.hola.statusLabel(snap.status);
  statusEl.textContent = snap.error ? `${label}: ${snap.error}` : label;
  if (name && !input.value) input.placeholder = name;
});

form.addEventListener("submit", (event) => {
  event.preventDefault();
  const value = input.value.trim();
  const current = sync.snapshot().greeting && sync.snapshot().greeting.name;
  void sync.setName(value || current || "");
  input.value = "";
});

void sync.start();
