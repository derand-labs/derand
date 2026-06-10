#include "common.hpp"

#include <algorithm>
#include <iostream>

namespace common
{
    std::string elapsed_ms(Clock::time_point start)
    {
        auto elapsed = Clock::now() - start;
        auto ms = std::chrono::duration_cast<std::chrono::milliseconds>(elapsed).count();
        return std::to_string(ms) + "ms";
    }

    void append_u16_be(std::vector<uint8_t> &out, uint16_t v)
    {
        out.push_back(static_cast<uint8_t>((v >> 8) & 0xff));
        out.push_back(static_cast<uint8_t>(v & 0xff));
    }

    void append_u32_be(std::vector<uint8_t> &out, uint32_t v)
    {
        out.push_back(static_cast<uint8_t>((v >> 24) & 0xff));
        out.push_back(static_cast<uint8_t>((v >> 16) & 0xff));
        out.push_back(static_cast<uint8_t>((v >> 8) & 0xff));
        out.push_back(static_cast<uint8_t>(v & 0xff));
    }

    void append_u64_be(std::vector<uint8_t> &out, uint64_t v)
    {
        out.push_back(static_cast<uint8_t>((v >> 56) & 0xff));
        out.push_back(static_cast<uint8_t>((v >> 48) & 0xff));
        out.push_back(static_cast<uint8_t>((v >> 40) & 0xff));
        out.push_back(static_cast<uint8_t>((v >> 32) & 0xff));
        out.push_back(static_cast<uint8_t>((v >> 24) & 0xff));
        out.push_back(static_cast<uint8_t>((v >> 16) & 0xff));
        out.push_back(static_cast<uint8_t>((v >> 8) & 0xff));
        out.push_back(static_cast<uint8_t>(v & 0xff));
    }

    std::vector<uint8_t> hex_to_bytes(const std::string &in)
    {
        if (in.size() % 2 != 0)
            throw std::invalid_argument("invalid hex size");

        if (in.empty())
            return {};

        auto hex_val = [](char c) -> int
        {
            if (c >= '0' && c <= '9')
                return c - '0';
            if (c >= 'a' && c <= 'f')
                return c - 'a' + 10;
            if (c >= 'A' && c <= 'F')
                return c - 'A' + 10;
            return -1;
        };

        std::vector<uint8_t> out;
        out.reserve(in.size() / 2);
        for (size_t i = 0; i < in.size(); i += 2)
        {
            int hi = hex_val(in[i]);
            int lo = hex_val(in[i + 1]);
            if (hi < 0 || lo < 0)
            {
                throw std::invalid_argument("invalid hex string");
            }
            out.push_back(static_cast<uint8_t>((hi << 4) | lo));
        }
        return out;
    }

    std::string bytes_to_hex(const std::vector<uint8_t> &bytes)
    {
        static const char *kHex = "0123456789abcdef";
        std::string out;
        out.reserve(bytes.size() * 2);
        for (uint8_t b : bytes)
        {
            out.push_back(kHex[(b >> 4) & 0x0f]);
            out.push_back(kHex[b & 0x0f]);
        }
        return out;
    }

    void mpz_set_hex(mpz_ptr out, const std::string hex)
    {
        int start = 0;
        if (hex.starts_with("0x"))
        {
            start = 2;
        }
        if (hex.starts_with("-0x"))
        {
            start = 3;
        }
        if (start == 0)
        {
            throw std::invalid_argument("hex must be started by 0x");
        }

        mpz_set_str(out, hex.substr(start).c_str(), 16);
        if (start == 3)
        {
            mpz_neg(out, out);
        }
    }

    std::string mpz_get_hex(mpz_srcptr x)
    {
        mpz_class tmp(x);

        std::string prefix = "0x";
        if (mpz_sgn(x) < 0)
        {
            prefix = "-0x";
            mpz_neg(tmp.get_mpz_t(), tmp.get_mpz_t());
        }

        return prefix + tmp.get_str(16);
    }

    void mpz_set_from_be(mpz_ptr out, const std::vector<uint8_t> bytes)
    {
        mpz_set_ui(out, 0);

        for (uint8_t b : bytes)
        {
            mpz_mul_2exp(out, out, 8);
            mpz_add_ui(out, out, b);
        }
    }

    void mpz_mod_pos(mpz_ptr out, mpz_srcptr a, mpz_srcptr m)
    {
        mpz_mod(out, a, m);
        if (mpz_sgn(out) < 0)
            mpz_add(out, out, m);
    }

    size_t mpz_bit_length(mpz_srcptr a)
    {
        if (mpz_sgn(a) == 0)
            return 0;

        return mpz_sizeinbase(a, 2);
    }

    static inline unsigned int lzcnt64_soft(unsigned long long x)
    {
        if (x == 0ULL)
            return 64;
        return static_cast<unsigned int>(__builtin_clzll(x));
    }

    static inline void signed_shift(uint64_t op, int64_t shift, int64_t &r)
    {
        if (shift > 0)
            r = static_cast<int64_t>(op << shift);
        else if (shift <= -64)
            r = 0;
        else
            r = static_cast<int64_t>(op >> (-shift));
    }

