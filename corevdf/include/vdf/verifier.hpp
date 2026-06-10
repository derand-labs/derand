#pragma once

#include "vdf/prover.hpp"

namespace vdf
{
    struct VdfVerifyTranscriptIntermediatePow
    {
        mpz_class l;
        classgroup::Form pil;
        classgroup::Form pil_base;

        mpz_class r;
        classgroup::Form xr;
        classgroup::Form xr_base;
    };

    struct VdfVerifyTranscript
    {
        std::string system_id;
        uint64_t T;

        bool ok;
        mpz_class x_seed;

        classgroup::Form x;
        classgroup::Form y;
        classgroup::Form pi;

        std::vector<VdfVerifyTranscriptIntermediatePow> intermediate_pows;

        std::vector<mpz_class> challenge_l_transcript;
    };

    void to_json(nlohmann::json &j, const VdfVerifyTranscriptIntermediatePow &s);
    void to_json(nlohmann::json &j, const VdfVerifyTranscript &s);

    VdfVerifyTranscript verify_wesolowski(
        const System &system,
        const PublicStatement &stmt,
        const VdfProof &proof);
} // namespace classgroup
