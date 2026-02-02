package cmd

import (
	"github.com/spf13/cobra"
)

func Execute() {
	var rootCmd = &cobra.Command{
		Use:   "announcer",
		Short: "This is a short description",
	}

	rootCmd.AddCommand(
		cronjobCommand(),
	)

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
