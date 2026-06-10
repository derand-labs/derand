#include "common.hpp"
#include "vdf/system.hpp"
#include "hash/sha256.hpp"
#include "hash/poseidon2.hpp"

#include <fstream>
#include <iostream>

namespace vdf
{
    void to_json(nlohmann::json &j, const System &s)
    {
        j = nlohmann::json{
            {"d_bits", s.d_bits},
            {"D", common::mpz_get_hex(s.D)},
            {"l_bits", s.l_bits},
            {"limb_bits", s.limb_bits},
            {"split_exp", s.split_exp},
            {"hash_to_form_steps", s.hash_to_form_steps},
            {"hash_to_form_generators", s.hash_to_form_generators}};
    }

    void from_json(const nlohmann::json &j, System &s)
    {
        j.at("d_bits").get_to(s.d_bits);
        common::mpz_set_hex(s.D, j.at("D").get<std::string>());
        j.at("l_bits").get_to(s.l_bits);
        j.at("limb_bits").get_to(s.limb_bits);
        j.at("split_exp").get_to(s.split_exp);
        j.at("hash_to_form_steps").get_to(s.hash_to_form_steps);
        j.at("hash_to_form_generators").get_to(s.hash_to_form_generators);
    }

    static void pow2exp_mod(mpz_t rop, mpz_srcptr base, unsigned long k, mpz_srcptr mod)
    {
        mpz_set(rop, base);
        for (unsigned long i = 0; i < k; ++i)
        {
            mpz_mul(rop, rop, rop);
            mpz_mod(rop, rop, mod);
        }
    }

    static bool tonelli_shanks(mpz_t root, mpz_srcptr n, mpz_srcptr p)
    {
        if (mpz_cmp_ui(p, 2) == 0)
        {
            mpz_mod(root, n, p);
            return true;
        }

        mpz_t nmod, q, pminus1, z, c, t, r, tmp, exp, b;
        mpz_inits(nmod, q, pminus1, z, c, t, r, tmp, exp, b, nullptr);

        mpz_mod(nmod, n, p);
        if (mpz_sgn(nmod) == 0)
        {
            mpz_set_ui(root, 0);
            mpz_clears(nmod, q, pminus1, z, c, t, r, tmp, exp, b, nullptr);
            return true;
        }

        if (mpz_legendre(nmod, p) != 1)
        {
            mpz_clears(nmod, q, pminus1, z, c, t, r, tmp, exp, b, nullptr);
            return false;
        }

        mpz_sub_ui(pminus1, p, 1);
        mpz_set(q, pminus1);

        unsigned long s = 0;
        while (mpz_even_p(q))
        {
            mpz_fdiv_q_2exp(q, q, 1);
            ++s;
        }

        mpz_set_ui(z, 2);
        while (mpz_legendre(z, p) != -1)
        {
            mpz_add_ui(z, z, 1);
        }

        mpz_add_ui(exp, q, 1);
        mpz_fdiv_q_2exp(exp, exp, 1);
        mpz_powm(r, nmod, exp, p);

        mpz_powm(t, nmod, q, p);
        mpz_powm(c, z, q, p);

        unsigned long m = s;

        while (mpz_cmp_ui(t, 1) != 0)
        {
            unsigned long i = 0;
            mpz_set(tmp, t);

            while (mpz_cmp_ui(tmp, 1) != 0)
            {
                mpz_mul(tmp, tmp, tmp);
                mpz_mod(tmp, tmp, p);
                ++i;
                if (i == m)
                {
                    mpz_clears(nmod, q, pminus1, z, c, t, r, tmp, exp, b, nullptr);
                    return false;
                }
            }

            unsigned long e = m - i - 1;
            pow2exp_mod(b, c, e, p);

            mpz_mul(r, r, b);
            mpz_mod(r, r, p);

            mpz_mul(tmp, b, b);
            mpz_mod(tmp, tmp, p);

            mpz_mul(t, t, tmp);
            mpz_mod(t, t, p);

            mpz_set(c, tmp);
            m = i;
        }

        mpz_set(root, r);
        mpz_clears(nmod, q, pminus1, z, c, t, r, tmp, exp, b, nullptr);
        return true;
    }

