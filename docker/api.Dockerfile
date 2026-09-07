FROM golang:1.22-alpine AS build
WORKDIR /src
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ .
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server \
 && CGO_ENABLED=0 go build -o /out/relay ./cmd/relay \
 && CGO_ENABLED=0 go build -o /out/projector ./cmd/projector

FROM alpine:3.20
COPY --from=build /out/server /out/relay /out/projector /usr/local/bin/
ENV PORT=8080
EXPOSE 8080
HEALTHCHECK --interval=2s --timeout=3s --retries=20 \
  CMD wget -qO- http://127.0.0.1:8080/health >/dev/null
CMD ["server"]
