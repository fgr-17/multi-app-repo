FROM node:22-bookworm-slim
RUN corepack enable
WORKDIR /app

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY apps/web/package.json apps/web/
COPY apps/api/package.json apps/api/
COPY apps/mobile/package.json apps/mobile/
COPY apps/desktop/package.json apps/desktop/
COPY packages/api-client/package.json packages/api-client/

RUN pnpm install --frozen-lockfile --filter @hola/web --filter @hola/api-client

COPY packages/api-client packages/api-client
COPY apps/web apps/web

ENV NEXT_TELEMETRY_DISABLED=1 \
    WATCHPACK_POLLING=true
EXPOSE 3000
CMD ["pnpm", "--filter", "@hola/web", "dev"]
