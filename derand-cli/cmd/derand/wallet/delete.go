package wallet

import (
	"bytes"
	"derand-cli/config"
	"derand-cli/utils"
	"fmt"
	"slices"
	"strconv"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete wallet",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(args) < 1 {
			return fmt.Errorf("must specify wallet index, run `derand wallet ls` to get more detail")
		}

		if len(args) > 1 {
			return fmt.Errorf("derand wallet delete [index]")
		}

		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid wallet index: %w", err)
		}

		currentWallet, err := cfg.GetWallet(index)
		if err != nil {
			return err
		}

		if !utils.Confirm("%s Delete wallet '%s'? [y/N]: ", utils.Bold("[IMPORTANT]"), currentWallet.Name) {
			return fmt.Errorf("cancelled")
		}

		password := utils.AskPassword("Enter password: ")
		privKey, err := currentWallet.GetPrivKey(password)
		if err != nil {
			return fmt.Errorf("incorrect password: %w", err)
		}

		if !bytes.Equal(crypto.FromECDSAPub(&privKey.PublicKey), cfg.Wallets[index].PublicKey) {
			return fmt.Errorf("incorrect password")
		}

		if !utils.Confirm("%s This action cannot be undone. Continue? [y/N]: ", utils.Bold("[IMPORTANT]")) {
			return fmt.Errorf("cancelled")
		}

		cfg.Wallets = slices.Delete(cfg.Wallets, index, index+1)
		if cfg.CurrentWalletIndex == index {
			fmt.Println("You deleted the current wallet, switched to wallet 0")
			cfg.CurrentWalletIndex = 0
		}

		if err := cfg.Save(); err != nil {
			return err
		}

		utils.PrintTitle("OK")
		return nil
	},
}
