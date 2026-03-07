package cronjob

import (
	"announcer/config"
	"announcer/internal/app"
	"announcer/pkg/logger"

	"github.com/spf13/cobra"
)

func AnnounceBreakfast() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "breakfast",
		Short: "Announce breakfast to Discord",
		Run: func(*cobra.Command, []string) {
			cfg := config.LoadConfig()
			logger.InitZerolog()
			app.RunAnnounceBreakfast(cfg.BreakfastConfig)
		},
	}

	return cmd
}
