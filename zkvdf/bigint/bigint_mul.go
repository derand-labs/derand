package bigint

func (api bigIntAPI) Mul(x BigInt, y BigInt) BigInt {
	mulMag := api.BigUintAPI().Mul(x.Mag, y.Mag)
	return BigInt{Sign: api.core.Mul(x.Sign, y.Sign), Mag: mulMag}
}
