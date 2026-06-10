#include "common.hpp"
#include "classgroup/form.hpp"

namespace classgroup
{
    static void fast_xgcd_partial(mpz_t co2, mpz_t co1, mpz_t r2, mpz_t r1, const mpz_t L);
    static void gcdinv(mpz_ptr g, mpz_ptr inv, mpz_srcptr x, mpz_srcptr m);

    void Form::compose_inplace(const Form &f2, const bool reduce)
    {
        mpz_t tmp;

        mpz_t halfbsum, g1, A;
        mpz_t m1, m2, g, m1p, m2p, diff, inv, tcrt, x0, lcm12;
        mpz_t M, k, rhs, s, q, t, den;

        mpz_inits(tmp,
                  halfbsum, g1, A,
                  m1, m2, g, m1p, m2p, diff, inv, tcrt, x0, lcm12,
                  M, k, rhs, s, q, t, den,
                  NULL);

        // halfbsum = (b1 + b2)/2
        mpz_add(halfbsum, this->b, f2.b);
        mpz_divexact_ui(halfbsum, halfbsum, 2);

        // g1 = gcd(a, a2, halfbsum)
        mpz_gcd(g1, this->a, f2.a);
        mpz_gcd(g1, g1, halfbsum);

        // A = a*a2 / g1^2
        mpz_mul(A, this->a, f2.a);
        mpz_mul(tmp, g1, g1);
        mpz_divexact(A, A, tmp);

        // m1 = 2a1/g1, m2 = 2a2/g1
        mpz_divexact(m1, this->a, g1);
        mpz_add(m1, m1, m1);

        mpz_divexact(m2, f2.a, g1);
        mpz_add(m2, m2, m2);

        // ==== CRT: x ≡ b (mod m1), x ≡ b2 (mod m2) ====
        mpz_gcd(g, m1, m2);

        mpz_sub(diff, f2.b, this->b);

        mpz_divexact(m1p, m1, g);
        mpz_divexact(m2p, m2, g);
        mpz_divexact(diff, diff, g);

        mpz_invert(inv, m1p, m2p);

        mpz_mul(tcrt, diff, inv);
        mpz_mod(tcrt, tcrt, m2p);

        mpz_mul(x0, m1, tcrt);
        mpz_add(x0, x0, this->b);

        mpz_mul(lcm12, m1, m2p);
        mpz_mod(x0, x0, lcm12);

        // ==== Solve main congruence ====

        mpz_mul_ui(M, A, 2);           // M = 2A
        mpz_divexact(k, halfbsum, g1); // k = Bmu / e

        // rhs = (D + b*b2)/(2e)
        mpz_mul(rhs, this->b, f2.b);
        mpz_add(rhs, rhs, Form::D.get_mpz_t());
        mpz_mul_ui(den, g1, 2);
        mpz_divexact(rhs, rhs, den);

        // s = (rhs - k*x0)/lcm12
        mpz_mul(tmp, k, x0);
        mpz_sub(tmp, rhs, tmp);

        mpz_divexact(s, tmp, lcm12);
        mpz_divexact(q, M, lcm12);

        // Solve k*t ≡ s (mod q)
        mpz_gcd(g, k, q);

        mpz_divexact(k, k, g);
        mpz_divexact(s, s, g);
        mpz_divexact(q, q, g);

        if (mpz_cmp_ui(q, 1) == 0)
        {
            mpz_set_ui(t, 0);
        }
        else
        {
            mpz_invert(inv, k, q);
            mpz_mul(t, s, inv);
            mpz_mod(t, t, q);
        }

        // B = x0 + lcm12 * t
        mpz_mul(this->b, lcm12, t);
        mpz_add(this->b, this->b, x0);
        mpz_mod(this->b, this->b, M);

        // C = (B^2 - D)/(4A)
        mpz_mul(this->c, this->b, this->b);
        mpz_sub(this->c, this->c, Form::D.get_mpz_t());
        mpz_mul_ui(tmp, A, 4);
        mpz_divexact(this->c, this->c, tmp);

        // a = A
        mpz_set(this->a, A);

        mpz_clears(tmp,
                   halfbsum, g1, A,
                   m1, m2, g, m1p, m2p, diff, inv, tcrt, x0, lcm12,
                   M, k, rhs, s, q, t, den,
                   NULL);

        if (reduce)
            this->reduce_inplace();
    }

