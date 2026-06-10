#include "common.hpp"
#include "hash/sha256.hpp"

#include <openssl/sha.h>

namespace hash
{
    void enforce_bit_length(std::vector<uint8_t> &bytes, size_t d_bits);

    std::vector<uint8_t> sha256(const std::vector<uint8_t> &data)
    {
        std::vector<uint8_t> out(SHA256_DIGEST_LENGTH);
        SHA256(data.data(), data.size(), out.data());
        return out;
    }

    std::vector<uint8_t> hash_expand_sha256(
        const std::vector<uint8_t> &seed,
        size_t out_bytes)
    {
        if (out_bytes <= 32)
        {
            std::vector<uint8_t> h = sha256(seed);
            h.resize(out_bytes);
            return h;
        }

        std::vector<uint8_t> out;
        out.reserve(out_bytes);
        uint32_t block = 0;
        while (out.size() < out_bytes)
        {
            std::vector<uint8_t> msg;
            common::append_u32_be(msg, block);
            msg.insert(msg.end(), seed.begin(), seed.end());
            std::vector<uint8_t> h = sha256(msg);
            size_t need = std::min(h.size(), out_bytes - out.size());
            out.insert(out.end(), h.begin(), h.begin() + static_cast<long>(need));
            ++block;
        }
        return out;
    }

    static const uint32_t SMALL_PRIMES[] = {
        3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97,
        101, 103, 107, 109, 113, 127, 131, 137, 139, 149, 151, 157, 163, 167, 173, 179, 181, 191,
        193, 197, 199, 211, 223, 227, 229, 233, 239, 241, 251, 257, 263, 269, 271, 277, 281, 283,
        293, 307, 311, 313, 317, 331, 337, 347, 349, 353, 359, 367, 373, 379, 383, 389, 397, 401,
        409, 419, 421, 431, 433, 439, 443, 449, 457, 461, 463, 467, 479, 487, 491, 499};

    static int trial_division(mpz_t factor, const mpz_t n)
    {
        mpz_t r;
        mpz_init(r);
        for (size_t i = 0; i < sizeof(SMALL_PRIMES) / sizeof(SMALL_PRIMES[0]); ++i)
        {
            uint32_t p = SMALL_PRIMES[i];
            mpz_mod_ui(r, n, p);
            if (mpz_cmp_ui(r, 0) == 0)
            {
                mpz_set_ui(factor, p);
                mpz_clear(r);
                return 1;
            }
        }
        mpz_clear(r);
        return 0;
    }

    // Pollard Rho - Brent variant
    static int pollard_rho_brent(mpz_t factor, const mpz_t n, gmp_randstate_t rs)
    {
        static const unsigned long MAX_OUTER_LOOPS = 14;
        static const unsigned long MAX_BACKTRACK_STEPS = 1UL << 10;

        if (mpz_even_p(n))
        {
            mpz_set_ui(factor, 2);
            return 1;
        }
        if (mpz_probab_prime_p(n, 12) > 0)
            return 0; // n prime => no non-trivial factor

        mpz_t y, c, m, g, r, q, x, ys, tmp, diff;
        mpz_inits(y, c, m, g, r, q, x, ys, tmp, diff, NULL);

        mpz_set_ui(m, 128); // batch size
        mpz_set_ui(g, 1);
        mpz_set_ui(r, 1);
        mpz_set_ui(q, 1);

        // random y in [1, n-1], c in [1, n-1]
        mpz_urandomm(y, rs, n);
        if (mpz_cmp_ui(y, 0) == 0)
            mpz_set_ui(y, 1);

        mpz_urandomm(c, rs, n);
        if (mpz_cmp_ui(c, 0) == 0)
            mpz_set_ui(c, 1);

        unsigned long outer_loops = 0;
        while (mpz_cmp_ui(g, 1) == 0)
        {
            if (++outer_loops > MAX_OUTER_LOOPS)
            {
                mpz_clears(y, c, m, g, r, q, x, ys, tmp, diff, NULL);
                return 0;
            }

            mpz_set(x, y);

            // y = f^r(y), f(v)=v^2+c mod n
            unsigned long rr = mpz_get_ui(r);
            for (unsigned long i = 0; i < rr; ++i)
            {
                mpz_mul(y, y, y);
                mpz_add(y, y, c);
                mpz_mod(y, y, n);
            }

            mpz_set_ui(q, 1);
            unsigned long k = 0;
            while (k < rr && mpz_cmp_ui(g, 1) == 0)
            {
                mpz_set(ys, y);
                unsigned long lim = (rr - k < mpz_get_ui(m)) ? (rr - k) : mpz_get_ui(m);

                for (unsigned long i = 0; i < lim; ++i)
                {
                    mpz_mul(y, y, y);
                    mpz_add(y, y, c);
                    mpz_mod(y, y, n);

                    mpz_sub(diff, x, y);
                    mpz_abs(diff, diff);

                    mpz_mul(q, q, diff);
                    mpz_mod(q, q, n);
                }

                mpz_gcd(g, q, n);
                k += lim;
            }

            mpz_mul_ui(r, r, 2);
        }

        if (mpz_cmp(g, n) == 0)
        {
            unsigned long backtrack_steps = 0;
            do
            {
                if (++backtrack_steps > MAX_BACKTRACK_STEPS)
                {
                    mpz_clears(y, c, m, g, r, q, x, ys, tmp, diff, NULL);
                    return 0;
                }

                mpz_mul(ys, ys, ys);
                mpz_add(ys, ys, c);
                mpz_mod(ys, ys, n);

                mpz_sub(diff, x, ys);
                mpz_abs(diff, diff);
                mpz_gcd(g, diff, n);
            } while (mpz_cmp_ui(g, 1) == 0);
        }

        if (mpz_cmp(g, n) == 0)
        {
            mpz_clears(y, c, m, g, r, q, x, ys, tmp, diff, NULL);
            return 0;
        }

        mpz_set(factor, g);
        mpz_clears(y, c, m, g, r, q, x, ys, tmp, diff, NULL);
        return 1;
    }

