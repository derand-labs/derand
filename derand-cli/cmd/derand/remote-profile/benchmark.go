package remoteprofile

import (
	"derand-cli/proverlogic"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	flagBenchmarkRepeats int
)

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "benchmark the remote profile to assess system suitability",

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("must specify remote profile id, run `derand remote-profile` to get more detail")
		}

		profileID, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid id: %w", err)
		}

		prover, err := proverlogic.NewBenchmarkProver(profileID)
		if err != nil {
			return err
		}

		if err := prover.Benchmark(flagBenchmarkRepeats); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	benchmarkCmd.Flags().IntVar(&flagBenchmarkRepeats, "repeats", 5, "number of repeats")
}