    void Form::nucomp_inplace(const Form &g, const bool reduce)
    {
        struct context
        {
            mpz_t a1, a2, c2, ss, m, t, sp, v1, s, v2, u2, k;
            mpz_t m1, m2, r1, r2, co1, co2, temp, outa, outb, outc;
            mpz_t b2, b2subd, a4;
            context()
            {
                mpz_inits(a1, a2, c2, ss, m, t, sp, v1, s, v2, u2, k, nullptr);
                mpz_inits(m1, m2, r1, r2, co1, co2, temp, outa, outb, outc, nullptr);
                mpz_inits(b2, b2subd, a4, nullptr);
            }
            ~context()
            {
                mpz_clears(a1, a2, c2, ss, m, t, sp, v1, s, v2, u2, k, nullptr);
                mpz_clears(m1, m2, r1, r2, co1, co2, temp, outa, outb, outc, nullptr);
                mpz_clears(b2, b2subd, a4, nullptr);
            }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };

        static thread_local context ctx;

        const Form *f = this;
        const Form *h = &g;
        if (this->a > g.a)
        {
            f = &g;
            h = this;
        }

        mpz_set(ctx.a1, f->a);
        mpz_set(ctx.a2, h->a);
        mpz_set(ctx.c2, h->c);

        mpz_add(ctx.ss, f->b, h->b);
        mpz_divexact_ui(ctx.ss, ctx.ss, 2);

        mpz_sub(ctx.m, f->b, h->b);
        mpz_divexact_ui(ctx.m, ctx.m, 2);

        mpz_set(ctx.t, ctx.a2);
        common::mpz_mod_pos(ctx.t, ctx.t, ctx.a1);
        if (mpz_sgn(ctx.t) == 0)
        {
            mpz_set_ui(ctx.v1, 0);
            mpz_set(ctx.sp, ctx.a1);
        }
        else
        {
            gcdinv(ctx.sp, ctx.v1, ctx.t, ctx.a1);
        }

        mpz_mul(ctx.k, ctx.m, ctx.v1);
        common::mpz_mod_pos(ctx.k, ctx.k, ctx.a1);

        if (mpz_cmp_ui(ctx.sp, 1) != 0)
        {
            mpz_gcdext(ctx.s, ctx.v2, ctx.u2, ctx.ss, ctx.sp);
            mpz_mul(ctx.k, ctx.k, ctx.u2);
            mpz_submul(ctx.k, ctx.v2, ctx.c2);

            if (mpz_cmp_ui(ctx.s, 1) != 0)
            {
                mpz_divexact(ctx.a1, ctx.a1, ctx.s);
                mpz_divexact(ctx.a2, ctx.a2, ctx.s);
                mpz_mul(ctx.c2, ctx.c2, ctx.s);
            }

            common::mpz_mod_pos(ctx.k, ctx.k, ctx.a1);
        }

        if (mpz_cmp(ctx.a1, Form::nucomp_bound.get_mpz_t()) < 0)
        {
            mpz_mul(ctx.t, ctx.a2, ctx.k);

            mpz_mul(ctx.outa, ctx.a2, ctx.a1);

            mpz_mul_2exp(ctx.outb, ctx.t, 1);
            mpz_add(ctx.outb, ctx.outb, h->b);

            mpz_add(ctx.outc, h->b, ctx.t);
            mpz_mul(ctx.outc, ctx.outc, ctx.k);
            mpz_add(ctx.outc, ctx.outc, ctx.c2);
            mpz_fdiv_q(ctx.outc, ctx.outc, ctx.a1);
        }
        else
        {
            mpz_set(ctx.r2, ctx.a1);
            mpz_set(ctx.r1, ctx.k);
            fast_xgcd_partial(
                ctx.co2,
                ctx.co1,
                ctx.r2,
                ctx.r1,
                Form::nucomp_bound.get_mpz_t());

            mpz_mul(ctx.t, ctx.a2, ctx.r1);

            mpz_mul(ctx.m1, ctx.m, ctx.co1);
            mpz_add(ctx.m1, ctx.m1, ctx.t);
            mpz_divexact(ctx.m1, ctx.m1, ctx.a1);

            mpz_mul(ctx.m2, ctx.ss, ctx.r1);
            mpz_submul(ctx.m2, ctx.c2, ctx.co1);
            mpz_fdiv_q(ctx.m2, ctx.m2, ctx.a1);

            mpz_mul(ctx.outa, ctx.r1, ctx.m1);
            mpz_mul(ctx.temp, ctx.co1, ctx.m2);
            if (mpz_sgn(ctx.co1) < 0)
            {
                mpz_sub(ctx.outa, ctx.outa, ctx.temp);
            }
            else
            {
                mpz_sub(ctx.outa, ctx.temp, ctx.outa);
            }

            mpz_mul(ctx.temp, ctx.outa, ctx.co2);
            mpz_sub(ctx.outb, ctx.t, ctx.temp);
            mpz_mul_ui(ctx.outb, ctx.outb, 2);
            mpz_fdiv_q(ctx.outb, ctx.outb, ctx.co1);
            mpz_sub(ctx.outb, ctx.outb, h->b);
            mpz_mul_ui(ctx.temp, ctx.outa, 2);
            common::mpz_mod_pos(ctx.outb, ctx.outb, ctx.temp);

            //  outc = (outb * outb - Form::D) / (bigint::BigInt(uint64_t(4)) * outa);
            mpz_mul(ctx.b2, ctx.outb, ctx.outb);
            mpz_sub(ctx.b2subd, ctx.b2, Form::D.get_mpz_t());
            mpz_mul_ui(ctx.a4, ctx.outa, 4);
            mpz_fdiv_q(ctx.outc, ctx.b2subd, ctx.a4);

            if (mpz_sgn(ctx.outa) < 0)
            {
                mpz_neg(ctx.outa, ctx.outa);
                mpz_neg(ctx.outc, ctx.outc);
            }
        }

        mpz_set(this->a, ctx.outa);
        mpz_set(this->b, ctx.outb);
        mpz_set(this->c, ctx.outc);

        if (reduce)
            this->reduce_inplace();
    }

