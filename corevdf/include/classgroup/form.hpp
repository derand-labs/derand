#pragma once

#include <gmpxx.h>
#include <nlohmann/json.hpp>

namespace classgroup
{
    class Form
    {
    private:
        bool is_valid() const;

    public:
        mpz_t a;
        mpz_t b;
        mpz_t c;

        Form();
        ~Form();

        Form(const Form &other);
        Form &operator=(const Form &other);

        Form(mpz_srcptr _a, mpz_srcptr _b, mpz_srcptr _c);
        static Form from_ab(mpz_srcptr _a, mpz_srcptr _b);

        static mpz_class D;
        static mpz_class nucomp_bound;

        static void setup(mpz_srcptr D);
        static Form principal();

        bool operator==(const Form &other) const;

        bool is_primitive() const;

        void reduce_inplace();

        // Compose
        void compose_inplace(const Form &g, const bool reduce = true);

        // NUCOMP/NUDUPL
        void nucomp_inplace(const Form &g, const bool reduce = true);
        void nudupl_inplace();

        void fast_pow_inplace(mpz_srcptr exp);
        Form pow(mpz_srcptr exp) const;
        std::tuple<Form, Form> partial_pow(mpz_srcptr exp, size_t bits, Form prev_base) const;
    };

    void derive_discriminant(mpz_ptr d, const std::vector<uint8_t> seed, const size_t d_bits);

    void to_json(nlohmann::json &j, const Form &f);
    void from_json(const nlohmann::json &j, Form &f);
}
