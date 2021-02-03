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


# Build the Go app. The netgo tag ensures we build a static binary.
RUN go build -tags netgo -o jarvis-connector ./cmd/connector


FROM alpine:latest as connector

# Add required packages for jarvis-project useage
RUN apk --no-cache add ca-certificates curl openssh-client

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/jarvis-connector /usr/bin/jarvis-connector
ENTRYPOINT [ "/usr/bin/jarvis-connector" ]
CMD []
