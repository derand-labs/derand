package balance

import (
	"derand-cli/config"
	"derand-cli/gen"
	"derand-cli/utils"
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

var (
	flagAddress string
)

var Cmd = &cobra.Command{
	Use:   "balance",
	Short: "get the balance in derand",

	RunE: func(cmd *cobra.Command, args []string) error {
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

		var addr ethcommon.Address
		if flagAddress != "" {
			addr = ethcommon.HexToAddress(flagAddress)
		} else {
			currentWallet, err := cfg.GetCurrentWallet()
			if err != nil {
				return err
			}

			addr, err = currentWallet.GetAddress()
			if err != nil {
				return err
			}
		}

		balance, err := derand.BalanceOf(nil, addr)
		if err != nil {
			utils.PrintTitle(utils.Bold("Balance in DeRand:"), err)
		} else {
			utils.PrintTitle(utils.Bold("Balance in DeRand:"), utils.WeiToETHString(balance), currentChain.Symbol)
		}

		proverBalance, err := derand.ProverBalanceOf(nil, addr)
		if err != nil {
			utils.PrintTitle(utils.Bold("Prover Balance in DeRand:"), err)
		} else {
			utils.PrintTitle(utils.Bold("Prover Balance in DeRand:"), utils.WeiToETHString(proverBalance), currentChain.Symbol)
		}

		walletBalance, err := backend.BalanceAt(cmd.Context(), addr, nil)
		if err != nil {
			utils.PrintSubtitle("Balance in Wallet:", err)
		} else {
			utils.PrintTitle(utils.Bold("Balance in Wallet:"), utils.WeiToETHString(walletBalance), currentChain.Symbol)
		}

		return nil
	},
}

func init() {
	Cmd.Flags().StringVar(&flagAddress, "address", "", "specify address")

	Cmd.AddCommand(depositCmd)
	Cmd.AddCommand(withdrawCmd)
	Cmd.AddCommand(depositProverCmd)
	Cmd.AddCommand(withdrawProverCmd)
}