    void Form::fast_pow_inplace(mpz_srcptr exp)
    {
        if (mpz_sgn(exp) <= 0)
        {
            *this = Form::principal();
            return;
        }

        Form t = *this;
        this->reduce_inplace();

        const size_t bits = common::mpz_bit_length(exp);
        const size_t max_a_limbs = std::max<size_t>(2, mpz_size(Form::D.get_mpz_t()) / 2);

        for (size_t i = bits - 1; i > 0; --i)
        {
            this->nudupl_inplace();
            if (mpz_size(this->a) > max_a_limbs)
                this->reduce_inplace();

            if (mpz_tstbit(exp, i - 1) != 0)
                this->nucomp_inplace(t);
        }
        this->reduce_inplace();
    }

    Form Form::pow(mpz_srcptr exp) const
    {
        struct context
        {
            mpz_t e, one, tmp;
            context()
            {
                mpz_inits(e, one, tmp, nullptr);
                mpz_set_ui(one, 1);
            }
            ~context() { mpz_clears(e, one, tmp, nullptr); }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };
        static thread_local context ctx;

        Form acc = Form::principal();
        Form cur = *this;
        mpz_set(ctx.e, exp);

        while (mpz_sgn(ctx.e) > 0)
        {
            mpz_and(ctx.tmp, ctx.e, ctx.one);
            if (mpz_cmp_ui(ctx.tmp, 1) == 0)
            {
                acc.compose_inplace(cur);
            }
            mpz_fdiv_q_2exp(ctx.e, ctx.e, 1);

            if (mpz_sgn(ctx.e) > 0)
            {
                cur.compose_inplace(cur);
            }
        }

        acc.reduce_inplace();
        return acc;
    }

