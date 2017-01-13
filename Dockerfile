FROM golang:1.7.4-alpine
MAINTAINER Larry Anderson <larryboymi@hotmail.com>

ARG GO_MAIN
ARG GO_MAIN_EXEC

RUN apk add --no-cache git \
    && go get $GO_MAIN \
    && apk del git

EXPOSE 8080

COPY ./cert.pem /go/bin
COPY ./key.pem /go/bin

WORKDIR /go/bin

CMD ./$GO_MAIN_EXEC
