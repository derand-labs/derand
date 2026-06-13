package prover

import (
	"derand-cli/config"
	"derand-cli/gen"
	"derand-cli/proverlogic"
	"derand-cli/utils"
	"fmt"
	"slices"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "wait and prove the request as soon as you are assigned",

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		currentChain, err := cfg.GetCurrentChain()
		if err != nil {
			return err
		}

		currentWallet, err := cfg.GetCurrentWallet()
		if err != nil {
			return err
		}

		backend, err := cfg.GetCurrentChainBackend()
		if err != nil {
			return err
		}

		derand, err := gen.NewDeRand(currentChain.DeRand.Address, backend)
		if err != nil {
			return err
		}

		addr, err := currentWallet.GetAddress()
		if err != nil {
			return err
		}

		registeredProfiles, err := derand.RegisteredProfilesOf(nil, addr)
		if err != nil {
			return err
		}

		password := utils.AskPassword("Enter password: ")

		if len(registeredProfiles) == 0 {
			return fmt.Errorf("please register a remote profile first")
		}

		eventChan := make(chan *gen.DeRandAssignRequest, 1)

		latestBlockNumber, err := backend.BlockNumber(cmd.Context())
		if err != nil {
			return err
		}

		if currentChain.LastProverWatchedBlock < latestBlockNumber-50000 {
			currentChain.LastProverWatchedBlock = latestBlockNumber - 49990
		}

		logs, err := derand.FilterAssignRequest(&bind.FilterOpts{Start: currentChain.LastProverWatchedBlock}, nil, []common.Address{addr})
		if err != nil {
			return err
		}

		go func() {
			for logs.Next() {
				eventChan <- logs.Event
			}

			if err := logs.Error(); err != nil {
				fmt.Println("log iteration error:", err)
			}
		}()

		s, err := derand.WatchAssignRequest(&bind.WatchOpts{Start: &currentChain.LastProverWatchedBlock}, eventChan, nil, []common.Address{addr})
		if err != nil {
			return err
		}
		defer s.Unsubscribe()

		stop := false
		for !stop {
			fmt.Println("================================================================")
			utils.PrintTitle("Waiting for new request")
			select {
			case <-s.Err():
				stop = true

			case ev := <-eventChan:
				utils.PrintTitle(fmt.Sprintf("Proving request tx %s from block %d", ev.Raw.TxHash, ev.Raw.BlockNumber))

				prover, err := proverlogic.NewProver(ev.RequestId, true, password)
				if err != nil {
					return err
				}

				if err := prover.Prove(); err != nil {
					return err
				}

				cfg, err := config.Load()
				if err != nil {
					return err
				}

				currentChain, err := cfg.GetCurrentChain()
				if err != nil {
					return err
				}

				if currentChain.LastProverWatchedBlock <= ev.Raw.BlockNumber {
					currentChain.LastProverWatchedBlock = ev.Raw.BlockNumber + 1
					if err := cfg.Save(); err != nil {
						return err
					}
				}

				currentRegisteredProfiles, err := derand.RegisteredProfilesOf(nil, addr)
				if err != nil {
					return err
				}

				for _, oldProfile := range registeredProfiles {
					if !slices.Contains(currentRegisteredProfiles, oldProfile) {
						utils.PrintTitle(utils.Red(
							fmt.Sprintf(
								"[IMPORTANT] You are automatically unregistered from profile %d[%d]",
								oldProfile.Id, oldProfile.Version,
							)))
					}
				}
				// Do not re-assign registeredProfile by newRegisteredProfiles.
				// The reason is the log for proving a request is too long, the
				// prover may not see the IMPORTANT warnings of unregistering.
				// So that we need to repeat this warning every request.
			}
		}

		return nil
	},
}
