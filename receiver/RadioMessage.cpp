#include <ostream>
#include <string>
#include <nlohmann/json.hpp>
#include "RadioMessage.hpp"

using nlohmann::json;

std::string RadioMessageTypeToString(const RadioMessage::Type t) {
    switch(t) {
        case RadioMessage::Type::Temperature: return "Temperature";
        case RadioMessage::Type::Humidity: return "Humidity";
        case RadioMessage::Type::Error: return "Error";
        default: return "Unknown type";
    }
}

std::string RadioMessage::toJsonString() const {
    return json(*this).dump();
}

std::ostream& operator<<(std::ostream& os, const struct RadioMessage& msg) {
    os << "Timestamp: " << msg.timestamp << ", Device id: " << +msg.device_id << ", ";

    os << RadioMessageTypeToString(msg.type) << ": ";

    if (msg.type == RadioMessage::Type::Error) {
        os << msg.error;
    } else {
        os << msg.value;
    }

    return os;
}

void to_json(json& j, const struct RadioMessage& msg) {
    j = json{
            { "timestamp", msg.timestamp },
            { "device_id", msg.device_id },
            { "type", RadioMessageTypeToString(msg.type) },
            };

    if (msg.type == RadioMessage::Type::Error) {
        j["error"] = msg.error;
    } else {
        j["value"] = msg.value;
    }
}
