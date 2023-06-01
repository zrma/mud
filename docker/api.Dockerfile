# Go binaries are standalone, so use a multi-stage build to produce smaller images.

# Use base golang image from Docker Hub
FROM golang:1.19-alpine as build

WORKDIR /app

## module download
COPY go.mod .
COPY go.sum .
RUN go mod download

## copy source
COPY ./pkg ./pkg
COPY ./pb ./pb
COPY ./cmd ./cmd

RUN go build -o /app ./cmd/api

# Now create separate deployment image
FROM gcr.io/distroless/base

# Definition of this variable is used by 'skaffold debug' to identify a golang binary.
# Default behavior - a failure prints a stack trace for the current goroutine.
# See https://golang.org/pkg/runtime/
ENV GOTRACEBACK=single

WORKDIR /app
COPY --from=build /app .
ENTRYPOINT ["./app"]
