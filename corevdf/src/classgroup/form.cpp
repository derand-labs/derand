#include "common.hpp"
#include "classgroup/form.hpp"
#include "hash/sha256.hpp"

namespace classgroup
{
    mpz_class Form::D = 0;
    mpz_class Form::nucomp_bound = 0;

    Form::Form()
    {
        mpz_inits(a, b, c, nullptr);
    }

    Form::~Form()
    {
        mpz_clears(a, b, c, nullptr);
    }

    Form::Form(const Form &other) : Form()
    {
        mpz_set(a, other.a);
        mpz_set(b, other.b);
        mpz_set(c, other.c);
    }

    Form &Form::operator=(const Form &other)
    {
        if (this == &other)
            return *this;

        mpz_set(a, other.a);
        mpz_set(b, other.b);
        mpz_set(c, other.c);

        return *this;
    }

    Form::Form(mpz_srcptr _a, mpz_srcptr _b, mpz_srcptr _c) : Form()
    {
        mpz_set(a, _a);
        mpz_set(b, _b);
        mpz_set(c, _c);
        if (!is_valid())
        {
            throw std::invalid_argument("invalid form");
        }
    }

    Form Form::from_ab(mpz_srcptr _a, mpz_srcptr _b)
    {
        struct context
        {
            mpz_t _c, b2, b2subd, a4;

            context() { mpz_inits(_c, b2, b2subd, a4, nullptr); }
            ~context() { mpz_clears(_c, b2, b2subd, a4, nullptr); }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };

        static thread_local context ctx;

        mpz_mul(ctx.b2, _b, _b);
        mpz_sub(ctx.b2subd, ctx.b2, Form::D.get_mpz_t());
        mpz_mul_ui(ctx.a4, _a, 4);

        mpz_fdiv_q(ctx._c, ctx.b2subd, ctx.a4);

        Form f = Form();

        mpz_set(f.a, _a);
        mpz_set(f.b, _b);
        mpz_set(f.c, ctx._c);

        return f;
    }

    void Form::setup(mpz_srcptr _D)
    {
        mpz_set(D.get_mpz_t(), _D);

        mpz_abs(nucomp_bound.get_mpz_t(), D.get_mpz_t());
        mpz_root(nucomp_bound.get_mpz_t(), nucomp_bound.get_mpz_t(), 4);

        if (mpz_cmp_ui(nucomp_bound.get_mpz_t(), 1) <= 0)
            mpz_set_ui(nucomp_bound.get_mpz_t(), 2);
    }

    Form Form::principal()
    {
        struct context
        {
            mpz_t one, four, c;
            context()
            {
                mpz_inits(one, four, c, nullptr);
                mpz_set_ui(one, 1);
                mpz_set_ui(four, 4);
            }
            ~context() { mpz_clears(one, four, c, nullptr); }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };

        static context s;

        mpz_sub(s.c, s.one, Form::D.get_mpz_t());
        mpz_fdiv_q(s.c, s.c, s.four);

        return Form(s.one, s.one, s.c);
    }

    bool Form::operator==(const Form &other) const
    {
        return mpz_cmp(a, other.a) == 0 && mpz_cmp(b, other.b) == 0 && mpz_cmp(c, other.c) == 0;
    }

    bool Form::is_primitive() const
    {
        struct context
        {
            mpz_t g;
            context() { mpz_inits(g, nullptr); }
            ~context() { mpz_clears(g, nullptr); }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };

        static thread_local context ctx;

        mpz_gcd(ctx.g, a, b);

        if (mpz_cmp_ui(ctx.g, 1) != 0)
            return false;

        mpz_gcd(ctx.g, ctx.g, c);
        return mpz_cmp_ui(ctx.g, 1) == 0;
    }

    static inline void calc_uvwx(
        int_fast64_t &u,
        int_fast64_t &v,
        int_fast64_t &w,
        int_fast64_t &x,
        int_fast64_t &a,
        int_fast64_t &b,
        int_fast64_t &c)
    {
        int below_threshold;
        int_fast64_t u_{1}, v_{0}, w_{0}, x_{1};
        int_fast64_t a_, b_, s;
        do
        {
            u = u_;
            v = v_;
            w = w_;
            x = x_;

            s = b >= 0 ? (b + c) / (c << 1) : -(-b + c) / (c << 1);

            a_ = a;
            b_ = b;

            a = c;
            b = -b + (uint_fast64_t(c * s) << 1);
            c = a_ - s * (b_ - c * s);

            u_ = v;
            v_ = -u + s * v;
            w_ = x;
            x_ = -w + s * x;

            below_threshold = (llabs(v_) | llabs(x_)) <= (1ul << 31) ? 1 : 0;
        } while (below_threshold && a > c && c > 0);

        if (below_threshold)
        {
            u = u_;
            v = v_;
            w = w_;
            x = x_;
        }
    }

