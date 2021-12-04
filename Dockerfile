FROM golang:buster as builder

COPY . /src

RUN cd /src && go build -o /src/bin/proxy

FROM gcr.io/distroless/base

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/bin/proxy /bin/proxy

EXPOSE 8080

ENTRYPOINT ["/bin/proxy"]
