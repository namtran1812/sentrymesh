FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /out/sentrymesh \
    ./cmd/sentrymesh

FROM alpine:3.22

RUN addgroup -S sentrymesh && \
    adduser -S sentrymesh -G sentrymesh

WORKDIR /app

COPY --from=builder /out/sentrymesh /usr/local/bin/sentrymesh

RUN mkdir -p /data && \
    chown -R sentrymesh:sentrymesh /data

USER sentrymesh

EXPOSE 8080

ENV SENTRYMESH_AUDIT_DB=/data/sentrymesh-audit.db
ENV SENTRYMESH_APPROVAL_DB=/data/sentrymesh-approvals.db
ENV SENTRYMESH_AUTH_DB=/data/sentrymesh-auth.db
ENV SENTRYMESH_ABUSE_DB=/data/sentrymesh-abuse.db
ENV SENTRYMESH_ROOT=/app

CMD ["sentrymesh"]
