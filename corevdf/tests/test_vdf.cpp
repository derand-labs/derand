#include "vdf/prover.hpp"
#include "vdf/verifier.hpp"

#include <doctest/doctest.h>

TEST_CASE("accepts valid Wesolowski proofs and rejects tampered proofs")
{
    vdf::System system(
        std::vector<uint8_t>{0x01, 0x02, 0x03, 0x04},
        1024,
        128,
        64,
        1,
        8,
        14);

    vdf::PublicStatement stmt{
        std::vector<uint8_t>{0x10, 0x20, 0x30, 0x40},
        1000,
    };

    classgroup::Form x = vdf::hash_to_form(system, stmt.x_seed);
    vdf::VdfOutput eval = vdf::evaluate_vdf(system, stmt);

    vdf::VdfProof proof = vdf::prove_wesolowski(system, stmt, eval);
    CHECK(vdf::verify_wesolowski(system, stmt, proof).ok);

    vdf::VdfProof tampered_proof = proof;
    tampered_proof.pi.nucomp_inplace(x);
    CHECK_FALSE(vdf::verify_wesolowski(system, stmt, tampered_proof).ok);

    for (uint32_t t_case : {64u, 4096u, 16384u})
    {
        vdf::PublicStatement integ_stmt{
            std::vector<uint8_t>{0x10, 0x20, 0x30, 0x40},
            t_case,
        };
        vdf::VdfOutput vdf_output = vdf::evaluate_vdf(system, integ_stmt);
        vdf::VdfProof integ_proof = vdf::prove_wesolowski(system, integ_stmt, vdf_output);
        CHECK(vdf::verify_wesolowski(system, integ_stmt, integ_proof).ok);
    }
}
