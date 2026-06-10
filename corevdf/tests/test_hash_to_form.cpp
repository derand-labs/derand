#include "common.hpp"
#include "vdf/system.hpp"

#include <doctest/doctest.h>

TEST_CASE("hash a public seed to the expected class-group form")
{
    vdf::System system(std::vector<uint8_t>{0x01, 0x02, 0x03, 0x04}, 1024, 4, 64, 1, 8, 12);

    vdf::PublicStatement stmt{
        std::vector<uint8_t>{0x12, 0x34, 0x56, 0x78,
                             0x90, 0xab, 0xcd, 0xef,
                             0x11, 0x22, 0x33, 0x44,
                             0x55, 0x66, 0x77, 0x88,
                             0xde, 0xad, 0xbe, 0xef,
                             0xca, 0xfe, 0xba, 0xbe,
                             0xfe, 0xed, 0xfa, 0xce,
                             0x01, 0x23, 0x45, 0x67},
    };

    classgroup::Form x = vdf::hash_to_form(system, stmt.x_seed);

    mpz_class expected_a("6ee48d13ab8a4124947213c6227147338f2320aa0435de71bb317f0e07c484939770cb3cf180843046b998cde1a232cc95b524937928d528dbdedda148e4d7f", 16);
    mpz_class expected_b("-665a36089a280facf6d036c5b4bb6d9c04ee87bd6d02de06e652d6699222dee6765a3bd963a0da014a4a7faec749a4cdd0c2fbe527a98463a800ea8ed35dbff", 16);
    mpz_class expected_c("8700050617ffc4a7479e9be86e77ce6442f8f6d4588e1a26c655b32e18b4d0136a0511e15f4f7a39bac23f77b744bc70b7ce97f7eda1ace5b11faf203ebcafd62", 16);

    classgroup::Form expected(expected_a.get_mpz_t(), expected_b.get_mpz_t(), expected_c.get_mpz_t());

    CHECK(x == expected);
}
