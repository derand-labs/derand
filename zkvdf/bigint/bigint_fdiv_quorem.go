package bigint

import (
	"github.com/consensys/gnark/frontend"
)

func (api bigIntAPI) FloorDivQuoRem(x BigInt, y BigInt) (BigInt, BigInt) {
	biguintapi := api.BigUintAPI()

	sameSignOrZero := api.IsSameSignOrZero(x, y)

	qTMag, rTMag := biguintapi.DivQuoRem(x.Mag, y.Mag)
	rMagIsZero := biguintapi.IsZero(rTMag)

	qFMag := biguintapi.Select(rMagIsZero, qTMag, biguintapi.AddOne(qTMag))
	qMag := biguintapi.Select(sameSignOrZero, qTMag, qFMag)
	qMagIsZero := biguintapi.IsZero(qMag)
	qSign := api.core.Mul(x.Sign, y.Sign)
	qSign = api.core.Select(qMagIsZero, 0, qSign)
	q := BigInt{Sign: qSign, Mag: qMag}

	rT := BigInt{Sign: api.core.Select(rMagIsZero, 0, x.Sign), Mag: rTMag}
	rF := api.Add(rT, y)
	rF = api.Select(rMagIsZero, rT, rF)
	r := api.Select(sameSignOrZero, rT, rF)

	return api.Cast(q, len(x.Mag.Limbs)), api.Cast(r, len(y.Mag.Limbs))
}

func (api bigIntAPI) FloorDiv2QuoRem(x BigInt) (BigInt, frontend.Variable) {
	biguintapi := api.BigUintAPI()

	qTMag, rT := biguintapi.Div2QuoRem(x.Mag)
	rIsZero := api.core.IsZero(rT)

	diffSign := api.core.IsZero(api.core.Add(x.Sign, 1))

	qFMag := biguintapi.Select(rIsZero, qTMag, biguintapi.AddOne(qTMag))
	qMag := biguintapi.Select(diffSign, qFMag, qTMag)
	qMagIsZero := biguintapi.IsZero(qMag)
	q := BigInt{Sign: api.core.Select(qMagIsZero, 0, x.Sign), Mag: qMag}

	rT = api.core.Mul(x.Sign, rT)
	rF := api.core.Add(rT, api.core.Select(rIsZero, 0, 2))
	r := api.core.Select(diffSign, rF, rT)

	return api.Cast(q, len(x.Mag.Limbs)), r
}

func (api bigIntAPI) FloorDiv4QuoRem(x BigInt) (BigInt, frontend.Variable) {
	biguintapi := api.BigUintAPI()

	qTMag, rT := biguintapi.Div4QuoRem(x.Mag)
	rIsZero := api.core.IsZero(rT)

	diffSign := api.core.IsZero(api.core.Add(x.Sign, 1))

	qFMag := biguintapi.Select(rIsZero, qTMag, biguintapi.AddOne(qTMag))
	qMag := biguintapi.Select(diffSign, qFMag, qTMag)
	qMagIsZero := biguintapi.IsZero(qMag)
	q := BigInt{Sign: api.core.Select(qMagIsZero, 0, x.Sign), Mag: qMag}

	rT = api.core.Mul(x.Sign, rT)
	rF := api.core.Add(rT, api.core.Select(rIsZero, 0, 4))
	r := api.core.Select(diffSign, rF, rT)

	return api.Cast(q, len(x.Mag.Limbs)), r
}