    std::optional<classgroup::Form> find_entry_for_prime(mpz_srcptr D, mpz_srcptr p)
    {
        if (mpz_cmp_ui(p, 2) == 0)
        {
            if (mpz_fdiv_ui(D, 8) != 1)
            {
                return std::nullopt;
            }

            mpz_t b, c, num, fourp;
            mpz_inits(b, c, num, fourp, nullptr);

            mpz_set_ui(b, 1);
            mpz_mul_ui(fourp, p, 4);

            mpz_mul(num, b, b);
            mpz_sub(num, num, D);

            if (!mpz_divisible_p(num, fourp))
            {
                mpz_clears(b, c, num, fourp, nullptr);
                return std::nullopt;
            }

            mpz_fdiv_q(c, num, fourp);
            classgroup::Form cand(p, b, c);

            mpz_clears(b, c, num, fourp, nullptr);
            cand.reduce_inplace();
            return cand;
        }

        if (mpz_divisible_p(D, p))
        {
            mpz_t b, c, num, fourp;
            mpz_inits(b, c, num, fourp, nullptr);

            mpz_set(b, p);
            mpz_mul_ui(fourp, p, 4);

            mpz_mul(num, b, b);
            mpz_sub(num, num, D);

            if (!mpz_divisible_p(num, fourp))
            {
                mpz_clears(b, c, num, fourp, nullptr);
                return std::nullopt;
            }

            mpz_fdiv_q(c, num, fourp);
            classgroup::Form cand(p, b, c);

            mpz_clears(b, c, num, fourp, nullptr);
            cand.reduce_inplace();
            return cand;
        }

        if (mpz_legendre(D, p) != 1)
        {
            return std::nullopt;
        }

        mpz_t r, b, c, num, fourp;
        mpz_inits(r, b, c, num, fourp, nullptr);

        if (!tonelli_shanks(r, D, p))
        {
            mpz_clears(r, b, c, num, fourp, nullptr);
            return std::nullopt;
        }

        if (mpz_tstbit(r, 0))
        {
            mpz_set(b, r);
        }
        else
        {
            mpz_sub(b, p, r);
        }

        mpz_mul_ui(fourp, p, 4);

        mpz_mul(num, b, b);
        mpz_sub(num, num, D);

        if (!mpz_divisible_p(num, fourp))
        {
            mpz_clears(r, b, c, num, fourp, nullptr);
            return std::nullopt;
        }

        mpz_fdiv_q(c, num, fourp);

        classgroup::Form cand(p, b, c);

        mpz_clears(r, b, c, num, fourp, nullptr);
        cand.reduce_inplace();
        return cand;
    }

    std::string System::compute_system_id(const std::vector<uint8_t> &seed)
    {
        std::vector<uint8_t> payload;
        payload.insert(payload.end(), seed.begin(), seed.end());
        common::append_u16_be(payload, this->d_bits);
        common::append_u16_be(payload, this->l_bits);
        common::append_u16_be(payload, this->limb_bits);
        common::append_u16_be(payload, this->split_exp);
        common::append_u16_be(payload, this->hash_to_form_generators.size());
        common::append_u16_be(payload, this->hash_to_form_steps);
        return common::bytes_to_hex(hash::sha256(payload));
    }

    System::System()
    {
        mpz_init(D);
    }

    System::~System()
    {
        mpz_clear(D);
    }

    System::System(const System &other) : System()
    {
        mpz_set(D, other.D);
    }

    System &System::operator=(const System &other)
    {

        if (this == &other)
            return *this;

        mpz_set(D, other.D);

        return *this;
    }

    void printProcess(int n, int target, int attempt, int max_attempts, bool end)
    {
        static thread_local std::string prev_n_s, prev_attempt_s;

        std::string target_s = std::to_string(target);
        std::string max_attempt_s = std::to_string(max_attempts);

        if (prev_n_s.size() > 0 && prev_attempt_s.size() > 0)
        {
            for (size_t j = 0; j < prev_n_s.size() + target_s.size() + prev_attempt_s.size() + max_attempt_s.size() + 11; ++j)
                std::cout << "\b \b" << std::flush;
        }

        if (!end)
        {
            std::string n_s = std::to_string(n);
            std::string attempt_s = std::to_string(attempt);
            std::cout << n_s << "/" << target_s << " (tried " << attempt_s << "/" << max_attempt_s << ")" << std::flush;

            prev_n_s = n_s;
            prev_attempt_s = attempt_s;
        }
    }

