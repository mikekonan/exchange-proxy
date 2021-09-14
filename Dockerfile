FROM golang:buster as builder

COPY . /src

RUN cd /src && go build -o proxy

FROM debian:buster

RUN mkdir /app
COPY --from=builder /src/proxy /app/app
RUN chmod +x /app/app

WORKDIR /app
ENTRYPOINT ./app
