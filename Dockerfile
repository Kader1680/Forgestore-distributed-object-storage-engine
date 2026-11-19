FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/coordinator ./cmd/coordinator
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/storage-node ./cmd/storage-node

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /out/coordinator /app/coordinator
COPY --from=builder /out/storage-node /app/storage-node

USER app

CMD ["/app/coordinator"]