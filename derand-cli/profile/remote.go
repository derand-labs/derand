package profile

import (
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

type RemoteProfile struct {
	Verifier     ethcommon.Address `json:"verifier"`
	BaseTime     time.Duration     `json:"base_time"`
	DelayTime    time.Duration     `json:"delay_time"`
	DelayScale   int               `json:"delay_scale"`
	MaximumDelay int               `json:"maximum_delay"`
}
