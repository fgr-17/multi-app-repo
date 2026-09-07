# Hola ecosystem

Monorepo de un producto que corre en **web**, **iOS/Android** y **desktop**. El nombre se modela con **event sourcing + CQRS**. Las apps siguen editando **offline** (last-write-wins).

En el host solo hace falta **Docker** y **Make**.

```
make up
# http://localhost:3000              → Hola {nombre}
# http://localhost:8080/api/hello    → read model (Mongo) o replay
# http://localhost:8080/api/events   → log de hechos (Postgres)
```

## Cómo está partido

```
  usuarios  ── CRUD ──►  Postgres.users          (relacional)
  PUT /api/hello  →  comando
                      │
                      ▼
                 Postgres.events     Kafka         Mongo
                 + outbox         (log / bus)   greeting actual
                      └── relay ─────┘── projector ──►
```

| Pieza | Rol |
| --- | --- |
| **Postgres `users`** | Identidad relacional (id, email, nombre, ciudad). CRUD, no event sourcing. |
| **Postgres `events`** | Event store del saludo. Append-only, PK `(stream_id, version)`. |
| **Postgres `outbox`** | Publicación atómica con el evento. El relay lo manda a Kafka. |
| **Kafka `greeting.events`** | Log distribuido. Fan-out a N consumidores. |
| **Mongo `greeting`** | Read model CQRS: un documento con el nombre actual. |
| **Clientes** | Cache local + LWW. No hablan Kafka. |

`GET /api/hello` lee Mongo. Si la proyección está vacía o atrás, el API **replayea** el stream de Postgres.

Si venías de la versión que guardaba una fila `greeting`, corré `make clean` una vez para recrear volúmenes.

## Mongo: ¿sirve para event sourcing?

**No como event store.** Un event store quiere: append ordenado, unicidad `(stream, version)`, replay barato, retención larga. Postgres (tabla `events`) hace eso con una transacción. Mongo puede guardar documentos de eventos, pero no es un log: las queries son por documento, el orden global es más débil, y no reemplaza a Kafka como bus.

**Sí como modelo de lectura.** Acá Mongo guarda `{ _id: "greeting", name, updatedAt, updatedBy, version }`. Eso escala lecturas, admite documentos más ricos después (perfil, preferencias, estado por device) y se reconstruye desde Kafka si lo borramos.

También serviría para: sesiones, inbox de sync por dispositivo, o un catálogo si el saludo deja de ser un singleton. No para el historial de hechos.

## Comandos

```bash
make up          # postgres + mongo + kafka + api + relay + projector + web
make logs
make test
make psql        # event store
make mongosh     # read model
make down
```

## Offline

Las apps no cambian: editan local y reconcilian con `PUT`. El API decide LWW contra el estado foldeado del stream y, si gana, **append** un `GreetingRenamed` (no un `UPDATE` de fila).

## API

```http
GET  /api/hello
PUT  /api/hello   { "name", "updatedAt", "updatedBy" }
GET  /api/events
GET  /api/users
GET  /api/users/{id}
POST /api/users
GET  /health
```

Los usuarios **no** van a Kafka: son un agregado relacional (joins, unique email, datos personales). El saludo sí, porque es un hecho que otros sistemas pueden proyectar.

Identidad vive en Postgres porque querés constraints (`email` único), consultas por ciudad/país y updates in-place. El saludo vive en el event store porque el hecho “se llamó X” se proyecta a Mongo y a lo que venga.
