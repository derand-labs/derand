package main

import (
	"bufio"
	"derand-cli/config"
	"derand-cli/utils"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "cleanup configuration",

	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := config.Load(); err != nil {
			return err
		}

		if !utils.Confirm("Clean up will delete all configurations. Continue? [y/N]: ") {
			return fmt.Errorf("cancelled")
		}

		if !utils.Confirm("All information will be lost (including wallet information). Continue? [y/N]:") {
			return fmt.Errorf("cancelled")
		}

		fmt.Printf("Enter 'clear all wallets': ")
		reader := bufio.NewReader(os.Stdin)

		answer, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		if strings.TrimSpace(answer) != "clear all wallets" {
			return fmt.Errorf("cancelled")
		}

		if err := config.Cleanup(); err != nil {
			return err
		}
		utils.PrintTitle("OK")
		return nil
	},
}
