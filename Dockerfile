# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/backend ./cmd/api

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app && apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=build /out/backend /app/backend

USER app
EXPOSE 8080

ENTRYPOINT ["/app/backend"]
