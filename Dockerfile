FROM golang:buster

COPY . /src

RUN cd /src && go build -o /src/bin/proxy

EXPOSE 8080

ENTRYPOINT /src/bin/proxy
