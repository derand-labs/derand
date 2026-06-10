package bigint

func (api bigIntAPI) Neg(in BigInt) BigInt {
	var out BigInt
	out.Sign = api.core.Mul(in.Sign, -1)
	out.Mag = in.Mag
	return out
}

func (api bigIntAPI) assertNeg(out, in BigInt) {
	api.BigUintAPI().AssertIsEqual(in.Mag, out.Mag)
	api.core.AssertIsEqual(api.core.Mul(in.Sign, -1), out.Sign)
}
