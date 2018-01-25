FROM alpine

ADD build/pcache /pcache
VOLUME /var/pcache
EXPOSE 8080

ENTRYPOINT ["/pcache"]
