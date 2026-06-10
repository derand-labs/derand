package commontest

import (
	"math/big"
	"zkvdf/utils"
)

func BigInt1() *big.Int {
	return utils.BigIntFromString("380148d8063e1c1a87b5e5738e6baaafda9b2e7a60a153de18a38c987edc87e2280c9ffe813339972eeee1ff97cc3fccaab6711116dea72be95d5f766ec9e0ae5ec49b5d321ebe1cfaae4838c0f43811a92e2", 16)

}

func BigInt2() *big.Int {
	return utils.BigIntFromString("22275e2ea1bb93e419ae061c77ad70fa35daddca0a3139e4455a329834b34a5fcd707824b382fc0a3afe827e9197f86c65305d863db19fb3036bf00c5e7d3b2f0d6b61f1b7a37135baaa1e38100063366e103", 16)
}
