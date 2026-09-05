const greeting = document.getElementById("greeting");
const status = document.getElementById("status");
const apiUrl = (window.hola && window.hola.apiUrl) || "http://localhost:8080";

async function loadHello() {
  try {
    const res = await fetch(`${apiUrl.replace(/\/$/, "")}/api/hello`);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    greeting.textContent = `Hola ${data.name}`;
    status.textContent = `nombre servido por ${apiUrl}/api/hello`;
  } catch (err) {
    greeting.textContent = "Hola ?";
    status.textContent = `no pude hablar con el API: ${err.message}`;
  }
}

loadHello();
