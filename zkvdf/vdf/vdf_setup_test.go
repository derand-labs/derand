package vdf_test

import (
	"zkvdf/nonzk"
	"zkvdf/utils"
	"zkvdf/vdf"
)

const (
	targetbits = 1024
	limbbits   = 64
)

var setup *vdf.Setup

func init() {
	D := utils.BigIntFromString("-e9c136470efb66c903f740ecbc5bf3ea9f0350f19da32fee52b3d8ce05431a9904dd45f6789695d8e2e1c594fe7a40b155e1fffeea5df2f6c25093b6f6d3765bd3d0b20cf82ec652f51f21a08a3892fc62d46ef830bf521e72b21e4bdb946286c1ae1e59e12d642db66794a2b9e9bf54e9c7fcb2de14b2213520277df6746e77", 16)
	hashToFormSteps := 12
	hashToFormGenerators := make([]nonzk.Form, 4)
	hashToFormGenerators[0] = nonzk.Form{
		A: utils.BigIntFromString("5", 16),
		B: utils.BigIntFromString("3", 16),
		C: utils.BigIntFromString("bb00f838d8c91f07365f6723c9e329887f35da5ae4828cbea88fe0a4d1027bad9d7dd191fa12117a4f1b047731fb66f444b4cccbeeb18f2bcea6dc925f0f91e30fda280a60256b7590e5b4806e93a8c9e8a9f26026ff74e5288e7ea316104ed2348b4b7b1a8ab68af852dd4efb2165dd87d3308f18108e80f74cec64c529f20", 16),
	}
	hashToFormGenerators[1] = nonzk.Form{
		A: utils.BigIntFromString("b", 16),
		B: utils.BigIntFromString("3", 16),
		C: utils.BigIntFromString("550070d40572b1034742a33ed0216ff839d2a9120ac6fa281e12da797646f266305076882bd9aada810c47d916b817864dc68ba283c5128846a8efe5710713c44d06123342f9bc7b41dcc668ec71a9d023f0285a404592399e12398ff2c19831009c6820aef93bb38825aa69b7f7e87bf7eba1b568078697b63a3ce7fc876e0", 16),
	}
	hashToFormGenerators[2] = nonzk.Form{
		A: utils.BigIntFromString("13", 16),
		B: utils.BigIntFromString("7", 16),
		C: utils.BigIntFromString("3136265fcd42667b294186ee787f25e08d44110a7207686811693b1de6291312bda7d8d592a670638e149562bc4fa1d47ddebca174b57684ca97b34ef0984ecff6b2ac38a009d8e90b2ef99b0fa01ef1c3f6d3fe5b1acdeb76765738641f3d29d7eec30572c630099faa04582723bc7daa7af1d4d070258041143e357740eee", 16),
	}
	hashToFormGenerators[3] = nonzk.Form{
		A: utils.BigIntFromString("1f", 16),
		B: utils.BigIntFromString("15", 16),
		C: utils.BigIntFromString("1e29725bbfde5fd7df7abe0e07c9cce4779d54c45667a3167e48c122d7637f4d8d03c6f6832c345e0cc24b0af78ba5402c1d294a2ec1ccc5004c76281fd9388ffa4c7a12308a2a1b38671d1cf90f8ed63e4cf58b6120f1d261614e3b5623a99dc6690c2ca12f25b35157b014ff36ef65cb95acfe4e34380448eb5fef3893274", 16),
	}

	setup = vdf.NewSetup(limbbits, D, targetbits, 128, 1, hashToFormSteps, hashToFormGenerators)
}
