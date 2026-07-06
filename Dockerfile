FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/ai-gateway ./cmd/server

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /out/ai-gateway /app/ai-gateway
COPY api/openapi.yaml /app/api/openapi.yaml
EXPOSE 8080
ENTRYPOINT ["/app/ai-gateway"]