    std::tuple<Form, Form> Form::partial_pow(mpz_srcptr exp, size_t bits, Form prev_base) const
    {
        struct context
        {
            mpz_t e, one, tmp;
            context()
            {
                mpz_inits(e, one, tmp, nullptr);
                mpz_set_ui(one, 1);
            }
            ~context() { mpz_clears(e, one, tmp, nullptr); }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };
        static thread_local context ctx;

        Form acc = *this;
        Form base = prev_base;
        mpz_set(ctx.e, exp);

        for (size_t i = 0; i < bits; i++)
        {
            mpz_and(ctx.tmp, ctx.e, ctx.one);
            if (mpz_cmp_ui(ctx.tmp, 1) == 0)
            {
                acc.compose_inplace(base);
            }
            mpz_fdiv_q_2exp(ctx.e, ctx.e, 1);

            base.compose_inplace(base);
        }

        return {acc, base};
    }

    static inline mp_limb_signed_t bitlen_nonneg(const mpz_t x)
    {
        const size_t n = mpz_size(x);
        if (n == 0)
        {
            return 1;
        }
        const mp_limb_t top = mpz_getlimbn(x, static_cast<mp_size_t>(n - 1));
        if (top == 0)
        {
            return 1;
        }

        const int lead = __builtin_clzll(static_cast<unsigned long long>(top));

        return static_cast<mp_limb_signed_t>((n - 1) * static_cast<size_t>(GMP_LIMB_BITS)) +
               static_cast<mp_limb_signed_t>(GMP_LIMB_BITS - lead);
    }

    static inline mp_limb_signed_t extract_uword_from_shift_nonneg(const mpz_t x, mp_limb_signed_t shift_bits)
    {
        if (shift_bits <= 0)
            return static_cast<mp_limb_signed_t>(mpz_getlimbn(x, 0));

        const mp_limb_signed_t limb_bits = static_cast<mp_limb_signed_t>(GMP_LIMB_BITS);
        const mp_limb_signed_t limb_idx = shift_bits / limb_bits;
        const mp_limb_signed_t off = shift_bits - limb_idx * limb_bits;
        mp_limb_t lo = mpz_getlimbn(x, static_cast<mp_size_t>(limb_idx));
        if (off == 0)
            return static_cast<mp_limb_signed_t>(lo);

        mp_limb_t hi = mpz_getlimbn(x, static_cast<mp_size_t>(limb_idx + 1));
        lo >>= static_cast<unsigned>(off);
        hi <<= static_cast<unsigned>(limb_bits - off);
        return static_cast<mp_limb_signed_t>(lo | hi);
    }

