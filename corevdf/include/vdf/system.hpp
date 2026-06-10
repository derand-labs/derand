#pragma once

#include "classgroup/form.hpp"

#include <nlohmann/json.hpp>

namespace vdf
{
    struct PublicStatement
    {
        std::vector<uint8_t> x_seed;
        uint64_t T;
    };

    struct System
    {
        std::string system_id;
        uint16_t d_bits;
        mpz_t D;
        uint16_t l_bits;
        uint16_t limb_bits;
        uint16_t split_exp;
        uint16_t hash_to_form_steps;
        std::vector<classgroup::Form> hash_to_form_generators;

        System();
        ~System();

        System(const System &other);
        System &operator=(const System &other);

        System(
            const std::vector<uint8_t> &seed,
            const uint16_t d_bits,
            const uint16_t l_bits,
            const uint16_t limb_bits,
            const uint16_t split_exp,
            const uint16_t hash_to_form_nb_generators,
            const uint16_t hash_to_form_steps);

        static System load(std::string dir, std::string id);
        std::string save(std::string dir = ".system") const;

    private:
        std::string compute_system_id(const std::vector<uint8_t> &seed);
    };

    classgroup::Form hash_to_form(const System &system, const std::vector<uint8_t> &x_seed);

    std::vector<mpz_class> derive_challenge_l(
        mpz_ptr l,
        const System &system,
        const PublicStatement &stmt,
        const classgroup::Form &y,
        const bool export_transcript = false);

    void to_json(nlohmann::json &j, const System &s);
    void from_json(const nlohmann::json &j, System &s);
} // namespace classgroup
