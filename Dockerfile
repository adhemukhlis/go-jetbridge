# Build Stage
FROM golang:1.25-alpine AS builder

# Install dependencies
RUN apk add --no-cache git curl bash

# Install Buf
RUN BIN="/usr/local/bin" && \
    VERSION="1.50.0" && \
    curl -sSL "https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m)" -o "${BIN}/buf" && \
    chmod +x "${BIN}/buf"

# Install Jet and Proto Plugins
RUN go install github.com/go-jet/jet/v2@latest && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

ARG PGHOST
ARG PGPORT
ARG PGUSER
ARG PGPASSWORD
ARG PGDATABASE

ENV PGHOST=$PGHOST
ENV PGPORT=$PGPORT
ENV PGUSER=$PGUSER
ENV PGPASSWORD=$PGPASSWORD
ENV PGDATABASE=$PGDATABASE

# 1. Jet Generate
RUN TEMP_GEN_DIR=$(mktemp -d) && \
    jet -source=PostgreSQL -host=${PGHOST} -port=${PGPORT} -user=${PGUSER} -password=${PGPASSWORD} -dbname=${PGDATABASE} -path=$TEMP_GEN_DIR -ignore-tables=_prisma_migrations && \
    rm -rf ./gen/jet && mkdir -p ./gen/jet && \
    mv $TEMP_GEN_DIR/${PGDATABASE}/* ./gen/jet/ && \
    rm -rf $TEMP_GEN_DIR

# 2. Protobuf Generate
RUN buf generate

# 3. Go Vet (Linting)
RUN go vet ./cmd/... ./internal/...

# 4. Go Build (Production Target)
RUN GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o server cmd/server/main.go

# Final Stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/server .

# Expose gRPC port
EXPOSE 50051

CMD ["./server"]