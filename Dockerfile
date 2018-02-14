FROM alpine

RUN apk add --no-cache ca-certificates

ADD build/pcache /pcache
VOLUME /var/pcache
EXPOSE 8080

ENTRYPOINT ["/pcache"]
