package request

import (
	"derand-cli/config"
	"derand-cli/gen"
	"derand-cli/utils"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "get info of a request",

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("invalid argument: derand prover prove [request-id]")
		}

		requestID, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid request id: %w", err)
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		currentChain, err := cfg.GetCurrentChain()
		if err != nil {
			return err
		}

		if currentChain.DeRand == nil {
			return fmt.Errorf("derand has not been deployed yet, please run `derand deploy` or `derand chain setup` first!")
		}

		backend, err := cfg.GetCurrentChainBackend()
		if err != nil {
			return err
		}

		derand, err := gen.NewDeRand(currentChain.DeRand.Address, backend)
		if err != nil {
			return fmt.Errorf("failed to initialize derand: %w", err)
		}

		request, err := derand.RequestOf(nil, requestID)
		if err != nil {
			return err
		}

		utils.PrintTitle(utils.Bold("Request Info"))
		utils.PrintSubtitle("profile:", fmt.Sprintf("%d[%d]", request.ProfileId, request.ProfileVersion))
		switch request.Status {
		case 0:
			utils.PrintSubtitle("status: Pending")
		case 1:
			utils.PrintSubtitle("status: Assigned")
		case 2:
			utils.PrintSubtitle("status: Fulfilled")
		case 3:
			utils.PrintSubtitle("status: Open")
		}
		utils.PrintSubtitle("delay:", request.Delay)
		utils.PrintSubtitle("seed:", hexutil.Bytes(request.Seed.Bytes()))
		if request.Status == 2 {
			utils.PrintSubtitle("random number:", hexutil.Bytes(request.RandomNumber.Bytes()))
		} else {
			utils.PrintSubtitle("random number: waiting")
		}

		return nil
	},
}
