#pragma once
#include <string>
#include <vector>
#include "RF24.h"
#include "RadioMessage.hpp"

class Receiver {
public:
    Receiver(rf24_gpio_pin_t cepin, rf24_gpio_pin_t cspin,
                const std::vector<std::string>& addr);
    ~Receiver();

    rf24_gpio_pin_t cepin() const;
    rf24_gpio_pin_t cspin() const;
    std::vector<std::string> addr() const;

    void init();
    bool messageAvailable();
    struct RadioMessage receiveMessage();

private:
    RF24 *radio;
    rf24_gpio_pin_t cepin_;
    rf24_gpio_pin_t cspin_;
    std::vector<std::string> addr_;

    uint8_t buffer[64];
};
