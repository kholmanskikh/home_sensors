# Radio receiver

This program listens an nRF24L01+ device for messages of the format
described in `radio_protocol.h`.

Upon receiving a valid message, the program converts it to JSON and publishes
to a ZMQ socket. 

# Dependencies

A C++ compiler (tested only with GCC), `cmake` and the following libs:

* [Optimized High Speed Driver for nRF24L01(+) 2.4GHz Wireless Transceiver][1]
* [ZeroMQ C library][2]
* [cxxopts][3]
* [nlohmann json library][4]

# Build steps

Install [2] in such a way so the compiler/linker will know where
to find them. All other dependencies are in `extern` as git submodules.

Typical build procedure:

```sh
git clone --recurse-submodules --shallow-submodules <repo>
cd <repo>
# Note, the SPIDEV driver will be used. It's set in CMakeLists.txt
# Other drivers/options were not tested
cmake -B build_dir
cmake --build build_dir
# This will install the executable along with the RF24 shared lib
cmake --install build_dir --prefix <your prefix>
```

# How to run

To get a list of supported options:

```sh
LD_LIBRARY_PATH="<your prefix>/lib" "<your prefix>/bin/radio_receiver" --help
```


[1]: https://nrf24.github.io/RF24/
[2]: https://zeromq.org/
[3]: https://github.com/jarro2783/cxxopts
[4]: https://github.com/nlohmann/json
