FROM node:22-bookworm-slim
RUN corepack enable
WORKDIR /app

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY apps/web/package.json apps/web/
COPY apps/api/package.json apps/api/
COPY apps/mobile/package.json apps/mobile/
COPY apps/desktop/package.json apps/desktop/
COPY packages/api-client/package.json packages/api-client/

RUN pnpm install --frozen-lockfile --filter @hola/mobile --filter @hola/api-client

COPY packages/api-client packages/api-client
COPY apps/mobile apps/mobile

ENV CI=1 \
    EXPO_NO_TELEMETRY=1 \
    EXPO_DEVTOOLS_LISTEN_ADDRESS=0.0.0.0
EXPOSE 8081 19000 19001 19002
CMD ["pnpm", "--filter", "@hola/mobile", "exec", "expo", "start", "--lan", "--port", "8081"]
