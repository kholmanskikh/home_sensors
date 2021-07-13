# Do not forget to 'go vendor' the source code.
# If you need to install additional CA SSL certificates,
# create the 'certs' dir and put corresponding *.crt files there.

FROM golang:latest AS builder

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y libzmq3-dev
WORKDIR /go/src/zmq_gateway
COPY . .
RUN test -d certs || mkdir certs
RUN go build -v -mod vendor

FROM debian:stable-slim

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y libzmq5 ca-certificates

COPY --from=builder /go/src/zmq_gateway/certs/ /usr/local/share/ca-certificates
RUN update-ca-certificates

COPY --from=builder /go/src/zmq_gateway/zmq_gateway /bin/zmq_gateway
RUN groupadd zmq_gateway && \
    useradd -g zmq_gateway -s /usr/sbin/nologin -d /tmp zmq_gateway
USER zmq_gateway
ENTRYPOINT ["/bin/zmq_gateway"]
