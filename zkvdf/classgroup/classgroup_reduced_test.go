package classgroup_test

import (
	"testing"
	"zkvdf/bigint"
	"zkvdf/classgroup"
	"zkvdf/commontest"
	"zkvdf/utils"

	"github.com/consensys/gnark/frontend"
)

type ReduceTestCircuit struct {
	F   classgroup.Form `gnark:",public"`
	OUT classgroup.Form `gnark:",public"`
}

func (c *ReduceTestCircuit) Define(api frontend.API) error {
	cgapi := classgroup.NewAPI(api, setup)

	cgapi.AssertValid(c.F, c.OUT)

	OUT := cgapi.Reduce(c.F)
	cgapi.AssertIsEqual(OUT, c.OUT)
	return nil
}

func TestClassgroupReduceCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[classgroup.Form]{
			A: classgroup.Form{
				A: bigint.New(setup.DNumLimbs),
				B: bigint.New(setup.DNumLimbs + 1),
				C: bigint.New(setup.DNumLimbs),
			},
			F: func(api frontend.API, a classgroup.Form) {
				classgroup.NewAPI(api, setup).Reduce(a)
			},
		},
		125857,
	)
}

func TestClassgroupAssertIsReducedCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[classgroup.Form]{
			A: classgroup.Form{
				A: bigint.New(setup.DNumLimbs),
				B: bigint.New(setup.DNumLimbs),
				C: bigint.New(setup.DNumLimbs),
			},
			F: func(api frontend.API, a classgroup.Form) {
				classgroup.NewAPI(api, setup).AssertIsReduced(a)
			},
		},
		4654,
	)
}

func TestClassgroupReduceCorrect(t *testing.T) {
	a1 := utils.BigIntFromString("7ddb80df4ef8d8b92bdcc6e0cf409ddf5a3ee5a28af572214543e91d246d4890e4b3128abf9cb458a44707ae577fbab764a8fdb2d40c7b3a0ebea4b9981122cf08867d5d8a5010d6b45256e0e292125bdd605ec1e7bb6345f4e0856cae85460f024941a0ac02a385d1a6dc8a850cab4207559c501b00b4b9802a8afa10019ce", 16)
	b1 := utils.BigIntFromString("ba77d64312d95bdd03338ff5abe1c103ca0b2459ba33d22fa66628956f48d864e054950de32459e04d7d77af29a162eae432eb9f808014798c341c1de31c8f29cbf070c46da4a581b5c2d56517b35e835ceb099bfe7cdb4cc0c2380147c1460e62a669b4baced87beb444675a53f5b657dd7c4e26119b7ad29f35980529453", 16)
	c1 := utils.BigIntFromString("451123b2a8a58e9acfcbc7167b02fc22ead83964c43447e57e33726f80e7ea22899847f4d5ffc86b58fe7e3fec667777320ae22ffb2c39cbff269dd91f6e61b0c91fa2c680a3fbf06aa55fbd213a6e35f900339b3a3e9a59d2ac231d1dc2472842993408236cf21e80bfa0a28cded402135d94b42c15db894b6046b4f9cf4", 16)

	aout := utils.BigIntFromString("39aaa2f52045486f351ca81ee32f996d730d243171c17f98730fc384e5f1b8b30d1ed0f53387013d14e73a3252c70053ee7df6ebff23e241eb838f2d06638646", 16)
	bout := utils.BigIntFromString("8d397767ba5957fd2fe496996af65dbc047a746a8064e0008d6bdb9224fbdf895756647fe568cf680914848577e357adf456ba49295b1f01be75beda49b2a11", 16)
	cout := utils.BigIntFromString("103c3f7c608285629581ed5d8be1f8222980be658fd13796dcde3e2c5b00c0f7f1010e110271128ef8d18753497de25bcd7349b1dd244dbed29037f9da9b80a71", 16)

	commontest.TestCircuitValid(
		t,
		&ReduceTestCircuit{
			F: classgroup.Form{
				A: setup.BigInt.From(a1, setup.DNumLimbs),
				B: setup.BigInt.From(b1, setup.DNumLimbs),
				C: setup.BigInt.From(c1, setup.DNumLimbs*2),
			},
			OUT: classgroup.Form{
				A: setup.BigInt.From(aout, setup.DNumLimbs/2),
				B: setup.BigInt.From(bout, setup.DNumLimbs/2),
				C: setup.BigInt.From(cout, setup.DNumLimbs),
			},
		},
	)
}
