package main

import (
	"fmt"
	"gitlab.com/dgb9/todo-api/internal/start"
	"log/slog"
	"os"
)

func main() {
	// setup the logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	err := start.Start()
	if err == nil {
		slog.Info("successful exit")
	} else {
		slog.Error(fmt.Sprintf("error: %v", err))
	}
}
