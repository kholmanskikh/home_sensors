#include <iostream>
#include <thread>
#include <chrono>
#include <vector>
#include <string>
#include <sstream>
#include <cxxopts.hpp>
#include "Receiver.hpp"
#include "Publisher.hpp"
#include "RadioMessage.hpp"

template<typename It>
static std::string join(const std::string& sep, It it1, It it2) {
    std::ostringstream os;

    bool is_first = true;
    while (it1 != it2) {
        if (!is_first)
            os << sep;

        os << *it1;

        is_first = false;

        it1++;
    }

    return os.str();
}

int main(int argc, char** argv) {
    cxxopts::Options options(argv[0],
                            "Receive radio messages and publish them as JSON via ZMQ");
    options.add_options()
        ("l,listen", "Radio listen address",
            cxxopts::value<std::vector<std::string>>()->default_value("0Node"))
        ("e,cepin", "GPIO pin connected to the radio CE pin",
            cxxopts::value<unsigned int>()->default_value("22"))
        ("s,cspin", "SPI CS pin (0, 1, ...)",
            cxxopts::value<unsigned int>()->default_value("0"))
        ("p,publish", "ZMQ Publish interface",
            cxxopts::value<std::string>()->default_value("tcp://127.0.0.1:5555"))
        ("d,debug", "Enable debugging",
            cxxopts::value<bool>()->default_value("false"))
        ("h,help", "Print usage")
        ;

    auto opts = options.parse(argc, argv);

    if (opts.count("help")) {
        std::cout << options.help() << std::endl;
        return 0;
    }

    auto debug = opts["debug"].as<bool>();
    std::cout << "Debugging: " << ((debug) ? "enabled" : "disabled") << std::endl;

    Receiver receiver(opts["cepin"].as<unsigned int>(),
                        opts["cspin"].as<unsigned int>(),
                        opts["listen"].as<std::vector<std::string>>());
    receiver.init();

    Publisher publisher(opts["publish"].as<std::string>());

    std::cout << "Radio chip: Chip Enable pin = " << receiver.cepin() <<
        " , Chip Select SPI pin = " << receiver.cspin() << std::endl;

    auto addr = receiver.addr();
    std::cout << "Receiving radio messages for addresses: " <<
        join(", ", addr.begin(), addr.end()) << std::endl;

    std::cout << "Publishing JSON messages to ZMQ socket: " << 
        publisher.endpoint() << std::endl;

    while (true) {
        if (!receiver.messageAvailable()) {
            std::this_thread::sleep_for(std::chrono::seconds(1));
            continue;
        }

        struct RadioMessage msg = receiver.receiveMessage();

        if (debug)
            std::cout << "Received message: " << msg << std::endl;

        if (msg.type == RadioMessage::Type::Error) {
            std::cerr << "Device with id " << msg.device_id << " reported error: "
                    << msg.error << std::endl;
        }

        if (!publisher.publishMessage(msg))
            std::cerr << "Unable to publish message: " << msg << std::endl;
    }

    return 0;
}
