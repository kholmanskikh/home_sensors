FROM golang:latest AS builder

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y libzmq3-dev
WORKDIR /go/src/zmq_to_web
COPY . .
RUN go get -v
RUN go build -v

FROM debian:stable-slim
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y libzmq5
COPY --from=builder /go/src/zmq_to_web/zmq_to_web /bin/zmq_to_web
RUN groupadd zmq_to_web && \
    useradd -g zmq_to_web -s /usr/sbin/nologin -d /tmp zmq_to_web
USER zmq_to_web 
ENTRYPOINT ["/bin/zmq_to_web"]
