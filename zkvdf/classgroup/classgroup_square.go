package classgroup

import (
	"math/big"
)

func (api classgroupAPI) Square(f1 Form) Form {
	g1, m, k, _, _ := api.BigIntAPI().GCD(f1.A, f1.B)

	a3 := api.BigIntAPI().Mul(f1.A, f1.A)
	g1sqr := api.BigIntAPI().Mul(g1, g1)
	a3 = api.BigIntAPI().DivExact(a3, g1sqr)

	m = api.BigIntAPI().Add(m, m)

	_, x0 := api.BigIntAPI().FloorDivQuoRem(f1.B, m)

	M := api.BigIntAPI().Add(a3, a3)

	rhs := api.BigIntAPI().Mul(f1.B, f1.B)
	rhs = api.BigIntAPI().Add(rhs, api.D)

	den := api.BigIntAPI().Add(g1, g1)
	rhs = api.BigIntAPI().DivExact(rhs, den)

	tmp := api.BigIntAPI().Mul(k, x0)
	tmp = api.BigIntAPI().Sub(rhs, tmp)

	s := api.BigIntAPI().DivExact(tmp, m)
	q := api.BigIntAPI().DivExact(M, m)

	g, k, q, _, _ := api.BigIntAPI().GCD(k, q)
	s = api.BigIntAPI().DivExact(s, g)

	t1 := api.setup.BigInt.From(big.NewInt(0), 1)

	inv := api.BigIntAPI().InvMod(k, q)
	t2 := api.BigIntAPI().Mul(s, inv)
	_, t2 = api.BigIntAPI().FloorDivQuoRem(t2, q)

	t := api.BigIntAPI().Select(api.BigIntAPI().IsEqualUI(q, 1), t1, t2)

	b3 := api.BigIntAPI().Mul(m, t)
	b3 = api.BigIntAPI().Add(b3, x0)
	_, b3 = api.BigIntAPI().FloorDivQuoRem(b3, M)

	c3 := api.BigIntAPI().Mul(b3, b3)
	c3 = api.BigIntAPI().Sub(c3, api.D)
	tmp = api.BigIntAPI().Add(a3, a3)
	tmp = api.BigIntAPI().Add(tmp, tmp)
	c3 = api.BigIntAPI().DivExact(c3, tmp)

	return Form{A: a3, B: b3, C: c3}
}