    System::System(
        const std::vector<uint8_t> &seed,
        const uint16_t d_bits,
        const uint16_t l_bits,
        const uint16_t limb_bits,
        const uint16_t split_exp,
        const uint16_t hash_to_form_nb_generators,
        const uint16_t hash_to_form_steps) : System()
    {
        if (split_exp == 0 || l_bits % split_exp != 0)
        {
            throw std::invalid_argument("split_exp: it must divide l_bits exactly");
        }
        if (hash_to_form_nb_generators == 0)
        {
            throw std::invalid_argument("hash-to-form number generators must be > 0");
        }
        if (hash_to_form_steps == 0)
        {
            throw std::invalid_argument("hash-to-form steps must be > 0");
        }

        struct context
        {
            mpz_t candidate, signValue;
            context() { mpz_inits(candidate, signValue, nullptr); }
            ~context() { mpz_clears(candidate, signValue, nullptr); }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };
        static thread_local context ctx;

        this->d_bits = d_bits;
        classgroup::derive_discriminant(this->D, seed, d_bits);
        this->l_bits = l_bits;
        this->limb_bits = limb_bits;
        this->split_exp = split_exp;
        this->hash_to_form_steps = hash_to_form_steps;

        classgroup::Form::setup(this->D);

        this->hash_to_form_generators.reserve(static_cast<size_t>(hash_to_form_nb_generators));

        uint32_t attempt = 0;
        uint32_t max_attempts = hash_to_form_nb_generators * 1 << 16;
        printProcess(0, hash_to_form_nb_generators, 0, max_attempts, false);
        while (this->hash_to_form_generators.size() < hash_to_form_nb_generators && attempt < max_attempts)
        {
            std::vector<uint8_t> candidate_seed = seed;
            common::append_u32_be(candidate_seed, attempt++);
            hash::sha256_to_prime(ctx.candidate, candidate_seed, this->d_bits / 2, false, false);

            std::optional<classgroup::Form> entry = find_entry_for_prime(this->D, ctx.candidate);
            if (entry.has_value() && mpz_cmp_ui(entry.value().a, 1) != 0)
            {
                classgroup::Form x = entry.value();

                std::vector<uint8_t> signHash = hash::sha256(seed);
                common::mpz_set_from_be(ctx.signValue, signHash);

                classgroup::Form xNeg = x;
                mpz_neg(xNeg.b, xNeg.b);

                if (mpz_mod_ui(ctx.signValue, ctx.signValue, 2) == 1)
                {
                    mpz_neg(x.b, x.b);
                    mpz_neg(xNeg.b, xNeg.b);
                }

                bool include = true;
                for (const classgroup::Form generator : this->hash_to_form_generators)
                {
                    if (generator == x || generator == xNeg)
                    {
                        include = false;
                        break;
                    }
                }

                if (include)
                {
                    this->hash_to_form_generators.push_back(x);
                }
            }

            printProcess(this->hash_to_form_generators.size(), hash_to_form_nb_generators, attempt, max_attempts, false);
        }
        printProcess(this->hash_to_form_generators.size(), hash_to_form_nb_generators, attempt, max_attempts, true);

        if (this->hash_to_form_generators.size() < hash_to_form_nb_generators)
        {
            throw std::logic_error("not found enough generators, try another d-seed");
        }

        this->system_id = System::compute_system_id(seed);
    }

    std::string System::save(std::string dir) const
    {
        std::filesystem::path d(dir);
        std::filesystem::path p("system-" + this->system_id + ".json");
        p = d / p;

        if (p.has_parent_path())
        {
            std::filesystem::create_directories(p.parent_path());
        }

        std::ofstream file(p);
        if (!file.is_open())
            throw std::runtime_error("Cannot open file");

        nlohmann::json j = *this;
        file << j.dump(4);

        return p.string();
    }

