package bigint

func (api bigIntAPI) AssertRangeCheck(ins ...BigInt) {
	for _, in := range ins {
		api.core.AssertIsEqual(api.core.Mul(api.core.Add(in.Sign, 1), api.core.Mul(in.Sign, api.core.Sub(in.Sign, 1))), 0)
		api.BigUintAPI().AssertRangeCheck(in.Mag)
		api.core.AssertIsEqual(api.core.IsZero(in.Sign), api.BigUintAPI().IsZero(in.Mag))
	}
}
