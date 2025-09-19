FROM alpine:latest

VOLUME ["/config"]
WORKDIR /config

RUN apk update && \
    apk add --no-cache tzdata ca-certificates && \
    rm -rf /var/cache/apk/*

COPY ./ovhcloud /usr/local/bin/ovhcloud

ENTRYPOINT ["ovhcloud"]
CMD ["--help"]