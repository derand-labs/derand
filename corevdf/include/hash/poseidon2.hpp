#pragma once

#include <vector>
#include <gmp.h>

namespace hash
{
    void poseidon2(mpz_ptr out, const std::vector<mpz_srcptr> &inputs);
}
