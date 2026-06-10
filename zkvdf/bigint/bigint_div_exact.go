package bigint

func (api bigIntAPI) DivExact(x BigInt, y BigInt) BigInt {
	divExactMag := api.BigUintAPI().DivExact(x.Mag, y.Mag)
	return BigInt{Sign: api.core.Mul(x.Sign, y.Sign), Mag: divExactMag}
}

func (api bigIntAPI) Div2Exact(x BigInt) BigInt {
	divExactMag := api.BigUintAPI().Div2Exact(x.Mag)
	return BigInt{Sign: x.Sign, Mag: divExactMag}
}
