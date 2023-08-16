FROM golang:1.21-alpine3.18 as builder

RUN apk --no-cache add gcc musl-dev make && go get github.com/mailru/easyjson && go install github.com/mailru/easyjson/...@latest

COPY . /src

ARG VERSION=dev

RUN cd /src && make generate && go build -o /src/bin/proxy -ldflags "-s -w -X main.version=$VERSION"

FROM alpine:3.18

RUN adduser -g "proxy" -D -H proxy proxy

RUN apk --no-cache add ca-certificates \
    && rm -rf /var/cache/apk/*

COPY --from=builder /src/bin/proxy /bin/proxy

USER proxy

EXPOSE 8080

ENTRYPOINT ["/bin/proxy"]
