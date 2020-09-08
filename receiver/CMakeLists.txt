cmake_minimum_required(VERSION 3.10)

project(RadioReceiver LANGUAGES CXX)

set(JSON_BuildTests OFF CACHE INTERNAL "")
set(JSON_Install OFF)
set(JSON_MultipleHeaders ON)
add_subdirectory(extern/json)

add_executable(radio_receiver radio_receiver.cpp
                            RadioMessage.cpp
                            Receiver.cpp
                            Publisher.cpp
                            )

target_include_directories(radio_receiver PUBLIC extern/cxxopts/include)

target_link_libraries(radio_receiver PRIVATE zmq pthread rf24)
target_link_libraries(radio_receiver PRIVATE nlohmann_json::nlohmann_json)

target_compile_features(radio_receiver PUBLIC cxx_std_17)
set_target_properties(radio_receiver PROPERTIES CXX_EXTENSIONS OFF)

