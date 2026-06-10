package classgroup_test

import (
	"zkvdf/classgroup"
	"zkvdf/utils"
)

const (
	targetbits = 1024
	limbbits   = 64
)

var setup = classgroup.NewSetup(
	limbbits,
	utils.BigIntFromString("-e9c136470efb66c903f740ecbc5bf3ea9f0350f19da32fee52b3d8ce05431a9904dd45f6789695d8e2e1c594fe7a40b155e1fffeea5df2f6c25093b6f6d3765bd3d0b20cf82ec652f51f21a08a3892fc62d46ef830bf521e72b21e4bdb946286c1ae1e59e12d642db66794a2b9e9bf54e9c7fcb2de14b2213520277df6746e77", 16),
	targetbits,
)