    int find_factor(mpz_t factor, const mpz_t n)
    {
        if (mpz_cmp_ui(n, 2) < 0)
            return 0;
        if (mpz_even_p(n))
        {
            mpz_set_ui(factor, 2);
            return 1;
        }
        if (mpz_probab_prime_p(n, 12) > 0)
            return 0;

        if (trial_division(factor, n))
            return 1;

        gmp_randstate_t rs;
        gmp_randinit_mt(rs);

        mpz_t seed;
        mpz_init(seed);

        // Fixed seed to ensure reproducible factorization results.
        mpz_set_ui(seed, (unsigned long)0x9e3779b97f4a7c15ULL);
        gmp_randseed(rs, seed);
        mpz_clear(seed);

        for (int i = 0; i < 4; ++i)
        {
            if (pollard_rho_brent(factor, n, rs))
            {
                if (mpz_cmp_ui(factor, 1) > 0 && mpz_cmp(factor, n) < 0)
                {
                    gmp_randclear(rs);
                    return 1;
                }
            }
        }

        gmp_randclear(rs);
        return 0;
    }

    std::vector<mpz_class> sha256_to_prime(
        mpz_ptr out,
        const std::vector<uint8_t> &seed,
        const size_t nbits,
        const bool ensurebits,
        const bool export_transcript)
    {
        static thread_local mpz_class one = mpz_class(1);
        std::vector<uint8_t> tmp = sha256(seed);

        mpz_class normalized_seed;
        common::mpz_set_from_be(normalized_seed.get_mpz_t(), tmp);

        const size_t out_bytes = (nbits + 7) / 8;
        std::vector<mpz_class> transcript;
        for (uint32_t attempt = 0;; ++attempt)
        {
            mpz_add_ui(normalized_seed.get_mpz_t(), normalized_seed.get_mpz_t(), 1);
            std::vector<uint8_t> attempt_seed = common::mpz_to_big_endian(normalized_seed.get_mpz_t(), 32);
            std::vector<uint8_t> bytes = hash::hash_expand_sha256(attempt_seed, out_bytes);

            if (ensurebits)
            {
                enforce_bit_length(bytes, nbits);
            }

            common::mpz_set_from_be(out, bytes);
            mpz_ior(out, out, one.get_mpz_t());

            if (ensurebits && common::mpz_bit_length(out) != nbits)
            {
                throw std::logic_error("must be nbits");
            }

            if (mpz_probab_prime_p(out, 12) > 0)
                break;

            if (export_transcript)
            {
                mpz_class factor;
                int found = find_factor(factor.get_mpz_t(), out);

                if (!found)
                    // If no factor is found for the composite number, this factor is
                    // set to zero so the verifier falls back to the Miller-Rabin.
                    transcript.push_back(mpz_class(0));
                else
                    transcript.push_back(factor);
            }
        }

        return transcript;
    }

    void sha256_to_prime_3mod4(mpz_ptr out, const std::vector<uint8_t> &seed, const size_t nbits)
    {
        static thread_local mpz_class three = mpz_class(3);

        const size_t out_bytes = (nbits + 7) / 8;
        for (uint32_t attempt = 0;; ++attempt)
        {
            std::vector<uint8_t> attempt_seed = seed;
            common::append_u32_be(attempt_seed, attempt);

            std::vector<uint8_t> bytes = hash::hash_expand_sha256(attempt_seed, out_bytes);

            enforce_bit_length(bytes, nbits);

            common::mpz_set_from_be(out, bytes);
            if (common::mpz_bit_length(out) != nbits)
                continue;

            mpz_ior(out, out, three.get_mpz_t());

            if (mpz_probab_prime_p(out, 12) > 0)
                return;
        }
    }

    void enforce_bit_length(std::vector<uint8_t> &bytes, size_t d_bits)
    {
        const uint32_t rem_bits = d_bits % 8;

        if (rem_bits == 0)
        {
            bytes[0] |= 0x80;
            return;
        }

        uint8_t mask = (1u << rem_bits) - 1u;

        bytes[0] &= mask;
        bytes[0] |= (1u << (rem_bits - 1));
    }
} // namespace classgroup
