package biguint

type Setup struct {
	LimbBits int
}

func NewSetup(limbbits int) *Setup {
	return &Setup{LimbBits: limbbits}
}
