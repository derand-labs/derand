package chain

import (
	"derand-cli/config"
	"derand-cli/utils"
	"slices"

	"github.com/spf13/cobra"
)

var (
	flagRemoveRPCHTTPIndex int
	flagRemoveRPCWSIndex   int
)

var removeRPCCmd = &cobra.Command{
	Use:   "remove-rpc",
	Short: "remove a rpc by index",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		currentChain, err := cfg.GetCurrentChain()
		if err != nil {
			return err
		}

		if flagRemoveRPCHTTPIndex != -1 {
			currentChain.RPCs = slices.Delete(currentChain.RPCs, flagRemoveRPCHTTPIndex, flagRemoveRPCHTTPIndex+1)
		}

		if flagRemoveRPCWSIndex != -1 {
			currentChain.WSRPCs = slices.Delete(currentChain.WSRPCs, flagRemoveRPCWSIndex, flagRemoveRPCWSIndex+1)
		}

		if err := cfg.Save(); err != nil {
			return err
		}

		utils.PrintTitle("OK")
		return nil
	},
}

func init() {
	removeRPCCmd.Flags().IntVar(&flagRemoveRPCHTTPIndex, "http", -1, "http/https rpc index")
	removeRPCCmd.Flags().IntVar(&flagRemoveRPCWSIndex, "ws", -1, "ws/wss rpc index")

	removeRPCCmd.MarkFlagsOneRequired("http", "ws")
}
