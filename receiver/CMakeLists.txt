cmake_minimum_required(VERSION 3.10)

project(RadioReceiver LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

set(JSON_BuildTests OFF CACHE INTERNAL "")
set(JSON_Install OFF CACHE INTERNAL "")
add_subdirectory(extern/json)

# Other drivers are not tested
set(RF24_DRIVER SPIDEV CACHE INTERNAL "")
# RF24/CMakeLists.txt incorrectly passes link info with pigpio
set(RF24_NO_INTERRUPT ON BOOL INTERNAL "")
add_subdirectory(extern/RF24)

add_executable(radio_receiver radio_receiver.cpp
                            RadioMessage.cpp
                            Receiver.cpp
                            Publisher.cpp
                            )

target_include_directories(radio_receiver PUBLIC extern/cxxopts/include)
target_include_directories(radio_receiver PRIVATE extern/RF24)

target_link_libraries(radio_receiver PRIVATE zmq pthread rf24)
target_link_libraries(radio_receiver PRIVATE nlohmann_json::nlohmann_json)

install(TARGETS radio_receiver)
