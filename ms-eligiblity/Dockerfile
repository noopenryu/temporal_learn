# Build stage
FROM golang:1.27-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o ms-eligiblity ./cmd/server

# Final runtime image
FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/ms-eligiblity /app/ms-eligiblity

EXPOSE 8081

ENTRYPOINT ["/app/ms-eligiblity"]
