#pragma once
#include <string>
#include "RadioMessage.hpp"

class Publisher {
public:
    Publisher(const std::string& endpoint);
    ~Publisher();

    bool publishMessage(const struct RadioMessage& msg);

    std::string endpoint() const;

private:
    std::string endpoint_;
    void *ctx_ = NULL;
    void *sock_ = NULL;
};
