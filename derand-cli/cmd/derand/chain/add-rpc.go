package chain

import (
	"derand-cli/config"
	"derand-cli/utils"

	"github.com/spf13/cobra"
)

var (
	flagAddRPCHTTP string
	flagAddRPCWS   string
)

var addRPCCmd = &cobra.Command{
	Use:   "add-rpc",
	Short: "add new rpc",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		currentChain, err := cfg.GetCurrentChain()
		if err != nil {
			return err
		}

		if flagAddRPCHTTP != "" {
			currentChain.RPCs = append(currentChain.RPCs, flagAddRPCHTTP)
		}

		if flagAddRPCWS != "" {
			currentChain.WSRPCs = append(currentChain.WSRPCs, flagAddRPCWS)
		}

		if err := cfg.Save(); err != nil {
			return err
		}

		utils.PrintTitle("OK")
		return nil
	},
}

func init() {
	addRPCCmd.Flags().StringVar(&flagAddRPCHTTP, "http", "", "http/https rpc")
	addRPCCmd.Flags().StringVar(&flagAddRPCWS, "ws", "", "ws/wss rpc")

	addRPCCmd.MarkFlagsOneRequired("http", "ws")
}
