package app

import (
	"github.com/re-tofl/tofl-gpt-chat/internal/app"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "poll",
		Short: "Starts Telegram Bot polling",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.Run(cmd.Context(), &app.PollEntrypoint{Config: rootCmd.Config})
		},
	})
}
