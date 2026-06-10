package wallet

import (
	"derand-cli/config"
	"derand-cli/utils"
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

var getPrivateCmd = &cobra.Command{
	Use:   "get-private",
	Short: "get private key of the current wallet",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		currentWallet, err := cfg.GetCurrentWallet()
		if err != nil {
			return err
		}

		addr, err := currentWallet.GetAddress()
		if err != nil {
			return err
		}

		utils.PrintTitle(utils.Bold("Name:"), currentWallet.Name)
		utils.PrintTitle(utils.Bold("Address:"), addr)

		password := utils.AskPassword("Enter password: ")
		privKey, err := currentWallet.GetPrivKey(password)
		if err != nil {
			return fmt.Errorf("incorrect password: %w", err)
		}

		utils.PrintTitle("Private key")
		utils.PrintSubtitle("0x" + ethcommon.Bytes2Hex(crypto.FromECDSA(privKey)))

		return nil
	},
}
