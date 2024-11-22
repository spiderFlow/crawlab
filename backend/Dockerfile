FROM golang:1.22.9 AS build

WORKDIR /go/src/app

COPY . .

ENV GO111MODULE on

RUN go install -v ./...

FROM alpine:3.14

# copy files
COPY --from=build /go/bin/crawlab /go/bin/crawlab
