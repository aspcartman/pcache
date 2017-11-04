FROM alpine

ADD build/pcache .
VOLUME /var/pcache
EXPOSE 8080

ENTRYPOINT ["./pcache"]
CMD []
