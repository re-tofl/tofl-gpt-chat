package app

import (
	"context"

	"github.com/re-tofl/tofl-gpt-chat/cmd"
)

var rootCmd = cmd.Init("server")

func MustExecute(ctx context.Context) {
	rootCmd.MustExecute(ctx)
}
