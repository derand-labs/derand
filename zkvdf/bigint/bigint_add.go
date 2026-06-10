package bigint

func (api bigIntAPI) Add(x, y BigInt) BigInt {
	sameSignOrZero := api.IsSameSignOrZero(x, y)

	eq := api.BigUintAPI().IsEqual(x.Mag, y.Mag)

	// For b-a: if borrow is 1 <==> a greater than b.
	sub1, gt := api.BigUintAPI().SubWithBorrow(y.Mag, x.Mag)
	sub2 := api.BigUintAPI().Neg(sub1)
	sub := api.BigUintAPI().Select(gt, sub2, sub1)
	add := api.BigUintAPI().Add(x.Mag, y.Mag)

	mag := api.BigUintAPI().Select(sameSignOrZero, add, sub)

	signForSum := api.core.Select(api.core.IsZero(x.Sign), y.Sign, x.Sign)
	signForSub := api.core.Select(eq, 0, y.Sign)
	signForSub = api.core.Select(gt, x.Sign, signForSub)
	sign := api.core.Select(sameSignOrZero, signForSum, signForSub)

	return BigInt{Sign: sign, Mag: mag}
}

func (api bigIntAPI) AddOne(x BigInt) BigInt {
	xMagPlus1 := api.BigUintAPI().AddOne(x.Mag)
	xMagSub1, borrow := api.BigUintAPI().SubOneWithBorrow(x.Mag)

	BorrowOrNonNegative := api.core.Or(borrow, api.IsNonNegative(x))
	mag := api.BigUintAPI().Select(BorrowOrNonNegative, xMagPlus1, xMagSub1)
	magIsZero := api.BigUintAPI().IsZero(mag)

	sign := api.core.Select(api.core.IsZero(x.Sign), 1, x.Sign)
	sign = api.core.Select(magIsZero, 0, sign)

	return BigInt{Sign: sign, Mag: mag}
}
