package chain

import (
	"derand-cli/config"
	"derand-cli/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagAddName                  string
	flagAddSymbol                string
	flagAddChainID               int
	flagAddRPCs                  []string
	flagAddWSRPCs                []string
	flagAddDeRandAddress         string
	flagAddDeRandTxHash          string
	flagAddHashToPrime128Address string
	flagAddHashToPrime128TxHash  string
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add chain info",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		for i := range cfg.Chains {
			if flagAddName == cfg.Chains[i].Name {
				return fmt.Errorf("duplicate chain name with chain id %d", cfg.Chains[i].ChainID)
			}
		}

		var derand *config.ContractInfo
		var hashToPrime128 []config.ContractInfo
		if flagAddDeRandAddress != "" {
			derand = new(config.ContractInfo)
			if err := derand.Address.UnmarshalText([]byte(flagAddDeRandAddress)); err != nil {
				return fmt.Errorf("invalid DeRand address: %w", err)
			}
			if err := derand.DeployTxHash.UnmarshalText([]byte(flagAddDeRandTxHash)); err != nil {
				return fmt.Errorf("invalid DeRand tx hash: %w", err)
			}
		}

		if flagAddHashToPrime128Address != "" {
			hashToPrime128 = append(hashToPrime128, config.ContractInfo{})
			if err := hashToPrime128[0].Address.UnmarshalText([]byte(flagAddHashToPrime128Address)); err != nil {
				return fmt.Errorf("invalid HashToPrime128 address: %w", err)
			}
			if err := hashToPrime128[0].DeployTxHash.UnmarshalText([]byte(flagAddHashToPrime128TxHash)); err != nil {
				return fmt.Errorf("invalid HashToPrime128 tx hash: %w", err)
			}
		}

		cfg.Chains = append(cfg.Chains, config.ChainInfo{
			Name:           flagAddName,
			Symbol:         flagAddSymbol,
			ChainID:        flagAddChainID,
			RPCs:           flagAddRPCs,
			WSRPCs:         flagAddWSRPCs,
			DeRand:         derand,
			HashToPrime128: hashToPrime128,
		})
		if err := cfg.Save(); err != nil {
			return err
		}

		utils.PrintTitle("OK")
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&flagAddName, "name", "", "chain name")
	addCmd.Flags().StringVar(&flagAddSymbol, "symbol", "", "chain native coin symbol")
	addCmd.Flags().IntVar(&flagAddChainID, "chain-id", 0, "chain id")
	addCmd.Flags().StringArrayVar(&flagAddRPCs, "rpc", []string{}, "rpc urls")
	addCmd.Flags().StringArrayVar(&flagAddWSRPCs, "ws-rpc", []string{}, "ws rpc urls")
	addCmd.Flags().StringVar(&flagAddDeRandAddress, "derand", "", "DeRand smart contract address")
	addCmd.Flags().StringVar(&flagAddDeRandTxHash, "derand-tx", "", "DeRand smart contract deployment transaction")
	addCmd.Flags().StringVar(&flagAddHashToPrime128Address, "hash-to-prime-128", "", "HashToPrime128 contract address")
	addCmd.Flags().StringVar(&flagAddHashToPrime128TxHash, "hash-to-prime-128-tx", "", "HashToPrime128 contract deployment transaction")

	addCmd.MarkFlagRequired("name")
	addCmd.MarkFlagRequired("symbol")
	addCmd.MarkFlagRequired("chain-id")
	addCmd.MarkFlagsRequiredTogether("derand", "derand-tx")
	addCmd.MarkFlagsRequiredTogether("hash-to-prime-128", "hash-to-prime-128-tx")
}
