package cmd

import (
	"announcer/cmd/cronjob"

	"github.com/spf13/cobra"
)

func cronjobCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cronjob",
		Short: "Run a cronjob with job name",
	}

	cmd.AddCommand(cronjob.AnnounceBreakfast())

	return cmd
}
