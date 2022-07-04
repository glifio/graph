# syntax=docker/dockerfile:1
FROM golang:1.18-bullseye AS build_base
# FROM golang:1.17.4-alpine AS build_base

# Set the Current Working Directory inside the container
WORKDIR /build

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

# Unit tests
# RUN CGO_ENABLED=0 go test -v

# Build the Go app
RUN go build -o /build/giraph

# Start fresh from a smaller image
FROM debian:bullseye-slim
# FROM alpine:3.15.0


RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates

RUN update-ca-certificates

# RUN apt-get install -y ca-certificates
#RUN apk add ca-certificates

COPY --from=build_base /build/giraph /app/giraph

# This container exposes port 8080 to the outside world
EXPOSE 9090

# Run the binary program produced by `go install`
CMD ["/app/giraph", "daemon"]
