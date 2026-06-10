#pragma once

#include <chrono>
#include <gmpxx.h>
#include <vector>
#include <tuple>

namespace common
{
    using Clock = std::chrono::steady_clock;
    std::string elapsed_ms(Clock::time_point start);

    std::vector<mpz_class> mpz_split_be(mpz_srcptr a, uint16_t bit_per_num, uint16_t target_out);
    void append_u16_be(std::vector<uint8_t> &out, uint16_t v);
    void append_u32_be(std::vector<uint8_t> &out, uint32_t v);
    void append_u64_be(std::vector<uint8_t> &out, uint64_t v);

    std::vector<uint8_t> hex_to_bytes(const std::string &in);
    std::string bytes_to_hex(const std::vector<uint8_t> &bytes);

    void mpz_set_hex(mpz_ptr out, const std::string hex);
    std::string mpz_get_hex(mpz_srcptr x);
    void mpz_set_from_be(mpz_ptr out, const std::vector<uint8_t> be);
    void mpz_mod_pos(mpz_ptr out, mpz_srcptr a, mpz_srcptr m);
    size_t mpz_bit_length(mpz_srcptr a);
    std::tuple<int64_t, int64_t> mpz_get_si_2exp(mpz_srcptr a);
    std::vector<uint8_t> mpz_to_big_endian_with_sign(mpz_srcptr a, size_t width);
    std::vector<uint8_t> mpz_to_big_endian(mpz_srcptr a, size_t width);
    std::vector<uint8_t> mpz_to_big_endian_minimal(mpz_srcptr a);
}
