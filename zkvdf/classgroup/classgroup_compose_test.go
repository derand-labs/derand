package classgroup_test

import (
	"testing"
	"zkvdf/bigint"
	"zkvdf/classgroup"
	"zkvdf/commontest"
	"zkvdf/utils"

	"github.com/consensys/gnark/frontend"
)

type ComposeTestCircuit struct {
	F1  classgroup.Form `gnark:",public"`
	F2  classgroup.Form `gnark:",public"`
	OUT classgroup.Form `gnark:",public"`
}

func (c *ComposeTestCircuit) Define(api frontend.API) error {
	cgapi := classgroup.NewAPI(api, setup)

	cgapi.AssertValid(c.F1, c.F2, c.OUT)

	OUT := cgapi.Compose(c.F1, c.F2)
	cgapi.AssertIsEqual(OUT, c.OUT)
	return nil
}

func TestClassgroupComposeCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit2[classgroup.Form]{
			A: classgroup.Form{
				A: bigint.New(setup.GetSmallNumLimbs()),
				B: bigint.New(setup.GetSmallNumLimbs()),
				C: bigint.New(setup.DNumLimbs),
			},
			B: classgroup.Form{
				A: bigint.New(setup.GetSmallNumLimbs()),
				B: bigint.New(setup.GetSmallNumLimbs()),
				C: bigint.New(setup.DNumLimbs),
			},
			F: func(api frontend.API, a, b classgroup.Form) {
				classgroup.NewAPI(api, setup).Compose(a, b)
			},
		},
		182087,
	)
}

func TestClassgroupComposeCorrect(t *testing.T) {
	a1 := utils.BigIntFromString("521639cd8c0f6f2168f9e37be67945ee819685e86c35a027f35f275cf0a79e812b00eac2b958cb8126abbd40231832d5189e6edd9ef3f26bb49b46c59e464f77", 16)
	b1 := utils.BigIntFromString("-f2532f84467be827f3432ca62d2ab29de1644627af062ad092eab9edaa631ba04426d23da1e7dbb08016a676a1109fcb5e3c89b31662545e125e9bf84d8dadd", 16)
	c1 := utils.BigIntFromString("b6f2c6ffc9285623d27a70d98e0ae08d3ec639fb25b1767d5ff23a6518ec00f1e4cd6007650a014fcf76ce5c7b3985635ccd81191c18613d29e8b04e0605f230", 16)

	a2 := utils.BigIntFromString("1888154b554f99567e82377de7b6bb7ebea790e5bde4ace5405465a4a5186d8f5133d1b97bbc4c426ce62b653469a8778dfecfd6a44b93e49662121bc1bdd422", 16)
	b2 := utils.BigIntFromString("36ab218675d712e4c1eb0fd350d7246af4c9a29170ca6300ed16972e45507778e79bbab555f85d62b2ec7233b54f8b9136631a553a277086d59f562731190fb", 16)
	c2 := utils.BigIntFromString("261f5280aeccdda01b93a923c6880f9c568f0aecaf37739d7febc0341b29b670a78fc9c03fb0711e49bfec9a9499e79b9efd93eeb71e447d1910bc90ba65bad72", 16)

	a3 := utils.BigIntFromString("7ddb80df4ef8d8b92bdcc6e0cf409ddf5a3ee5a28af572214543e91d246d4890e4b3128abf9cb458a44707ae577fbab764a8fdb2d40c7b3a0ebea4b9981122cf08867d5d8a5010d6b45256e0e292125bdd605ec1e7bb6345f4e0856cae85460f024941a0ac02a385d1a6dc8a850cab4207559c501b00b4b9802a8afa10019ce", 16)
	b3 := utils.BigIntFromString("ba77d64312d95bdd03338ff5abe1c103ca0b2459ba33d22fa66628956f48d864e054950de32459e04d7d77af29a162eae432eb9f808014798c341c1de31c8f29cbf070c46da4a581b5c2d56517b35e835ceb099bfe7cdb4cc0c2380147c1460e62a669b4baced87beb444675a53f5b657dd7c4e26119b7ad29f35980529453", 16)
	c3 := utils.BigIntFromString("451123b2a8a58e9acfcbc7167b02fc22ead83964c43447e57e33726f80e7ea22899847f4d5ffc86b58fe7e3fec667777320ae22ffb2c39cbff269dd91f6e61b0c91fa2c680a3fbf06aa55fbd213a6e35f900339b3a3e9a59d2ac231d1dc2472842993408236cf21e80bfa0a28cded402135d94b42c15db894b6046b4f9cf4", 16)

	commontest.TestCircuitValid(
		t,
		&ComposeTestCircuit{
			F1: classgroup.Form{
				A: setup.BigInt.From(a1, setup.DNumLimbs/2),
				B: setup.BigInt.From(b1, setup.DNumLimbs/2),
				C: setup.BigInt.From(c1, setup.DNumLimbs),
			},
			F2: classgroup.Form{
				A: setup.BigInt.From(a2, setup.DNumLimbs/2),
				B: setup.BigInt.From(b2, setup.DNumLimbs/2),
				C: setup.BigInt.From(c2, setup.DNumLimbs),
			},
			OUT: classgroup.Form{
				A: setup.BigInt.From(a3, setup.DNumLimbs),
				B: setup.BigInt.From(b3, setup.DNumLimbs),
				C: setup.BigInt.From(c3, setup.DNumLimbs),
			},
		},
	)
}