    System System::load(const std::string dir, const std::string id)
    {
        std::filesystem::path d(dir);
        std::filesystem::path p("system-" + id + ".json");
        std::ifstream file(d / p);
        if (!file.is_open())
        {
            throw std::runtime_error("Cannot open system file: " + (d / p).string());
        }

        nlohmann::json j;
        file >> j;

        mpz_t D;
        mpz_init(D);
        common::mpz_set_hex(D, j.at("D").get<std::string>());
        classgroup::Form::setup(D);

        mpz_clear(D);

        return j.get<System>();
    }

    classgroup::Form hash_to_form(
        const System &system,
        const std::vector<uint8_t> &x_seed)
    {
        struct context
        {
            mpz_t seed_field, idx_big, idx_hash, entries_size;

            context() { mpz_inits(seed_field, idx_big, idx_hash, entries_size, nullptr); }
            ~context() { mpz_clears(seed_field, idx_big, idx_hash, entries_size, nullptr); }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };

        static thread_local context ctx;

        if (system.hash_to_form_generators.empty())
            throw std::invalid_argument("hash-to-form generators must not be empty");

        if (system.hash_to_form_steps == 0)
            throw std::invalid_argument("hash-to-form steps must be > 0");

        common::mpz_set_from_be(ctx.seed_field, x_seed);
        mpz_set_ui(ctx.entries_size, system.hash_to_form_generators.size());

        classgroup::Form acc = classgroup::Form::principal();

        for (uint32_t step = 0; step < system.hash_to_form_steps; ++step)
        {
            mpz_add_ui(ctx.seed_field, ctx.seed_field, 1);
            hash::poseidon2(ctx.idx_hash, {ctx.seed_field});

            mpz_mod(ctx.idx_big, ctx.idx_hash, ctx.entries_size);
            size_t idx = static_cast<size_t>(mpz_get_ui(ctx.idx_big));
            const classgroup::Form &selected = system.hash_to_form_generators[idx];

            acc.nucomp_inplace(selected);
        }

        acc.reduce_inplace();
        return acc;
    }

    void append_and_normalize_sign_for_seed(std::vector<uint8_t> &seed, mpz_ptr x)
    {
        if (mpz_sgn(x) < 0)
        {
            seed.push_back(1);
            mpz_neg(x, x);
        }
        else
        {
            seed.push_back(0);
        }
    }

    void split_and_append_num_for_seed(std::vector<uint8_t> &seed, mpz_srcptr x, size_t nbits, size_t limbbits)
    {
        std::vector<mpz_class> x128 = common::mpz_split_be(
            x,
            limbbits,
            (nbits + limbbits - 1) / limbbits);
        for (const mpz_class &x128limb : x128)
        {
            std::vector<uint8_t> b = common::mpz_to_big_endian(x128limb.get_mpz_t(), 16);
            seed.insert(seed.end(), b.begin(), b.end());
        }
    }

    std::vector<mpz_class>
    derive_challenge_l(
        mpz_ptr l,
        const System &system,
        const PublicStatement &stmt,
        const classgroup::Form &y,
        const bool export_transcript)
    {
        // X
        if (stmt.x_seed.size() > 32)
            throw std::invalid_argument("x_seed must be <= 32 bytes");

        std::vector<uint8_t> seed(32, 0);
        std::copy(stmt.x_seed.begin(), stmt.x_seed.end(), seed.end() - stmt.x_seed.size());

        // T
        common::append_u64_be(seed, stmt.T);

        // Y
        size_t small_bits = (system.d_bits + 1) / 2;

        mpz_class ya(y.a);
        append_and_normalize_sign_for_seed(seed, ya.get_mpz_t());
        split_and_append_num_for_seed(seed, ya.get_mpz_t(), small_bits, system.limb_bits);

        mpz_class yb(y.b);
        append_and_normalize_sign_for_seed(seed, yb.get_mpz_t());
        split_and_append_num_for_seed(seed, yb.get_mpz_t(), small_bits, system.limb_bits);

        mpz_class yc(y.c);
        append_and_normalize_sign_for_seed(seed, yc.get_mpz_t());
        split_and_append_num_for_seed(seed, yc.get_mpz_t(), system.d_bits, system.limb_bits);

        return hash::sha256_to_prime(l, seed, system.l_bits, true, export_transcript);
    }
} // namespace classgroup
