#pragma once
#include <string>
#include <nlohmann/json_fwd.hpp>

using nlohmann::json;

struct RadioMessage {
    enum class Type {
        Temperature,
        Humidity,
        Error,
    };
    Type type;

    int64_t timestamp = 0;
    uint8_t device_id = 0;
    double value = 0.0;
    std::string error;

    std::string toJsonString() const;
};

std::string RadioMessageTypeToString(const RadioMessage::Type t);

std::ostream& operator<<(std::ostream& os, const struct RadioMessage& msg);

void to_json(json& j, const struct RadioMessage& msg);
