FROM node:22-bookworm-slim
RUN corepack enable
WORKDIR /app

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY apps/web/package.json apps/web/
COPY apps/api/package.json apps/api/
COPY apps/mobile/package.json apps/mobile/
COPY apps/desktop/package.json apps/desktop/
COPY packages/api-client packages/api-client

RUN pnpm install --frozen-lockfile --filter @hola/api-client
CMD ["pnpm", "--filter", "@hola/api-client", "test"]
