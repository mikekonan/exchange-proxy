FROM golang:1.17.4-alpine3.15 as builder

RUN apk --no-cache add gcc musl-dev

COPY . /src

RUN cd /src && go build -o /src/bin/proxy

FROM alpine:3.15

RUN adduser -g "proxy" -D -H proxy proxy

RUN apk --no-cache add ca-certificates \
    && rm -rf /var/cache/apk/*

COPY --from=builder /src/bin/proxy /bin/proxy

USER proxy

EXPOSE 8080

ENTRYPOINT ["/bin/proxy"]
