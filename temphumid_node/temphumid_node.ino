#include <avr/wdt.h>
#include <avr/power.h>
#include <avr/sleep.h>
#include <util/delay.h>
#include <Arduino.h>
#include <RF24.h>
#include <TroykaMeteoSensor.h>
#include <math.h>
#include "radio_protocol.h"


// The sensor works only down to 2.15V, so we disable BOD (and the boot
// loader)

// device id of the node
#define DEVICE_ID 2

// sleep interval between taking measurements, in seconds
#define WORK_INTERVAL 300

// address of the master/receiver node
const uint8_t remote_address[6] = "0Node";

// set this to enable writing log messages to the serial port
#define LOGGING

#define RADIO_CHIP_ENABLE 10
#define RADIO_CHIP_SELECT 9
#define RADIO_RETRY_COUNT 5

RF24 radio(RADIO_CHIP_ENABLE, RADIO_CHIP_SELECT);
TroykaMeteoSensor sensor;

#ifdef LOGGING
# define logmsg(...) (Serial.print(__VA_ARGS__))
# define logmsgn(...) (Serial.println(__VA_ARGS__))
#else
# define logmsg(...) {}
# define logmsgn(...) {}
#endif /* LOGGING */

void die() {
    logmsgn("Died.");

    while (true) {
        digitalWrite(LED_BUILTIN, HIGH);
        delay(5000);
        digitalWrite(LED_BUILTIN, LOW);
        delay(5000);
    }
}

void init_msg(struct radio_message *msg, uint8_t type) {
    msg->device_id = DEVICE_ID;
    msg->type = type;
}

bool send_msg(const struct radio_message *msg, int retry_count) {
    radio.stopListening();
    radio.openWritingPipe(remote_address);

    bool ret = false;

    for (int i = 0; i < retry_count; i++) {
        if (radio.write(msg, sizeof(*msg))) {
            ret = true;
            break;
        }
        delay(100);
    }

    if (ret) {
        logmsgn("Message sent successfully.");
    } else {
        logmsgn("Unable to send the message.");
    }

    return ret;
}

bool send_error(uint8_t error_code) {
    struct radio_message msg;
    init_msg(&msg, MSG_ERR);
    msg.err_code = error_code;

    return send_msg(&msg, RADIO_RETRY_COUNT);
}

bool send_temperature(float value) {
    struct radio_message msg;
    init_msg(&msg, MSG_TEMP);

    float i;
    float f;
    f = modff(value, &i);

    msg.temp.i = (int8_t)i;
    msg.temp.f = (uint8_t)(fabs(f) * 100.0);

    return send_msg(&msg, RADIO_RETRY_COUNT);
}

bool send_humidity(float value) {
    struct radio_message msg;
    init_msg(&msg, MSG_HUMID);

    float i;
    float f;
    f = modff(value, &i);

    msg.humid.i = (uint8_t)i;
    msg.humid.f = (uint8_t)(f * 100.0);

    return send_msg(&msg, RADIO_RETRY_COUNT);
}

void do_work() {
    bool temp_measured = false;
    bool humid_measured = false;
    float temp;
    float humid;

    radio.begin();
    if (!radio.isChipConnected()) {
        logmsgn("Radio is not connected.");
        die();
    }

    radio.setPALevel(RF24_PA_LOW);

    sensor.begin();

    int sensor_state = sensor.read();
    if (sensor_state == SHT_OK) {
        temp_measured = true;
        humid_measured = true;
        temp = sensor.getTemperatureC();
        humid = sensor.getHumidity();
    }

    if (temp_measured) {
        logmsg("Measured temperature: ");
        logmsgn(temp);
        send_temperature(temp);
    } else {
        logmsg("Unable to measure temperature, error: ");
        logmsgn(sensor_state);
        send_error(ERR_TEMP_FAILURE);
    }
    if (humid_measured) {
        logmsg("Measured humidity: ");
        logmsgn(humid);
        send_humidity(humid);
    } else {
        logmsg("Unable to measure humidity, error: ");
        logmsgn(sensor_state);
        send_error(ERR_HUMID_FAILURE);
    }

    // found empirically, that it saves me ~ 0.08mA
    // so it seems that controlling the radio power using
    // a separate mcu pin may not make much sense in my case
    radio.powerDown();
}

void wake() {
    wdt_disable();
}

ISR (WDT_vect) {
    wake();
}

void sleep_until_wdt(const uint8_t interval) {
    noInterrupts();

    // reactivate the watchdog
    MCUSR = 0;
    WDTCSR |= 0b00011000; // set WDCE, WDE
    WDTCSR = 0b01000000 | interval; // set WDIE, and appropriate delay 
    wdt_reset(); 

    set_sleep_mode(SLEEP_MODE_PWR_DOWN);
    sleep_enable();

    interrupts();

    sleep_cpu();
}

void setup() {
    pinMode(LED_BUILTIN, OUTPUT);
    digitalWrite(LED_BUILTIN, LOW);

    Serial.begin(9600);
    while (!Serial) {
    }

    logmsgn("Started");
    delay(2000);

    power_all_disable();
}

void loop() {
    static uint32_t sleep_time = WORK_INTERVAL;
    static const uint8_t wdt_interval = 0b100001;
    static const byte wdt_interval_len = 8; // in seconds

    sleep_time += wdt_interval_len;
    if (sleep_time >= WORK_INTERVAL) {
        sleep_time = 0;

        power_all_enable();

        // let the stdio stabilize
        delay(100);
        do_work();
        delay(100);

        power_all_disable();
    }

    sleep_until_wdt(wdt_interval);
}