    void Form::nudupl_inplace()
    {
        struct context
        {
            mpz_t a1, c1, cb, k, s, t, u2, v2;
            mpz_t b_abs;
            mpz_t m2, r1, r2, co1, co2, temp; // only used in the "a1 >= L" branch

            context() { mpz_inits(a1, c1, cb, k, s, t, u2, v2, b_abs, m2, r1, r2, co1, co2, temp, nullptr); }
            ~context() { mpz_clears(a1, c1, cb, k, s, t, u2, v2, b_abs, m2, r1, r2, co1, co2, temp, nullptr); }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };
        static thread_local context ctx;

        mpz_set(ctx.a1, this->a);
        mpz_set(ctx.c1, this->c);

        if (mpz_sgn(this->b) < 0)
        {
            mpz_neg(ctx.b_abs, this->b);
            mpz_gcdext(ctx.s, ctx.v2, NULL, ctx.b_abs, ctx.a1);
            mpz_neg(ctx.v2, ctx.v2);
        }
        else
        {
            mpz_set(ctx.b_abs, this->b);
            mpz_gcdext(ctx.s, ctx.v2, NULL, ctx.b_abs, ctx.a1);
        }

        mpz_mul(ctx.k, ctx.v2, ctx.c1);
        mpz_neg(ctx.k, ctx.k);

        if (mpz_cmp_ui(ctx.s, 1) != 0)
        {
            mpz_fdiv_q(ctx.a1, ctx.a1, ctx.s);
            mpz_mul(ctx.c1, ctx.c1, ctx.s);
        }

        mpz_fdiv_r(ctx.k, ctx.k, ctx.a1);
        if (mpz_sgn(ctx.k) < 0)
        {
            mpz_add(ctx.k, ctx.k, ctx.a1);
        }

        if (mpz_cmp(ctx.a1, classgroup::Form::nucomp_bound.get_mpz_t()) < 0)
        {
            mpz_mul(ctx.t, ctx.a1, ctx.k);
            mpz_mul(this->a, ctx.a1, ctx.a1);
            mpz_mul_2exp(ctx.cb, ctx.t, 1);
            mpz_add(ctx.cb, ctx.cb, this->b);
            mpz_add(this->c, this->b, ctx.t);
            mpz_mul(this->c, this->c, ctx.k);
            mpz_add(this->c, this->c, ctx.c1);
            mpz_fdiv_q(this->c, this->c, ctx.a1);
        }
        else
        {
            mpz_set(ctx.r2, ctx.a1);
            mpz_swap(ctx.r1, ctx.k);

            fast_xgcd_partial(
                ctx.co2,
                ctx.co1,
                ctx.r2,
                ctx.r1,
                classgroup::Form::nucomp_bound.get_mpz_t());

            mpz_mul(ctx.m2, this->b, ctx.r1);
            mpz_submul(ctx.m2, ctx.c1, ctx.co1);
            mpz_divexact(ctx.m2, ctx.m2, ctx.a1);

            mpz_mul(this->a, ctx.r1, ctx.r1);
            mpz_submul(this->a, ctx.co1, ctx.m2);

            if (mpz_sgn(ctx.co1) >= 0)
                mpz_neg(this->a, this->a);

            mpz_mul(ctx.cb, this->a, ctx.co2);
            mpz_submul(ctx.cb, ctx.a1, ctx.r1);
            mpz_neg(ctx.cb, ctx.cb);
            mpz_mul_2exp(ctx.cb, ctx.cb, 1);
            mpz_divexact(ctx.cb, ctx.cb, ctx.co1);
            mpz_sub(ctx.cb, ctx.cb, this->b);
            mpz_mul_2exp(ctx.temp, this->a, 1);
            mpz_fdiv_r(ctx.cb, ctx.cb, ctx.temp);

            mpz_mul(this->c, ctx.cb, ctx.cb);
            mpz_sub(this->c, this->c, classgroup::Form::D.get_mpz_t());
            mpz_divexact(this->c, this->c, this->a);
            mpz_fdiv_q_2exp(this->c, this->c, 2);

            if (mpz_sgn(this->a) < 0)
            {
                mpz_neg(this->a, this->a);
                mpz_neg(this->c, this->c);
            }
        }

        mpz_set(this->b, ctx.cb);
        this->reduce_inplace();

        // auto e = elapsed_ns(t0, Clock::now());
        // std::cout << "EVAL=" << e << "\n";
    }

    static void gcdinv(mpz_ptr g, mpz_ptr inv, mpz_srcptr x, mpz_srcptr m)
    {
        struct context
        {
            mpz_t s, t, d;
            context() { mpz_inits(s, t, d, nullptr); }
            ~context() { mpz_clears(s, t, d, nullptr); }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };

        static thread_local context ctx;

        mpz_gcdext(g, ctx.s, ctx.t, x, m);

        if (mpz_sgn(g) == 0)
        {
            throw std::runtime_error("gcdinv: zero gcd");
        }

        mpz_fdiv_q(ctx.d, m, g);
        common::mpz_mod_pos(inv, ctx.s, ctx.d);
    }

