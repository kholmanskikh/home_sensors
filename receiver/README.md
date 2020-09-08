# Radio receiver

This program listens an nRF24L01+ device for messages of the format
described in `radio_protocol.h`.

Upon receiving a valid message, the program converts it to JSON and publishes
to a ZMQ socket. 

# Dependencies

* [Optimized High Speed Driver for nRF24L01(+) 2.4GHz Wireless Transceiver][1]
* [ZeroMQ C library][2]
* [cxxopts][3]
* [nlohmann json library][4]

# Build steps

Install RF24 and ZMQ libraries in such a way so the linker will know
where to find them.  

Place the sources of the cxxopts library to `extern/cxxopts`.

Place the sources of the json library to `extern/json`.

And build with `cmake`:

    mkdir build
    cd build
    cmake ..

The generated binary will be `radio_receiver`.

# How to run

To get a list of supported options:

    ./radio_receiver --help

[1]: http://tmrh20.github.io/RF24/
[2]: https://zeromq.org/
[3]: https://github.com/jarro2783/cxxopts
[4]: https://github.com/nlohmann/json
