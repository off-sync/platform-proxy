FROM alpine:3.2
MAINTAINER Bas Verweij "bas@off-sync.com"

RUN apk update

RUN apk add ca-certificates

WORKDIR /dist

ADD dist /dist/

ENTRYPOINT ["/dist/proxy"]