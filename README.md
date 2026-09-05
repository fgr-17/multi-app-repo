# Hola ecosystem

Monorepo mínimo de un producto que corre en **web**, **iOS/Android** y **desktop** (Windows, Linux, macOS). Los tres clientes piden el nombre a un **API REST en Go** y muestran `Hola {nombre}`.

```
                    ┌─────────────┐
   Next.js  web ───►│             │
   Expo   mobile ──►│  Go /api    │── { "name": "Mundo" }
   Electron desk ──►│  GET /hello │
                    └─────────────┘
```

## Estructura

```
apps/
  api/        backend Go (GET /api/hello, GET /health)
  web/        Next.js
  mobile/     Expo → iOS y Android
  desktop/    Electron → Windows / Linux / macOS
packages/
  api-client/ cliente TypeScript compartido
```

El contrato es deliberadamente chico: un JSON `{ "name": "…" }`. Web, mobile y desktop no comparten UI (React DOM ≠ React Native), pero sí el cliente HTTP y el tipo.

## Requisitos

- Go 1.22+
- Node 22+ y pnpm
- Para mobile: [Expo Go](https://expo.dev/go) o Xcode / Android Studio
- Para desktop: las binarias de Electron se bajan con `pnpm install`

## Cómo correrlo

```bash
pnpm install
```

Tres terminales (o `pnpm dev` para levantar todo lo que Turbo pueda en paralelo):

```bash
pnpm dev:api        # http://localhost:8080
pnpm dev:web        # http://localhost:3000  →  Hola Mundo
pnpm dev:desktop    # ventana nativa
pnpm dev:mobile     # QR de Expo
```

El nombre lo define el backend:

```bash
GREETING_NAME=Ada pnpm dev:api
```

También podés levantar solo el API con Docker:

```bash
docker compose up --build api
```

### Mobile en un device o emulador

| Dónde corre la app | `EXPO_PUBLIC_API_URL` |
| --- | --- |
| iOS Simulator | `http://localhost:8080` (default) |
| Android Emulator | `http://10.0.2.2:8080` (default en Android) |
| Teléfono físico | `http://<IP-LAN-de-tu-máquina>:8080` |

## API

```http
GET /api/hello
```

```json
{ "name": "Mundo" }
```

```http
GET /health
```

```
ok
```

CORS está abierto a `*` para que localhost, Expo y Electron puedan pegarle sin setup extra. En producción cerralo.

## Por qué este stack

- **Un API en Go**: un binario, un puerto, fácil de dockerizar. No hay DB: el nombre sale de `GREETING_NAME`.
- **Next.js en web**: App Router + fetch al REST. El HTML vive en el browser; el dato lo sirve Go.
- **Expo en mobile**: un solo codebase para iOS y Android. Metro está configurado para resolver `@hola/api-client` desde el monorepo.
- **Electron en desktop**: la misma idea que web, empaquetada en una ventana nativa para Win/Lin/Mac. Si más adelante querés un binario más liviano, el renderer se puede mover a Tauri sin tocar el API.
- **pnpm workspaces + Turborepo**: cada app es un paquete. El cliente HTTP se importa como `workspace:*`.

## Scripts

| Comando | Qué hace |
| --- | --- |
| `pnpm dev` | Turbo: api + web + desktop + metro |
| `pnpm dev:api` | solo Go |
| `pnpm --filter @hola/api test` | test del handler |
| `pnpm --filter @hola/web build` | build de Next.js |
