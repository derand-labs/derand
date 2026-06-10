package classgroup

import (
	"fmt"
	"math/big"
	"zkvdf/bigint"

	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
)

func init() {
	solver.RegisterHint(hintReduce)
}

func (api classgroupAPI) Reduce(f Form) Form {
	hintInputs := []frontend.Variable{
		api.setup.BigInt.BigUint.LimbBits,
		len(api.D.Mag.Limbs),
		len(f.A.Mag.Limbs),
		len(f.B.Mag.Limbs),
	}
	hintInputs = append(hintInputs, api.D.Sign)
	hintInputs = append(hintInputs, api.D.Mag.Limbs...)
	hintInputs = append(hintInputs, f.A.Sign)
	hintInputs = append(hintInputs, f.A.Mag.Limbs...)
	hintInputs = append(hintInputs, f.B.Sign)
	hintInputs = append(hintInputs, f.B.Mag.Limbs...)
	hintInputs = append(hintInputs, f.C.Sign)
	hintInputs = append(hintInputs, f.C.Mag.Limbs...)

	aOutNumLimbs := (len(api.D.Mag.Limbs)-1)/2 + 1
	bOutNumLimbs := aOutNumLimbs
	cOutNumLimbs := len(api.D.Mag.Limbs)

	// outputs = A B C p q r s
	reduceHint, err := api.core.Compiler().NewHint(
		hintReduce,
		aOutNumLimbs+bOutNumLimbs+cOutNumLimbs+len(api.D.Mag.Limbs)*4+7,
		hintInputs...,
	)
	if err != nil {
		panic(err)
	}

	offset := 0
	A := bigint.FromHint(reduceHint[offset : offset+aOutNumLimbs+1])
	offset += aOutNumLimbs + 1

	B := bigint.FromHint(reduceHint[offset : offset+bOutNumLimbs+1])
	offset += bOutNumLimbs + 1

	C := bigint.FromHint(reduceHint[offset : offset+cOutNumLimbs+1])
	offset += cOutNumLimbs + 1

	p := bigint.FromHint(reduceHint[offset : offset+len(api.D.Mag.Limbs)+1])
	offset += len(api.D.Mag.Limbs) + 1

	q := bigint.FromHint(reduceHint[offset : offset+len(api.D.Mag.Limbs)+1])
	offset += len(api.D.Mag.Limbs) + 1

	r := bigint.FromHint(reduceHint[offset : offset+len(api.D.Mag.Limbs)+1])
	offset += len(api.D.Mag.Limbs) + 1

	s := bigint.FromHint(reduceHint[offset : offset+len(api.D.Mag.Limbs)+1])

	out := Form{A: A, B: B, C: C}

	api.AssertValid(out)
	api.AssertIsReduced(out)

	api.BigIntAPI().AssertRangeCheck(p, q, r, s)

	p2 := api.BigIntAPI().Mul(p, p)
	q2 := api.BigIntAPI().Mul(q, q)
	r2 := api.BigIntAPI().Mul(r, r)
	s2 := api.BigIntAPI().Mul(s, s)
	sr := api.BigIntAPI().Mul(s, r)
	sq := api.BigIntAPI().Mul(s, q)
	rp := api.BigIntAPI().Mul(r, p)
	pq := api.BigIntAPI().Mul(p, q)
	ps := api.BigIntAPI().Mul(p, s)
	qr := api.BigIntAPI().Mul(q, r)

	// det(M) = 1
	api.BigIntAPI().AssertIsEqual(ps, api.BigIntAPI().AddOne(qr))

	as2 := api.BigIntAPI().Mul(A, s2)
	bsr := api.BigIntAPI().Mul(B, sr)
	cr2 := api.BigIntAPI().Mul(C, r2)
	tmp := api.BigIntAPI().Sub(as2, bsr)
	tmp = api.BigIntAPI().Add(tmp, cr2)
	api.BigIntAPI().AssertIsEqual(f.A, tmp)

	asq := api.BigIntAPI().Mul(A, sq)
	asq2 := api.BigIntAPI().Add(asq, asq)
	asq2neg := api.BigIntAPI().Neg(asq2)
	tmp = api.BigIntAPI().Add(ps, qr)
	btmp := api.BigIntAPI().Mul(B, tmp)
	crp := api.BigIntAPI().Mul(C, rp)
	crp2 := api.BigIntAPI().Add(crp, crp)
	tmp = api.BigIntAPI().Add(asq2neg, btmp)
	tmp = api.BigIntAPI().Sub(tmp, crp2)
	api.BigIntAPI().AssertIsEqual(f.B, tmp)

	aq2 := api.BigIntAPI().Mul(A, q2)
	bpq := api.BigIntAPI().Mul(B, pq)
	cp2 := api.BigIntAPI().Mul(C, p2)
	tmp = api.BigIntAPI().Sub(aq2, bpq)
	tmp = api.BigIntAPI().Add(tmp, cp2)
	api.BigIntAPI().AssertIsEqual(f.C, tmp)

	return out
}

func (api classgroupAPI) AssertIsReduced(f Form) {
	api.BigIntAPI().AssertIsPositive(f.A)
	api.BigIntAPI().AssertIsPositive(f.C)

	a := f.A.Mag
	c := f.C.Mag
	bAbs := f.B.Mag

	aGtBAbs := api.BigIntAPI().BigUintAPI().IsGreater(a, bAbs)
	cGtA := api.BigIntAPI().BigUintAPI().IsGreater(c, a)

	aEqBAbs := api.BigIntAPI().BigUintAPI().IsEqual(a, bAbs)
	cEqA := api.BigIntAPI().BigUintAPI().IsEqual(c, a)

	api.core.AssertIsEqual(api.core.Or(aGtBAbs, aEqBAbs), 1)
	api.core.AssertIsEqual(api.core.Or(cGtA, cEqA), 1)

	eqCondition := api.core.Or(aEqBAbs, cEqA)
	bNonNegative := api.BigIntAPI().IsNonNegative(f.B)

	api.core.AssertIsEqual(api.core.Or(api.core.Sub(1, eqCondition), bNonNegative), 1)
}

