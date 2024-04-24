# Build the application from source
FROM golang:1.22.0-alpine AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./transaction/main.go .
COPY ./telemetry /app/telemetry

RUN CGO_ENABLED=0 GOOS=linux go build -o /apiserver

FROM alpine:3.19

COPY --from=build-stage /apiserver /apiserver

RUN chmod +x /apiserver

EXPOSE 3000

ENTRYPOINT ["/apiserver"]