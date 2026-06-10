#include "classgroup/form.hpp"
#include "hash/sha256.hpp"

#include <doctest/doctest.h>
#include <vector>

TEST_CASE("derives a negative discriminant from a 3 mod 4 prime")
{
    mpz_t p, D, three, t;
    mpz_inits(p, D, three, t, nullptr);
    mpz_set_ui(three, 3);

    hash::sha256_to_prime_3mod4(p, std::vector<uint8_t>{0x01, 0x02, 0x03, 0x04}, 128);
    classgroup::derive_discriminant(D, std::vector<uint8_t>{0x01, 0x02, 0x03, 0x04}, 128);

    mpz_and(t, p, three);
    CHECK(mpz_cmp_ui(t, 3) == 0);
    CHECK(mpz_sizeinbase(p, 2) == 128);

    mpz_neg(p, p);
    CHECK(mpz_cmp(D, p) == 0);

    mpz_clears(p, D, three, t, nullptr);
}
