package classgroup

import (
	"math/big"
)

func (api classgroupAPI) Compose(f1, f2 Form) Form {
	bsum := api.BigIntAPI().Add(f1.B, f2.B)
	halfbsum := api.BigIntAPI().Div2Exact(bsum)
	g1, _, _, _, _ := api.BigIntAPI().GCD(f1.A, f2.A)
	g1, _, k, _, _ := api.BigIntAPI().GCD(g1, halfbsum)

	a3 := api.BigIntAPI().Mul(f1.A, f2.A)
	g1sqr := api.BigIntAPI().Mul(g1, g1)
	a3 = api.BigIntAPI().DivExact(a3, g1sqr)

	m1 := api.BigIntAPI().DivExact(f1.A, g1)
	m1 = api.BigIntAPI().Add(m1, m1)

	m2 := api.BigIntAPI().DivExact(f2.A, g1)
	m2 = api.BigIntAPI().Add(m2, m2)

	g, m1p, m2p, _, _ := api.BigIntAPI().GCD(m1, m2)

	diff := api.BigIntAPI().Sub(f2.B, f1.B)
	diff = api.BigIntAPI().DivExact(diff, g)

	inv := api.BigIntAPI().InvMod(m1p, m2p)

	tcrt := api.BigIntAPI().Mul(diff, inv)
	_, tcrt = api.BigIntAPI().FloorDivQuoRem(tcrt, m2p)

	x0 := api.BigIntAPI().Mul(m1, tcrt)
	x0 = api.BigIntAPI().Add(x0, f1.B)

	lcm12 := api.BigIntAPI().Mul(m1, m2p)
	_, x0 = api.BigIntAPI().FloorDivQuoRem(x0, lcm12)

	M := api.BigIntAPI().Add(a3, a3)

	rhs := api.BigIntAPI().Mul(f1.B, f2.B)
	rhs = api.BigIntAPI().Add(rhs, api.D)

	den := api.BigIntAPI().Add(g1, g1)
	rhs = api.BigIntAPI().DivExact(rhs, den)

	tmp := api.BigIntAPI().Mul(k, x0)
	tmp = api.BigIntAPI().Sub(rhs, tmp)

	s := api.BigIntAPI().DivExact(tmp, lcm12)
	q := api.BigIntAPI().DivExact(M, lcm12)

	g, k, q, _, _ = api.BigIntAPI().GCD(k, q)
	s = api.BigIntAPI().DivExact(s, g)

	t1 := api.setup.BigInt.From(big.NewInt(0), 1)

	inv = api.BigIntAPI().InvMod(k, q)
	t2 := api.BigIntAPI().Mul(s, inv)
	_, t2 = api.BigIntAPI().FloorDivQuoRem(t2, q)

	t := api.BigIntAPI().Select(api.BigIntAPI().IsEqualUI(q, 1), t1, t2)

	b3 := api.BigIntAPI().Mul(lcm12, t)
	b3 = api.BigIntAPI().Add(b3, x0)
	_, b3 = api.BigIntAPI().FloorDivQuoRem(b3, M)

	c3 := api.BigIntAPI().Mul(b3, b3)
	c3 = api.BigIntAPI().Sub(c3, api.D)
	tmp = api.BigIntAPI().Add(a3, a3)
	tmp = api.BigIntAPI().Add(tmp, tmp)
	c3 = api.BigIntAPI().DivExact(c3, tmp)

	return Form{
		A: api.BigIntAPI().Cast(a3, api.setup.DNumLimbs),
		B: api.BigIntAPI().Cast(b3, api.setup.DNumLimbs+1),
		C: api.BigIntAPI().Cast(c3, api.setup.DNumLimbs),
	}
}
