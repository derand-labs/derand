#pragma once

#include "vdf/system.hpp"

namespace vdf
{
    struct VdfOutput
    {
        classgroup::Form y;
        std::vector<classgroup::Form> intermediates;
        uint64_t stride = 0;
        uint32_t k = 0;
        uint32_t l = 0;
    };

    struct VdfProof
    {
        classgroup::Form y;
        classgroup::Form pi;
    };

    void to_json(nlohmann::json &j, const VdfProof &s);
    void from_json(const nlohmann::json &j, VdfProof &s);

    VdfOutput evaluate_vdf(const System &system, const PublicStatement &stmt);

    VdfProof prove_wesolowski(
        const System &system,
        const PublicStatement &stmt,
        const VdfOutput &vdf_output);
} // namespace classgroup
