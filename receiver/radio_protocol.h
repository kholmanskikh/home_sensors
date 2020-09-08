#pragma once
#include <stdint.h>

#define MSG_TEMP 0x0
#define MSG_HUMID 0x1
#define MSG_ERR 0xff

#define ERR_TEMP_FAILURE 0xfc
#define ERR_HUMID_FAILURE 0xfd
#define ERR_LOWPOWER 0xfe
#define ERR_OTHER 0xff

struct radio_message {
    uint8_t device_id;
    uint8_t type;
    union {
        uint8_t err_code;
        struct {
            int8_t i; // integer part
            uint8_t f; // fractional part
        } temp;
        struct {
            uint8_t i;
            uint8_t f;
        } humid;
        uint8_t raw_data[8];
    };
};
