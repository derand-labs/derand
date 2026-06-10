package bigint

func (api bigIntAPI) Sub(x, y BigInt) BigInt {
	return api.Add(x, api.Neg(y))
}

func (api bigIntAPI) SubOne(x BigInt) BigInt {
	biguintapi := api.BigUintAPI()

	xMagSub1, borrow := biguintapi.SubOneWithBorrow(x.Mag)
	xMagPlus1 := biguintapi.AddOne(x.Mag)

	borrowOrNegative := api.core.Or(borrow, api.IsNegative(x))
	mag := biguintapi.Select(borrowOrNegative, xMagPlus1, xMagSub1)
	magIsZero := biguintapi.IsZero(mag)

	sign := api.core.Select(api.core.IsZero(x.Sign), -1, x.Sign)
	sign = api.core.Select(magIsZero, 0, sign)

	return BigInt{Sign: sign, Mag: mag}
}
