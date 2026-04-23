# go-jetbridge

A high-performance Go microservice architecture purpose-built for **Server-to-Server (S2S)** communication. `go-jetbridge` delivers a **strict, low-latency gRPC** interface optimized for high-concurrency backend ecosystems, ensuring seamless interoperability with any system utilizing the gRPC protocol. By integrating the Jet SQL builder for type-safe database operations and Prisma for declarative schema management, it provides a robust and scalable foundation tailored for the demanding requirements of modern microservice infrastructures.

## run

```bash
go run cmd/server/main.go
```

## build

```bash
go build -ldflags "-s -w" -o server cmd/server/main.go
```

## generate protobuf

```
buf generate
```

## generate shared protos

```
buf export . --output ./contract
```

> use `contract` in other microservices

## generate jet

```
./jet-generate.sh
```

## migrate database

```
cd db-schema && pnpm prisma:reset
```

## deps tidy

```bash
go mod tidy
```

## linting

```bash
golangci-lint run
```

## formatting

```bash
goimports -w .
```

## dependencies update

```bash
rm go.sum && go get -u ./... && go mod tidy
```

## buf pull

```bash
buf export https://github.com/adhemukhlis/go-jetbridge.git#branch=main,subdir=contract --output ./contract
```

## stack

| name       | version | language |
| :--------- | :------ | :------- |
| Go         | 1.26.2  | go       |
| Jet        | 2.14.1  | go       |
| gRPC       | 1.80.0  | go       |
| PostgreSQL | 17      | sql      |
| Prisma     | 7.7.0   | ts       |
