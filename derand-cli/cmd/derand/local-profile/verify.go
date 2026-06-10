package localprofile

import (
	"bytes"
	"derand-cli/config"
	"derand-cli/profile"
	"derand-cli/utils"
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

var (
	flagVerifyName     string
	flagVerifyAddress  string
	flagVerifyDeployTx string
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "verify smart contract of verifier",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		localProfile, ok := cfg.LocalProfiles[flagVerifyName]
		if !ok {
			return fmt.Errorf("not found local profile")
		}

		currentChain, err := cfg.GetCurrentChain()
		if err != nil {
			return err
		}

		backend, err := cfg.GetCurrentChainBackend()
		if err != nil {
			return err
		}

		if localProfile.Data.Type != "standard_classgroup_zk_plonk_bn254" {
			return fmt.Errorf("unsupported profile type: %s", localProfile.Data.Type)
		}

		address := ethcommon.Address{}
		txHash := ethcommon.Hash{}
		if flagVerifyAddress == "" {
			if _, ok := localProfile.Data.VerifierByChainID[currentChain.ChainID]; !ok {
				return fmt.Errorf("local profile verifier has not been deployed yet, please run `derand local-profile deploy` first!")
			}

			address = localProfile.Data.VerifierByChainID[currentChain.ChainID].Address
			txHash = localProfile.Data.VerifierByChainID[currentChain.ChainID].DeployTxHash
		} else {
			address = ethcommon.HexToAddress(flagVerifyAddress)
			txHash = ethcommon.HexToHash(flagVerifyDeployTx)
		}

		abi, bytecode, err := localProfile.Data.StandardClassgroupZKPlonkBn254.CompileVerifier(config.GetVDFDir())
		if err != nil {
			return fmt.Errorf("failed to compile: %w", err)
		}

		tx, _, err := backend.TransactionByHash(cmd.Context(), txHash)
		if err != nil {
			return fmt.Errorf("failed to find trasaction by hash: %w", err)
		}

		matched := -1
		for i, hashToPrime128 := range currentChain.HashToPrime128 {
			constructorData, err := abi.Pack("", hashToPrime128.Address)
			if err != nil {
				return err
			}

			fullByteCode := append(bytecode, constructorData...)
			if bytes.Equal(fullByteCode, tx.Data()) {
				matched = i
				break
			}
		}

		if matched == -1 {
			utils.PrintTitle(utils.Bold("Verifier:"), utils.Red("MISMATCH"))
		} else {
			utils.PrintTitle(utils.Bold("Verifier:"), utils.Green(fmt.Sprintf("MATCH [HashToPrime128-%d]", matched)))
		}

		if matched != -1 {
			if localProfile.Data.VerifierByChainID == nil {
				localProfile.Data.VerifierByChainID = make(map[int]profile.ContractInfo)
			}

			localProfile.Data.VerifierByChainID[currentChain.ChainID] = profile.ContractInfo{
				Address:      address,
				DeployTxHash: txHash,
			}

			cfg.LocalProfiles[flagVerifyName] = localProfile
			if err := cfg.Save(); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	verifyCmd.Flags().StringVar(&flagVerifyName, "name", "", "local profile name")
	verifyCmd.Flags().StringVar(&flagVerifyAddress, "address", "", "verify with another address, if success, set this address as verifier")
	verifyCmd.Flags().StringVar(&flagVerifyDeployTx, "deploy-tx", "", "the deployment transaction of above address")

	verifyCmd.MarkFlagRequired("name")
	verifyCmd.MarkFlagsRequiredTogether("address", "deploy-tx")
}
