package main

import (
	"context"
	"github.com/re-tofl/tofl-gpt-chat/cmd/migrate/app"
)

func main() {
	app.MustExecute(context.Background())
}
