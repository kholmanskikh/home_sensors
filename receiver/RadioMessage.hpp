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

    unsigned long long timestamp = 0;
    unsigned int device_id = 0;
    Type type;
    double value = 0.0;
    std::string error;

    std::string toJsonString() const;
};

std::string RadioMessageTypeToString(const RadioMessage::Type t);

std::ostream& operator<<(std::ostream& os, const struct RadioMessage& msg);

void to_json(json& j, const struct RadioMessage& msg);
