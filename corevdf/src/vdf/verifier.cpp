#include "common.hpp"
#include "vdf/verifier.hpp"

namespace vdf
{

    void to_json(nlohmann::json &j, const VdfVerifyTranscriptIntermediatePow &s)
    {
        j = nlohmann::json{
            {"l", common::mpz_get_hex(s.l.get_mpz_t())},
            {"pil", s.pil},
            {"pil_base", s.pil_base},
            {"r", common::mpz_get_hex(s.r.get_mpz_t())},
            {"xr", s.xr},
            {"xr_base", s.xr_base}};
    }

    void to_json(nlohmann::json &j, const VdfVerifyTranscript &s)
    {
        std::vector<std::string> challenge_l_transcript;
        for (const mpz_class a : s.challenge_l_transcript)
        {
            challenge_l_transcript.push_back(common::mpz_get_hex(a.get_mpz_t()));
        }

        j = nlohmann::json{
            {"system_id", s.system_id},
            {"T", s.T},
            {"ok", s.ok},
            {"x_seed", common::mpz_get_hex(s.x_seed.get_mpz_t())},
            {"x", s.x},
            {"y", s.y},
            {"pi", s.pi},
            {"intermediate_pows", s.intermediate_pows},
            {"challenge_l_transcript", challenge_l_transcript}};
    }

    VdfVerifyTranscript verify_wesolowski(
        const System &system,
        const PublicStatement &stmt,
        const VdfProof &proof)
    {
        struct context
        {
            mpz_t x, l, r, two, expmask, exppart;

            context()
            {
                mpz_inits(x, l, r, two, expmask, exppart, nullptr);
                mpz_set_ui(two, 2);
            }
            ~context() { mpz_clears(x, l, r, two, expmask, exppart, nullptr); }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };

        static thread_local context ctx;

        int size_per_sub_exp = system.l_bits / system.split_exp;

        mpz_ui_pow_ui(ctx.expmask, 2, size_per_sub_exp);
        mpz_sub_ui(ctx.expmask, ctx.expmask, 1);

        classgroup::Form x = hash_to_form(system, stmt.x_seed);
        std::vector<mpz_class> challenge_l_transcript = derive_challenge_l(ctx.l, system, stmt, proof.y, true);
        mpz_powm_ui(ctx.r, ctx.two, stmt.T, ctx.l);

        std::vector<VdfVerifyTranscriptIntermediatePow> intermediate_pows(system.split_exp);
        for (int i = 0; i < system.split_exp; i++)
        {
            // part = x & mask
            mpz_and(ctx.exppart, ctx.l, ctx.expmask);
            intermediate_pows[i].l = mpz_class(ctx.exppart);

            mpz_and(ctx.exppart, ctx.r, ctx.expmask);
            intermediate_pows[i].r = mpz_class(ctx.exppart);

            // x >>= size_per_sub_exp
            mpz_fdiv_q_2exp(ctx.l, ctx.l, size_per_sub_exp);
            mpz_fdiv_q_2exp(ctx.r, ctx.r, size_per_sub_exp);
        }

        classgroup::Form pil = classgroup::Form::principal();
        classgroup::Form xr = classgroup::Form::principal();
        classgroup::Form pil_base = proof.pi;
        classgroup::Form xr_base = x;
        for (int i = 0; i < system.split_exp; i++)
        {
            std::tie(pil, pil_base) = pil.partial_pow(
                intermediate_pows[i].l.get_mpz_t(), size_per_sub_exp, pil_base);
            std::tie(xr, xr_base) = xr.partial_pow(
                intermediate_pows[i].r.get_mpz_t(), size_per_sub_exp, xr_base);

            intermediate_pows[i].pil = pil;
            intermediate_pows[i].xr = xr;
            intermediate_pows[i].pil_base = pil_base;
            intermediate_pows[i].xr_base = xr_base;
        }

        classgroup::Form rhs = pil;
        rhs.nucomp_inplace(xr);

        common::mpz_set_from_be(ctx.x, stmt.x_seed);

        bool ok = rhs == proof.y;

        return VdfVerifyTranscript{
            system.system_id,
            stmt.T,
            ok,
            mpz_class(ctx.x),
            x,
            proof.y,
            proof.pi,
            intermediate_pows,
            challenge_l_transcript};
    }
} // namespace classgroup
