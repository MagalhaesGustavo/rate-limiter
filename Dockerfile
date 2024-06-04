# Builder
FROM golang:1.22.2 as build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o rate-limiter ./cmd/server

FROM alpine:latest
EXPOSE 8080
WORKDIR /app
COPY --from=build /app/rate-limiter .
COPY cmd/server/.env .
ENTRYPOINT ["./rate-limiter"]
