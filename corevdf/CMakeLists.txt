cmake_minimum_required(VERSION 3.20)
project(corevdf LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
if(NOT CMAKE_BUILD_TYPE AND NOT CMAKE_CONFIGURATION_TYPES)
  set(CMAKE_BUILD_TYPE Release CACHE STRING "Build type" FORCE)
endif()

find_package(PkgConfig REQUIRED)
pkg_check_modules(GMP REQUIRED IMPORTED_TARGET gmp gmpxx)
find_package(OpenSSL REQUIRED)
find_package(cxxopts CONFIG REQUIRED)
include(FetchContent)

FetchContent_Declare(
  doctest
  GIT_REPOSITORY https://github.com/doctest/doctest.git
  GIT_TAG v2.4.12
)
FetchContent_MakeAvailable(doctest)

function(add_vdf_app app_name app_source)
  add_executable(${app_name} ${app_source})
  target_link_libraries(${app_name} PRIVATE corevdf_lib cxxopts::cxxopts)
  target_include_directories(${app_name} PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
endfunction()

add_library(corevdf_lib
  src/common.cpp
  src/classgroup/form.cpp
  src/classgroup/compose.cpp
  src/hash/poseidon2.cpp
  src/hash/sha256.cpp
  src/vdf/system.cpp
  src/vdf/prover.cpp
  src/vdf/verifier.cpp
)

target_include_directories(corevdf_lib PUBLIC include)
target_link_libraries(corevdf_lib PUBLIC PkgConfig::GMP OpenSSL::Crypto)

add_executable(vdf_tests
  tests/test.cpp
  tests/test_hash_sha256.cpp
  tests/test_discriminant.cpp
  tests/test_form_reduction_nucomp.cpp
  tests/test_hash_to_form.cpp
  tests/test_vdf.cpp
)
target_link_libraries(vdf_tests PRIVATE corevdf_lib doctest::doctest)

add_vdf_app(corevdf apps/main.cpp)

enable_testing()
add_test(NAME vdf_tests COMMAND vdf_tests)
