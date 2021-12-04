FROM golang:buster as builder

COPY . /src

RUN cd /src && go build -ldflags '-linkmode external -extldflags -static -w' -o /src/bin/proxy

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/bin/proxy /bin/proxy

EXPOSE 8080

ENTRYPOINT ["/bin/proxy"]
