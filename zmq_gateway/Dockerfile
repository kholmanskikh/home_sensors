# do not forget to 'go vendor' the source code
FROM golang:latest AS builder

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y libzmq3-dev
WORKDIR /go/src/zmq_gateway
COPY . .

FROM debian:stable-slim
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y libzmq5
COPY --from=builder /go/src/zmq_gateway/zmq_gateway /bin/zmq_gateway
RUN groupadd zmq_gateway && \
    useradd -g zmq_gateway -s /usr/sbin/nologin -d /tmp zmq_gateway
USER zmq_gateway
ENTRYPOINT ["/bin/zmq_gateway"]
