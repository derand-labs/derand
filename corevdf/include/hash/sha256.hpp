#pragma once

#include <vector>
#include <gmpxx.h>

namespace hash
{
    std::vector<uint8_t> sha256(const std::vector<uint8_t> &data);

    std::vector<mpz_class> sha256_to_prime(mpz_ptr out, const std::vector<uint8_t> &seed, const size_t nbits, const bool ensurebits, const bool export_transcript);
    void sha256_to_prime_3mod4(mpz_ptr out, const std::vector<uint8_t> &seed, const size_t nbits);
}
