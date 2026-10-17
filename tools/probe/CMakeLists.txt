cmake_minimum_required(VERSION 3.22)
project(dvg-probe CXX)

# dvg_probe — devilGo-vsx3 parity-fixture dumper.
#
# Expects to be added as a subdirectory of a devilutionX build configured with
# BUILD_TESTING=ON (libdevilutionx_so). See tools/probe/README.md.

add_executable(dvg_probe "${CMAKE_CURRENT_LIST_DIR}/dvg_probe.cpp")
target_link_libraries(dvg_probe PRIVATE libdevilutionx_so)
# Headless fatal-error handling: print to stderr and abort instead of dialog+exit.
target_sources(dvg_probe PRIVATE "${CMAKE_SOURCE_DIR}/test/app_fatal_for_testing.cpp")
target_include_directories(dvg_probe PRIVATE "${CMAKE_SOURCE_DIR}/3rdParty/PicoSHA2")
set_target_properties(dvg_probe PROPERTIES RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}")
