package main

import (
	"fmt"
	"os"
	"zkvdf/cmd/zkvdf/common"
	"zkvdf/cmd/zkvdf/compile"
	"zkvdf/cmd/zkvdf/prove"
	"zkvdf/cmd/zkvdf/setup"
	"zkvdf/cmd/zkvdf/verify"

	"github.com/spf13/cobra"
)

var (
	setupdir string
)

var Cmd = &cobra.Command{
	Use:           "zkvdf",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		common.SetupRootDir(setupdir)
		return nil
	},
}

func main() {
	if err := Cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	Cmd.PersistentFlags().StringVar(&setupdir, "setup-dir", ".vdf", "setup directory")

	Cmd.AddCommand(compile.Cmd)
	Cmd.AddCommand(setup.Cmd)
	Cmd.AddCommand(prove.Cmd)
	Cmd.AddCommand(verify.Cmd)
}