    void Form::reduce_inplace()
    {
        struct context
        {
            mpz_t r, s, m;
            mpz_t faa, fab, fac, fba, fbb, fbc, fca, fcb, fcc;

            context() { mpz_inits(r, s, m, faa, fab, fac, fba, fbb, fbc, fca, fcb, fcc, nullptr); }
            ~context() { mpz_clears(r, s, m, faa, fab, fac, fba, fbb, fbc, fca, fcb, fcc, nullptr); }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };
        static thread_local context ctx;

        while (true)
        {
            int a_b = mpz_cmpabs(this->a, this->b);
            int c_b = mpz_cmpabs(this->c, this->b);
            if (a_b >= 0 && c_b >= 0)
            {
                int a_c = mpz_cmp(this->a, this->c);
                if (a_c > 0)
                {
                    mpz_swap(this->a, this->c);
                    mpz_neg(this->b, this->b);
                }
                else if (a_c == 0 && mpz_sgn(this->b) < 0)
                {
                    mpz_neg(this->b, this->b);
                }
                return;
            }

            auto [a, a_exp] = common::mpz_get_si_2exp(this->a);
            auto [b, b_exp] = common::mpz_get_si_2exp(this->b);
            auto [c, c_exp] = common::mpz_get_si_2exp(this->c);
            auto mm = std::minmax({a_exp, b_exp, c_exp});
            if (mm.second - mm.first > 31)
            {
                mpz_fdiv_q(ctx.r, this->b, this->c);
                mpz_add_ui(ctx.r, ctx.r, 1);
                mpz_div_2exp(ctx.s, ctx.r, 1);
                mpz_mul(ctx.m, this->c, ctx.s);
                mpz_mul_2exp(ctx.r, ctx.m, 1);
                mpz_sub(ctx.m, ctx.m, this->b);

                mpz_sub(this->b, ctx.r, this->b);
                mpz_swap(this->a, this->c);
                mpz_addmul(this->c, ctx.s, ctx.m);
                continue;
            }

            int_fast64_t max_exp(++mm.second);
            a >>= (max_exp - a_exp);
            b >>= (max_exp - b_exp);
            c >>= (max_exp - c_exp);

            int_fast64_t u, v, w, x;
            calc_uvwx(u, v, w, x, a, b, c);

            mpz_mul_si(ctx.faa, this->a, u * u);
            mpz_mul_si(ctx.fab, this->b, u * w);
            mpz_mul_si(ctx.fac, this->c, w * w);

            mpz_mul_si(ctx.fba, this->a, int_fast64_t(u * v) << 1);
            mpz_mul_si(ctx.fbb, this->b, u * x + v * w);
            mpz_mul_si(ctx.fbc, this->c, int_fast64_t(w * x) << 1);

            mpz_mul_si(ctx.fca, this->a, v * v);
            mpz_mul_si(ctx.fcb, this->b, v * x);
            mpz_mul_si(ctx.fcc, this->c, x * x);

            mpz_add(this->a, ctx.faa, ctx.fab);
            mpz_add(this->a, this->a, ctx.fac);

            mpz_add(this->b, ctx.fba, ctx.fbb);
            mpz_add(this->b, this->b, ctx.fbc);

            mpz_add(this->c, ctx.fca, ctx.fcb);
            mpz_add(this->c, this->c, ctx.fcc);
        }
    }

    bool Form::is_valid() const
    {
        struct context
        {
            mpz_t b2, ac4, d;
            context() { mpz_inits(b2, ac4, d, nullptr); }
            ~context() { mpz_clears(b2, ac4, d, nullptr); }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };

        static thread_local context ctx;

        if (mpz_sgn(a) <= 0 || mpz_sgn(c) <= 0)
            return false;

        mpz_mul(ctx.b2, b, b);
        mpz_mul(ctx.ac4, a, c);
        mpz_mul_ui(ctx.ac4, ctx.ac4, 4);
        mpz_sub(ctx.d, ctx.b2, ctx.ac4);

        if (mpz_cmp(ctx.d, D.get_mpz_t()) != 0)
            return false;

        if (!is_primitive())
            return false;

        return true;
    }

    void derive_discriminant(mpz_ptr d, const std::vector<uint8_t> seed, const size_t d_bits)
    {
        hash::sha256_to_prime_3mod4(d, seed, d_bits);
        mpz_neg(d, d);
    }

    void to_json(nlohmann::json &j, const Form &f)
    {
        j = {
            {"a", common::mpz_get_hex(f.a)},
            {"b", common::mpz_get_hex(f.b)},
            {"c", common::mpz_get_hex(f.c)},
        };
    }

    void from_json(const nlohmann::json &j, Form &f)
    {
        common::mpz_set_hex(f.a, j.at("a").get<std::string>());
        common::mpz_set_hex(f.b, j.at("b").get<std::string>());
        common::mpz_set_hex(f.c, j.at("c").get<std::string>());
    }
} // namespace classgroup
