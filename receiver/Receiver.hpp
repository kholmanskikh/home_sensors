#pragma once
#include <string>
#include <vector>
#include <RF24/RF24.h>
#include "RadioMessage.hpp"

class Receiver {
public:
    Receiver(unsigned int cepin, unsigned int cspin,
                const std::vector<std::string>& addr);
    ~Receiver();

    unsigned int cepin() const;
    unsigned int cspin() const;
    std::vector<std::string> addr() const;

    void init();
    bool messageAvailable();
    struct RadioMessage receiveMessage();

private:
    RF24 *radio;
    unsigned int cepin_;
    unsigned int cspin_;
    std::vector<std::string> addr_;
    const int buf_len = 64;
};
