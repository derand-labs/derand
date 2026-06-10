package classgroup

func (api classgroupAPI) AssertValid(fs ...Form) {
	for _, f := range fs {
		b2 := api.BigIntAPI().Mul(f.B, f.B)
		ac := api.BigIntAPI().Mul(f.A, f.C)
		ac2 := api.BigIntAPI().Add(ac, ac)
		ac4 := api.BigIntAPI().Add(ac2, ac2)
		d := api.BigIntAPI().Sub(b2, ac4)
		api.BigIntAPI().AssertIsEqual(d, api.D)

		api.BigIntAPI().AssertRangeCheck(f.A, f.B, f.C)
	}
}
