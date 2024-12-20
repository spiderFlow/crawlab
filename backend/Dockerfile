FROM golang:1.22.9 AS build

WORKDIR /go/src/app

# Copy all required modules from parent directory
COPY .. .

ENV GO111MODULE on

# Build from the backend directory which contains the main.go
RUN cd backend && go install -v ./...

FROM alpine:3.14

# Copy the built binary
COPY --from=build /go/bin/crawlab /go/bin/crawlab
