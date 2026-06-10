package localprofile

import (
	"derand-cli/proverlogic"
	"time"

	"github.com/spf13/cobra"
)

var (
	flagBenchmarkName                    string
	flagBenchmarkTargetDelayTimeInSecond int
	flagBenchmarkTSeed                   uint64
	flagBenchmarkRepeats                 int
)

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "benchmark profile to get suitable parameter on the remote profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		prover, err := proverlogic.NewFindParameterProver(flagBenchmarkName)
		if err != nil {
			return err
		}

		targetDelayTime := time.Duration(flagBenchmarkTargetDelayTimeInSecond) * time.Second
		if err := prover.FindBestParameter(targetDelayTime, flagBenchmarkTSeed, flagBenchmarkRepeats); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	benchmarkCmd.Flags().StringVar(&flagBenchmarkName, "name", "", "local profile name")
	benchmarkCmd.Flags().Uint64Var(&flagBenchmarkTSeed, "t-seed", 30000000, "seed for VDF iteration count")
	benchmarkCmd.Flags().IntVar(&flagBenchmarkTargetDelayTimeInSecond, "target-delay-time", 12, "target delay time in second")
	benchmarkCmd.Flags().IntVar(&flagBenchmarkRepeats, "repeats", 5, "repeats for zkcircuit")

	benchmarkCmd.MarkFlagRequired("name")
}
