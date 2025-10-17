package main

import (
	"context"
	"log/slog"
	"os/exec"
	"time"

	"github.com/s-588/BOMViewer/cmd/http"
	"github.com/s-588/BOMViewer/internal/db"
)

const port = ":8080"

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	dbctx, _ := context.WithTimeout(ctx, time.Second*10)
	repo, err := db.NewRepository(dbctx, "Data Source=/bom.db;")
	if err != nil {
		slog.Info("Error creating repository:", "error", err)
		return
	}
	slog.Info("database connected")
	defer repo.Close()

	go func() {
		slog.Info("Starting server at " + port)
		err = http.NewServer(cancel, port).Start()
		if err != nil {
			slog.Error("Error starting server:", "error", err)
			cancel()
		}
	}()

	err = exec.Command("cmd", "/C", "start", "http://localhost"+port+"/welcome").Run()
	if err != nil {
		slog.Error("can't open browser", "error", err)
	}

	select {
	case <-ctx.Done():
		slog.Info("Stopping the app")
	}

}
