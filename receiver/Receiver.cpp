#include <stdexcept>
#include <string>
#include <sstream>
#include <vector>
#include <chrono>
#include "Receiver.hpp"
#include "RadioMessage.hpp"
#include "radio_protocol.h"

Receiver::Receiver(rf24_gpio_pin_t cepin, rf24_gpio_pin_t cspin,
                     const std::vector<std::string>& addr) {
    if ((addr.size() == 0) || (addr.size() > 5))
        throw std::runtime_error("Cannot listen more than 5 addresses");

    addr_ = std::vector<std::string>(addr);
    cepin_ = cepin;
    cspin_ = cspin;
    radio = new RF24(cepin_, cspin_);
}

Receiver::~Receiver() {
    radio->powerDown();
    delete radio;
}

rf24_gpio_pin_t Receiver::cepin() const {
    return cepin_;
}

rf24_gpio_pin_t Receiver::cspin() const {
    return cspin_;
}

std::vector<std::string> Receiver::addr() const {
    return addr_;
}

void Receiver::init() {
    radio->begin();        

    if (!radio->isChipConnected())
        throw std::runtime_error("Radio chip is not responding");

    if (sizeof(buffer) < radio->getPayloadSize())
        throw std::runtime_error("Radio payload size is greater than buffer");

    radio->setPALevel(RF24_PA_LOW);

    for (uint8_t i = 0; i < addr_.size(); i++)
        radio->openReadingPipe(i + 1, (uint8_t *)addr_[i].c_str());

    radio->startListening();
}

bool Receiver::messageAvailable() {
    return radio->available();
}

struct RadioMessage Receiver::receiveMessage() {
    auto now = std::chrono::system_clock::now();
    int64_t timestamp = std::chrono::duration_cast<std::chrono::seconds>(now.time_since_epoch()).count();

    memset(buffer, 0, sizeof(buffer));
    radio->read(buffer, sizeof(buffer));

    struct radio_message* rmsg = (struct radio_message*)buffer;

    struct RadioMessage msg;
    if (rmsg->type == MSG_TEMP) {
        if (rmsg->temp.f >= 100) {
            msg.type = RadioMessage::Type::Error;

            std::stringstream s;
            s << "Temperature value (" << +rmsg->temp.i << ", " <<
                +rmsg->temp.f << ") is invalid";

            msg.error = s.str();
        } else {
            msg.type = RadioMessage::Type::Temperature;
            msg.value = rmsg->temp.i + rmsg->temp.f * 0.01;
        }
    } else if (rmsg->type == MSG_HUMID) {
        if ((!rmsg->humid.f) || (rmsg->humid.f >= 100) || (!rmsg->humid.i) || (rmsg->humid.i >= 100)) {
            msg.type = RadioMessage::Type::Error;

            std::stringstream s;
            s << "Humidity value (" << +rmsg->humid.i << ", " <<
                +rmsg->humid.f << ") is invalid";

            msg.error = s.str();
        } else {
            msg.type = RadioMessage::Type::Humidity;
            msg.value = rmsg->humid.i + rmsg->humid.f * 0.01;
        }
    } else if (rmsg->type == MSG_ERR) {
        msg.type = RadioMessage::Type::Error;

        switch (rmsg->err_code) {
            case ERR_TEMP_FAILURE: msg.error = "Temperature measurement error";
                                   break;
            case ERR_HUMID_FAILURE: msg.error = "Humidity measurement error";
                                    break;
            case ERR_LOWPOWER: msg.error = "Low power";
                               break;
            case ERR_OTHER: msg.error = "Other error";
                            break;
            default: msg.error = "Unknown error";
                     break;
        }
    } else {
        msg.error = "Invalid type in radio_message";
    }

    msg.timestamp = timestamp;
    msg.device_id = rmsg->device_id;

    return msg;
}