    std::tuple<int64_t, int64_t> mpz_get_si_2exp(mpz_srcptr a)
    {
        int64_t size(static_cast<long>(mpz_size(a)));
        uint64_t last(mpz_getlimbn(a, (size - 1)));

        int64_t r, exp;

        int64_t lg2 = exp = ((63 - lzcnt64_soft(last)) + 1);
        signed_shift(last, (63 - exp), r);
        if (size > 1)
        {
            exp += (size - 1) * 64;
            uint64_t prev(mpz_getlimbn(a, (size - 2)));
            int64_t t;
            signed_shift(prev, -1 - lg2, t);
            r += t;
        }
        if (mpz_sgn(a) < 0)
            r = -r;

        return {r, exp};
    }

    std::vector<mpz_class> mpz_split_be(mpz_srcptr a, uint16_t bit_per_num, uint16_t target_out)
    {
        mpz_class x(a), tmp;
        mpz_class mask = (mpz_class(1) << bit_per_num) - 1;

        std::vector<mpz_class> out(target_out);

        if (mpz_sgn(x.get_mpz_t()) < 0)
            throw std::invalid_argument("mpz_split_be expects non-negative");

        for (int i = target_out - 1; i >= 0; i--)
        {
            mpz_and(tmp.get_mpz_t(), x.get_mpz_t(), mask.get_mpz_t());
            out[i] = tmp;
            mpz_fdiv_q_2exp(x.get_mpz_t(), x.get_mpz_t(), bit_per_num);
        }

        if (mpz_cmp_ui(x.get_mpz_t(), 0) != 0)
            throw std::overflow_error("mpz_split_be overflow");

        return out;
    }

    std::vector<uint8_t> mpz_to_big_endian_with_sign(mpz_srcptr a, size_t width)
    {
        struct context
        {
            mpz_t x, byte, ff;

            context()
            {
                mpz_inits(x, byte, ff, nullptr);
                mpz_set_ui(ff, 0xff);
            }
            ~context() { mpz_clears(x, byte, ff, nullptr); }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };

        static thread_local context ctx;

        std::vector<uint8_t> out(width + 1, 0);
        mpz_set(ctx.x, a);

        if (mpz_sgn(a) < 0)
        {
            out[0] = 1;
            mpz_neg(ctx.x, a);
        }

        for (size_t i = 0; i < width; ++i)
        {
            size_t idx = width - i;
            mpz_and(ctx.byte, ctx.x, ctx.ff);
            out[idx] = static_cast<uint8_t>(mpz_get_ui(ctx.byte));
            mpz_fdiv_q_2exp(ctx.x, ctx.x, 8);
        }

        if (mpz_sgn(ctx.x) != 0)
            throw std::invalid_argument("integer does not fit width");

        return out;
    }

    std::vector<uint8_t> mpz_to_big_endian(mpz_srcptr a, size_t width)
    {
        struct context
        {
            mpz_t x, byte, ff;

            context()
            {
                mpz_inits(x, byte, ff, nullptr);
                mpz_set_ui(ff, 0xff);
            }
            ~context() { mpz_clears(x, byte, ff, nullptr); }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };

        static thread_local context ctx;

        std::vector<uint8_t> out(width, 0);
        mpz_set(ctx.x, a);

        if (mpz_sgn(ctx.x) < 0)
            throw std::invalid_argument("to_big_endian expects non-negative");

        for (size_t i = 0; i < width; ++i)
        {
            size_t idx = width - i - 1;
            mpz_and(ctx.byte, ctx.x, ctx.ff);
            out[idx] = static_cast<uint8_t>(mpz_get_ui(ctx.byte));
            mpz_fdiv_q_2exp(ctx.x, ctx.x, 8);
        }

        if (mpz_sgn(ctx.x) != 0)
            throw std::invalid_argument("integer does not fit width");

        return out;
    }

    std::vector<uint8_t> mpz_to_big_endian_minimal(mpz_srcptr a)
    {
        struct context
        {
            mpz_t x, byte, ff;

            context()
            {
                mpz_inits(x, byte, ff, nullptr);
                mpz_set_ui(ff, 0xff);
            }
            ~context() { mpz_clears(x, byte, ff, nullptr); }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };

        static thread_local context ctx;

        mpz_set(ctx.x, a);

        if (mpz_sgn(ctx.x) < 0)
            throw std::invalid_argument("to_big_endian_minimal expects non-negative");

        if (mpz_sgn(ctx.x) == 0)
            return {0};

        std::vector<uint8_t> out;
        while (mpz_sgn(ctx.x) > 0)
        {
            mpz_and(ctx.byte, ctx.x, ctx.ff);
            out.push_back(static_cast<uint8_t>(mpz_get_ui(ctx.byte)));
            mpz_fdiv_q_2exp(ctx.x, ctx.x, 8);
        }
        std::reverse(out.begin(), out.end());
        return out;
    }
}
