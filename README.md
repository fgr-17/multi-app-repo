# Hola ecosystem

Monorepo de un producto que corre en **web**, **iOS/Android** y **desktop**. El nombre a saludar vive en **Postgres**. Cada app lo lee y lo edita **sin conexión**; al volver la red se **reconcilia** con last-write-wins.

```
  Next.js / Expo / Electron
           │  GET/PUT /api/hello
           ▼
        API Go  ──►  PostgreSQL  (una fila: name, updated_at, updated_by, version)
           ▲
     cache local (localStorage / AsyncStorage)
     si no hay red: se edita igual y queda dirty
```

## Estructura

```
apps/
  api/        Go + Postgres (GET/PUT /api/hello)
  web/        Next.js
  mobile/     Expo → iOS y Android
  desktop/    Electron → Windows / Linux / macOS
packages/
  api-client/ fetch + motor de sync offline compartido
```

Web, mobile y desktop no comparten UI (React DOM ≠ React Native), pero sí el cliente y el algoritmo de reconciliación.

## Requisitos

- Go 1.22+
- Node 22+ y pnpm
- PostgreSQL 16 (local o Docker)
- Para mobile: [Expo Go](https://expo.dev/go) o Xcode / Android Studio
- Para desktop: Electron se baja con `pnpm install`

## Cómo correrlo

```bash
pnpm install
pnpm dev:db          # postgres local (o docker compose up -d db)
pnpm dev:api         # http://localhost:8080
pnpm dev:web         # http://localhost:3000
pnpm dev:desktop
pnpm dev:mobile
```

Con Docker, API + DB juntos:

```bash
docker compose up --build
```

### Mobile en un device o emulador

| Dónde corre la app | `EXPO_PUBLIC_API_URL` |
| --- | --- |
| iOS Simulator | `http://localhost:8080` (default) |
| Android Emulator | `http://10.0.2.2:8080` (default en Android) |
| Teléfono físico | `http://<IP-LAN-de-tu-máquina>:8080` |

## Offline y reconciliación

Cada cliente guarda `{ name, updatedAt, updatedBy, version, dirty }` en storage local.

1. **Leer / editar siempre pega primero en local.** La UI no espera al API.
2. Si hay red, hace `GET /api/hello`. Si lo local está `dirty` y es más nuevo, hace `PUT`. Si lo remoto es más nuevo, gana Postgres.
3. Empate de timestamp: gana `updatedBy` (id de dispositivo) y después `version`.
4. Al volver `online` (evento del browser o el próximo poll de 4s) se vuelve a reconciliar.

Probalo: cortá el API, cambiá el nombre en web, levantá el API. El `PUT` sube el cambio a Postgres y las otras apps lo ven en el siguiente poll.

## API

```http
GET /api/hello
PUT /api/hello
Content-Type: application/json

{ "name": "Ada", "updatedAt": "2026-09-05T20:00:00Z", "updatedBy": "device-id" }
```

`PUT` responde `200` si aceptó el write, `409` si el row de Postgres era más nuevo (el body trae el ganador).

```http
GET /health
```

CORS está abierto a `*` para localhost / Expo / Electron. En producción cerralo.

## Por qué este stack

- **Postgres**: una sola fuente de verdad cuando hay red.
- **Go**: REST chico, transacción `SELECT … FOR UPDATE` para que dos PUTs concurrentes no se pisen mal.
- **Sync en `@hola/api-client`**: la misma LWW en web, mobile y desktop.
- **Expo / Electron / Next**: tres shells, un contrato.

## Scripts

| Comando | Qué hace |
| --- | --- |
| `pnpm dev:db` | levanta Postgres y crea usuario/db `hola` |
| `pnpm dev:api` | API Go |
| `pnpm --filter @hola/api test` | tests LWW + HTTP |
| `pnpm --filter @hola/api-client test` | tests de sync offline |
| `pnpm --filter @hola/web build` | build de Next.js |