    void fast_xgcd_partial(mpz_t co2, mpz_t co1, mpz_t r2, mpz_t r1, const mpz_t l)
    {
        struct context
        {
            mpz_t q, r;
            context() { mpz_inits(q, r, nullptr); }
            ~context() { mpz_clears(q, r, nullptr); }
            context(const context &) = delete;
            context &operator=(const context &) = delete;
        };
        static thread_local context ctx;

        mp_limb_signed_t aa2, aa1, bb2, bb1, rr1, rr2, qq, bb, t1, t2, t3, i;
        mp_limb_signed_t bits, bits1, bits2;

        mpz_set_ui(co2, 0);
        mpz_set_si(co1, -1);

        while (mpz_cmp_ui(r1, 0) != 0 && mpz_cmp(r1, l) > 0)
        {
            bits2 = bitlen_nonneg(r2);
            bits1 = bitlen_nonneg(r1);
            bits = __GMP_MAX(bits2, bits1) - GMP_LIMB_BITS + 1;
            if (bits < 0)
                bits = 0;

            rr2 = extract_uword_from_shift_nonneg(r2, bits);
            rr1 = extract_uword_from_shift_nonneg(r1, bits);
            bb = extract_uword_from_shift_nonneg(l, bits);

            aa2 = 0;
            aa1 = 1;
            bb2 = 1;
            bb1 = 0;

            for (i = 0; rr1 != 0 && rr1 > bb; i++)
            {
                qq = rr2 / rr1;

                t1 = rr2 - qq * rr1;
                t2 = aa2 - qq * aa1;
                t3 = bb2 - qq * bb1;

                if (i & 1)
                {
                    if (t1 < -t3 || rr1 - t1 < t2 - aa1)
                        break;
                }
                else
                {
                    if (t1 < -t2 || rr1 - t1 < t3 - bb1)
                        break;
                }

                rr2 = rr1;
                rr1 = t1;
                aa2 = aa1;
                aa1 = t2;
                bb2 = bb1;
                bb1 = t3;
            }

            if (i == 0)
            {
                mpz_fdiv_qr(ctx.q, r2, r2, r1);
                mpz_swap(r2, r1);

                mpz_submul(co2, co1, ctx.q);
                mpz_swap(co2, co1);
            }
            else
            {
                mpz_mul_si(ctx.r, r2, bb2);
                if (aa2 >= 0)
                    mpz_addmul_ui(ctx.r, r1, aa2);
                else
                    mpz_submul_ui(ctx.r, r1, -aa2);
                mpz_mul_si(r1, r1, aa1);
                if (bb1 >= 0)
                    mpz_addmul_ui(r1, r2, bb1);
                else
                    mpz_submul_ui(r1, r2, -bb1);
                mpz_set(r2, ctx.r);

                mpz_mul_si(ctx.r, co2, bb2);
                if (aa2 >= 0)
                    mpz_addmul_ui(ctx.r, co1, aa2);
                else
                    mpz_submul_ui(ctx.r, co1, -aa2);
                mpz_mul_si(co1, co1, aa1);
                if (bb1 >= 0)
                    mpz_addmul_ui(co1, co2, bb1);
                else
                    mpz_submul_ui(co1, co2, -bb1);
                mpz_set(co2, ctx.r);

                if (mpz_sgn(r1) < 0)
                {
                    mpz_neg(co1, co1);
                    mpz_neg(r1, r1);
                }
                if (mpz_sgn(r2) < 0)
                {
                    mpz_neg(co2, co2);
                    mpz_neg(r2, r2);
                }
            }
        }

        if (mpz_sgn(r2) < 0)
        {
            mpz_neg(co2, co2);
            mpz_neg(co1, co1);
            mpz_neg(r2, r2);
        }
    }
} // namespace classgroup
