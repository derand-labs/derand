package bigint

func (api bigIntAPI) Abs(in BigInt) BigInt {
	var out BigInt

	isNegative := api.core.IsZero(api.core.Add(in.Sign, 1))
	out.Sign = api.core.Select(isNegative, api.core.Mul(in.Sign, -1), in.Sign)
	out.Mag = in.Mag
	return out
}

func (api bigIntAPI) assertAbs(out, in BigInt) {
	api.core.AssertIsDifferent(out.Sign, -1)
	api.BigUintAPI().AssertIsEqual(out.Mag, in.Mag)
}
