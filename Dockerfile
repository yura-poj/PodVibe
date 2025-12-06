FROM golang:1.23 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir -p /app/storage
RUN CGO_ENABLED=0 GOOS=linux go build -o podvibe ./cmd/api

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/podvibe /app/podvibe
COPY --from=builder /app/storage /app/storage

ENV APP_PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/podvibe"]
