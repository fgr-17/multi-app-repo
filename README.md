# Hola ecosystem

Monorepo de un producto que corre en **web**, **iOS/Android** y **desktop**. El nombre vive en **Postgres**. Cada cliente edita **offline** y al volver la red se reconcilia (last-write-wins).

En el host solo hace falta **Docker** y **Make**. Go, Node, pnpm y Postgres corren adentro de los contenedores.

```
make up          # Postgres + API Go + Next.js
# http://localhost:3000  →  Hola {nombre}
# http://localhost:8080/api/hello
```

Elegí Make y no Bazel: Make ya está en Linux/macOS y no suma otra toolchain.

## Comandos

```bash
make help
make doctor          # ¿hay Docker?
make up              # db + api + web
make logs
make test            # tests de API y sync, en contenedores
make psql
make mobile          # Metro/Expo (perfil extra)
make down
make clean           # contenedores + volúmenes
```

`make desktop` no levanta Electron: una ventana nativa no se dockeriza bien en Win/Mac. La UI es la de `make up` (http://localhost:3000).

### Mobile

```bash
make mobile
```

En el teléfono, Expo Go contra el Metro del contenedor. El API tiene que ser alcanzable desde el device:

| Dónde corre la app | URL del API |
| --- | --- |
| Expo en la misma máquina | `http://localhost:8080` |
| Android emulator | `http://10.0.2.2:8080` |
| Teléfono físico | `http://<IP-LAN>:8080` (`make mobile` intenta inyectar `HOST_IP`) |

## Estructura

```
Makefile              único entrypoint
docker-compose.yml
docker/
  api.Dockerfile
  web.Dockerfile
  mobile.Dockerfile
  client-test.Dockerfile
apps/
  api/                Go + Postgres
  web/                Next.js
  mobile/             Expo
  desktop/            Electron opcional (fuera de Docker)
packages/
  api-client/         sync offline compartido
```

## Offline

Cada cliente guarda `{ name, updatedAt, updatedBy, version, dirty }` en storage local.

1. Leer / editar pega primero en local.
2. Si hay red y lo local es más nuevo → `PUT /api/hello`.
3. Si Postgres es más nuevo, gana la DB.
4. Al volver online se vuelve a reconciliar.

## API

```http
GET /api/hello
PUT /api/hello
Content-Type: application/json

{ "name": "Ada", "updatedAt": "2026-09-05T20:00:00Z", "updatedBy": "device-id" }
```

`200` si aceptó el write, `409` si el row de Postgres era más nuevo.
