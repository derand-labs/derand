package main

import (
	"derand-cli/cmd/derand/balance"
	"derand-cli/cmd/derand/chain"
	localprofile "derand-cli/cmd/derand/local-profile"
	"derand-cli/cmd/derand/prover"
	remoteprofile "derand-cli/cmd/derand/remote-profile"
	"derand-cli/cmd/derand/request"
	"derand-cli/cmd/derand/wallet"
	"derand-cli/config"
	"derand-cli/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	configDir string
	vdfDir    string
	buildDir  string
)

var Cmd = &cobra.Command{
	Use:           "derand",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		config.SetConfigDir(configDir)
		config.SetVDFDir(vdfDir)
		config.SetBuildDir(buildDir)

		cfg, err := config.Load()
		if err != nil {
			return nil
		}

		printed := false

		currentChain, err := cfg.GetCurrentChain()
		if err == nil {
			utils.PrintTitle(utils.Green("Current chain:"), currentChain.Name)
			printed = true
		}

		currentWallet, err := cfg.GetCurrentWallet()
		if err == nil {
			utils.PrintTitle(utils.Green("Current wallet:"), currentWallet.Name)
			printed = true
		}

		if printed {
			fmt.Println("========================================")
		}

		return nil
	},
}

func main() {
	if err := Cmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func init() {
	Cmd.PersistentFlags().StringVar(&configDir, "config-dir", "", "configuration directory")
	Cmd.PersistentFlags().StringVar(&vdfDir, "vdf-dir", "", "vdf configuration directory")
	Cmd.PersistentFlags().StringVar(&buildDir, "build-dir", "./build", "binary (zkvdf, corevdf) directory")
	Cmd.AddCommand(
		initCmd,
		cleanupCmd,
		verifyCmd,
		infoCmd,
		deployCmd,
	)
	Cmd.AddCommand(
		balance.Cmd,
		chain.Cmd,
		wallet.Cmd,
		localprofile.Cmd,
		remoteprofile.Cmd,
		prover.Cmd,
		request.Cmd,
	)
}
