#include "common.hpp"
#include "vdf/prover.hpp"

namespace vdf
{

    void to_json(nlohmann::json &j, const VdfProof &s)
    {
        j = nlohmann::json{
            {"y", s.y},
            {"pi", s.pi},
        };
    }

    void from_json(const nlohmann::json &j, VdfProof &s)
    {
        j.at("y").get_to(s.y);
        j.at("pi").get_to(s.pi);
    }

    namespace
    {
        struct FastProverParams
        {
            uint32_t k = 1;
            uint32_t l = 1;
            uint64_t stride = 1;
        };

        FastProverParams approximate_parameters(uint64_t t)
        {
            FastProverParams params;
            if (t == 0)
            {
                return params;
            }

            const double log_memory = 23.25349666;
            const double log_t = std::log2(static_cast<double>(t));

            if (log_t - log_memory > 1e-6)
            {
                const double l_estimate = std::ceil(std::pow(2.0, log_memory - 20.0));
                if (std::isfinite(l_estimate) && l_estimate > 0.0)
                {
                    params.l = static_cast<uint32_t>(l_estimate);
                    if (params.l == 0)
                    {
                        params.l = 1;
                    }
                }
            }

            const double intermediate =
                static_cast<double>(t) * 0.6931471 / (2.0 * static_cast<double>(params.l));
            if (intermediate <= 1.0)
            {
                params.k = 1;
            }
            else
            {
                const double k_estimate =
                    std::max(std::round(std::log(intermediate) - std::log(std::log(intermediate)) + 0.25), 1.0);
                if (std::isfinite(k_estimate) && k_estimate > 0.0)
                {
                    params.k = static_cast<uint32_t>(k_estimate);
                    if (params.k == 0)
                    {
                        params.k = 1;
                    }
                }
            }

            if (params.k > 20)
            {
                params.k = 20;
            }

            const uint64_t kl = static_cast<uint64_t>(params.k) * static_cast<uint64_t>(params.l);
            params.stride = kl == 0 ? 1 : kl;
            return params;
        }

        uint64_t get_block(uint64_t i, uint32_t k, uint64_t t, mpz_srcptr b)
        {
            struct context
            {
                mpz_t two, res;

                context()
                {
                    mpz_inits(two, res, nullptr);
                    mpz_set_ui(two, 2);
                }
                ~context() { mpz_clears(two, res, nullptr); }
                context(const context &) = delete;
                context &operator=(const context &) = delete;
            };

            static thread_local context ctx;

            const uint64_t exp = t - static_cast<uint64_t>(k) * (i + 1);

            mpz_powm_ui(ctx.res, ctx.two, exp, b);
            mpz_mul_2exp(ctx.res, ctx.res, k);
            mpz_fdiv_q(ctx.res, ctx.res, b);
            return mpz_get_ui(ctx.res);
        }
    } // namespace

    VdfOutput evaluate_vdf(
        const System &system,
        const PublicStatement &stmt)
    {
        VdfOutput out;

        classgroup::Form cur = hash_to_form(system, stmt.x_seed);
        const FastProverParams params = approximate_parameters(stmt.T);
        out.k = params.k;
        out.l = params.l;
        out.stride = params.stride;
        const uint64_t count = (stmt.T + params.stride - 1) / params.stride;
        out.intermediates.reserve(static_cast<size_t>(count));

        for (uint64_t i = 0; i < stmt.T; ++i)
        {
            if (i % params.stride == 0)
                out.intermediates.push_back(cur);

            cur.nudupl_inplace();
        }

        out.y = cur;
        return out;
    }

    classgroup::Form generate_wesolowski_fast(
        const std::vector<classgroup::Form> &intermediates,
        uint64_t t,
        uint32_t k,
        uint32_t l,
        mpz_srcptr challenge_l)
    {
        struct context
        {
            mpz_t one, step_exp, e;

            context()
            {
                mpz_inits(one, step_exp, e, nullptr);
                mpz_set_ui(one, 1);
            }
            ~context() { mpz_clears(one, step_exp, e, nullptr); }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };

        static thread_local context ctx;

        if (k == 0 || l == 0)
            throw std::invalid_argument("invalid k/l for fast wesolowski");

        if (k > 20)
            throw std::invalid_argument("k too large for fast wesolowski");

        const uint64_t stride = static_cast<uint64_t>(k) * static_cast<uint64_t>(l);
        const uint64_t expected = (t + stride - 1) / stride;
        if (intermediates.size() != expected)
            throw std::invalid_argument("sparse intermediates size mismatch");

        const uint64_t k1 = k / 2;
        const uint64_t k0 = k - k1;
        const uint64_t bucket_count = 1ULL << k;
        const uint64_t low_count = 1ULL << k0;

        classgroup::Form acc = classgroup::Form::principal();
        std::vector<classgroup::Form> ys(bucket_count);

        for (int64_t j = static_cast<int64_t>(l) - 1; j >= 0; --j)
        {
            for (classgroup::Form &y : ys)
                y = classgroup::Form::principal();

            mpz_mul_2exp(ctx.step_exp, ctx.one, k);
            acc.fast_pow_inplace(ctx.step_exp);

            for (uint64_t i = 0; i < expected; ++i)
            {
                const uint64_t idx = i * static_cast<uint64_t>(l) + static_cast<uint64_t>(j);
                if (t >= static_cast<uint64_t>(k) * (idx + 1))
                {
                    const uint64_t bucket = get_block(idx, k, t, challenge_l);
                    ys[bucket].nucomp_inplace(intermediates[i]);
                }
            }

            for (uint64_t b1 = 0; b1 < (1ULL << k1); ++b1)
            {
                classgroup::Form z = classgroup::Form::principal();
                for (uint64_t b0 = 0; b0 < low_count; ++b0)
                    z.nucomp_inplace(ys[b1 * low_count + b0]);

                mpz_set_ui(ctx.e, static_cast<unsigned long>(b1 * low_count));
                z.fast_pow_inplace(ctx.e);
                acc.nucomp_inplace(z);
            }

            for (uint64_t b0 = 0; b0 < low_count; ++b0)
            {
                classgroup::Form z = classgroup::Form::principal();
                for (uint64_t b1 = 0; b1 < (1ULL << k1); ++b1)
                    z.nucomp_inplace(ys[b1 * low_count + b0]);

                mpz_set_ui(ctx.e, static_cast<unsigned long>(b0));
                z.fast_pow_inplace(ctx.e);
                acc.nucomp_inplace(z);
            }
        }

        acc.reduce_inplace();
        return acc;
    }

    VdfProof prove_wesolowski(
        const System &system,
        const PublicStatement &stmt,
        const VdfOutput &vdf_output)
    {
        struct context
        {
            mpz_t l, r, two;

            context()
            {
                mpz_inits(l, r, two, nullptr);
                mpz_set_ui(two, 2);
            }
            ~context() { mpz_clears(l, r, two, nullptr); }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };

        static thread_local context ctx;

        derive_challenge_l(ctx.l, system, stmt, vdf_output.y);
        mpz_powm_ui(ctx.r, ctx.two, stmt.T, ctx.l);

        classgroup::Form pi;
        pi = generate_wesolowski_fast(
            vdf_output.intermediates,
            stmt.T,
            vdf_output.k,
            vdf_output.l,
            ctx.l);

        return VdfProof{vdf_output.y, pi};
    }
} // namespace classgroup
