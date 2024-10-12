package main

import (
	"context"
	"github.com/re-tofl/tofl-gpt-chat/cmd/tgbot/app"
)

func main() {
	app.MustExecute(context.Background())
}
