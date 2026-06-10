package prover

import (
	"derand-cli/proverlogic"
	"derand-cli/utils"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	flagProveSubmitPreview bool
)

var proveCmd = &cobra.Command{
	Use:   "prove",
	Short: "prove a single request",

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("invalid argument: derand prover prove [request-id]")
		}

		requestID, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid request id: %w", err)
		}

		prover, err := proverlogic.NewProver(requestID, flagProveSubmitPreview, utils.AskPassword("Enter password: "))
		if err != nil {
			return err
		}

		if err := prover.Prove(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	proveCmd.Flags().BoolVar(&flagProveSubmitPreview, "submit-preview", false, "allow submit preview (only for owner of request)")
}