func hintReduce(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	const prefixInputs = 4

	limbbits := int(inputs[0].Int64())
	DNumLimbs := int(inputs[1].Int64())
	aNumLimbs := int(inputs[2].Int64())
	bNumLimbs := int(inputs[3].Int64())

	offset := prefixInputs

	D := bigint.FromHintInput(inputs[offset:offset+DNumLimbs+1], limbbits)
	offset += DNumLimbs + 1

	a := bigint.FromHintInput(inputs[offset:offset+aNumLimbs+1], limbbits)
	offset += aNumLimbs + 1

	b := bigint.FromHintInput(inputs[offset:offset+bNumLimbs+1], limbbits)
	offset += bNumLimbs + 1

	c := bigint.FromHintInput(inputs[offset:], limbbits)

	A, B, C, p, q, r, s, err := reduceQFB(D, a, b, c)
	if err != nil {
		return err
	}

	aOutNumLimbs := (DNumLimbs-1)/2 + 1
	bOutNumLimbs := aOutNumLimbs
	cOutNumLimbs := DNumLimbs

	offset = 0

	bigint.ToHintOutput(outputs[offset:], A, aOutNumLimbs, limbbits)
	offset += aOutNumLimbs + 1

	bigint.ToHintOutput(outputs[offset:], B, bOutNumLimbs, limbbits)
	offset += bOutNumLimbs + 1

	bigint.ToHintOutput(outputs[offset:], C, cOutNumLimbs, limbbits)
	offset += cOutNumLimbs + 1

	bigint.ToHintOutput(outputs[offset:], p, DNumLimbs, limbbits)
	offset += DNumLimbs + 1

	bigint.ToHintOutput(outputs[offset:], q, DNumLimbs, limbbits)
	offset += DNumLimbs + 1

	bigint.ToHintOutput(outputs[offset:], r, DNumLimbs, limbbits)
	offset += DNumLimbs + 1

	bigint.ToHintOutput(outputs[offset:], s, DNumLimbs, limbbits)

	return nil
}

func reduceQFB(D, a, b, c *big.Int) (
	A, B, C *big.Int,
	p, q, r, s *big.Int,
	err error,
) {
	// floor(num/den), den>0
	floorDiv := func(num, den *big.Int) *big.Int {
		q := new(big.Int)
		rem := new(big.Int)

		q.QuoRem(num, den, rem)

		if num.Sign() < 0 && rem.Sign() != 0 {
			q.Sub(q, big.NewInt(1))
		}

		return q
	}

	// verify discriminant
	tmp := new(big.Int).Mul(a, c)
	tmp.Mul(tmp, big.NewInt(4))

	disc := new(big.Int).Mul(b, b)
	disc.Sub(disc, tmp)

	if disc.Cmp(D) != 0 {
		return nil, nil, nil, nil, nil, nil, nil,
			fmt.Errorf("invalid discriminant")
	}

	if a.Sign() <= 0 {
		return nil, nil, nil, nil, nil, nil, nil,
			fmt.Errorf("a must be positive")
	}

	A = new(big.Int).Set(a)
	B = new(big.Int).Set(b)
	C = new(big.Int).Set(c)

	// transcript matrix M = [[p,q],[r,s]]
	p = big.NewInt(1)
	q = big.NewInt(0)
	r = big.NewInt(0)
	s = big.NewInt(1)

	for {
		// normalize b into (-a, a]
		twoA := new(big.Int).Lsh(A, 1)

		num := new(big.Int).Sub(A, B)

		t := floorDiv(num, twoA)

		if t.Sign() != 0 {
			oldB := new(big.Int).Set(B)

			// B = B + 2At
			tmp = new(big.Int).Mul(twoA, t)
			B.Add(B, tmp)

			// C = At² + Bt + C
			t2 := new(big.Int).Mul(t, t)

			tmp = new(big.Int).Mul(A, t2)

			tmp2 := new(big.Int).Mul(oldB, t)

			C.Add(C, tmp)
			C.Add(C, tmp2)

			// transcript *= [[1,t],[0,1]]
			q.Add(q, new(big.Int).Mul(p, t))
			s.Add(s, new(big.Int).Mul(r, t))
		}

		// if a > c => rho step
		if A.Cmp(C) > 0 ||
			(A.Cmp(C) == 0 && B.Sign() < 0) {

			oldA := new(big.Int).Set(A)
			oldB := new(big.Int).Set(B)

			A.Set(C)
			B.Neg(oldB)
			C.Set(oldA)

			// transcript *= [[0,1],[-1,0]]
			oldP := new(big.Int).Set(p)
			oldQ := new(big.Int).Set(q)
			oldR := new(big.Int).Set(r)
			oldS := new(big.Int).Set(s)

			p.Neg(oldQ)
			q.Set(oldP)
			r.Neg(oldS)
			s.Set(oldR)

			continue
		}

		// reduced check
		absB := new(big.Int).Abs(B)

		ok1 := absB.Cmp(A) <= 0
		ok2 := A.Cmp(C) <= 0

		ok3 := true
		if absB.Cmp(A) == 0 || A.Cmp(C) == 0 {
			ok3 = B.Sign() >= 0
		}

		if ok1 && ok2 && ok3 {
			break
		}
	}

	return
}
