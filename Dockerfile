FROM golangci/golangci-lint:latest as linter

WORKDIR /app
COPY . .

RUN golangci-lint run

FROM golang:1.13 as builder

LABEL maintainer="Pete Birley <pete@port.direct>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

RUN go version

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

RUN go test -v ./...

# Build the Go app. The netgo tag ensures we build a static binary.
RUN go build -tags netgo -o jarvis-connector ./cmd/connector


FROM alpine:latest as connector

# Add opencontainer labels to publishing image
ARG REVISION
ARG SHORT_SHA
ARG CREATED
ARG VERSION

LABEL org.opencontainers.image.created=$CREATED \
      org.opencontainers.image.authors="AT&T Services INC" \
      org.opencontainers.image.url="https://github.com/att-comdev/jarvis-connector" \
      org.opencontainers.image.documentation="https://github.com/att-comdev/jarvis-connector" \
      org.opencontainers.image.source="https://github.com/att-comdev/jarvis-connector" \
      org.opencontainers.image.version=$VERSION \
      org.opencontainers.image.revision=$REVISION \
      org.opencontainers.image.vendor="AT&T Services INC" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.ref.name="quay.io/attcomdev/"$SHORT_SHA \
      org.opencontainers.image.title="jarvis-connector" \
      org.opencontainers.image.description="Tekton/Gerrit Connection Microservice"

# Add required packages for jarvis-project useage
RUN apk --no-cache add ca-certificates curl openssh-client git

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/jarvis-connector /usr/bin/jarvis-connector
ENTRYPOINT [ "/usr/bin/jarvis-connector" ]
CMD []
