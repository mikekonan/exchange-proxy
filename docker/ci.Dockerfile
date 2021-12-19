FROM alpine:3.15

RUN adduser -g "proxy" -D -H proxy proxy

RUN apk --no-cache add ca-certificates \
    && rm -rf /var/cache/apk/*

COPY exchange-proxy /bin/proxy

USER proxy

EXPOSE 8080

ENTRYPOINT ["/bin/proxy"]
