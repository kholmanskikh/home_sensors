#include <iostream>
#include <zmq.h>
#include "Publisher.hpp"
#include "RadioMessage.hpp"

Publisher::Publisher(const std::string& endpoint) {
    endpoint_ = endpoint;

    ctx_ = zmq_ctx_new();
    if (ctx_ == NULL)
        throw std::runtime_error("Unable to create a ZMQ context");

    sock_ = zmq_socket(ctx_, ZMQ_PUB);
    if (sock_ == NULL)
        throw std::runtime_error("Unable to create a ZMQ socket");

    if (zmq_bind(sock_, endpoint_.c_str()))
        throw std::runtime_error("Unable to bind the ZMQ socket");
}

Publisher::~Publisher() {
    if ((sock_ != NULL) && (zmq_close(sock_)))
        std::cerr << "Unable to close the ZMQ socket" << std::endl;

    if ((ctx_ != NULL) && (zmq_ctx_term(ctx_)))
        std::cerr << "Unable to terminate the ZMQ context" << std::endl;
}

bool Publisher::publishMessage(const struct RadioMessage& msg) {
    std::string s = msg.toJsonString();

    return (zmq_send(sock_, s.c_str(), s.size(), 0) != -1);
}

std::string Publisher::endpoint() const {
    return endpoint_;
}
