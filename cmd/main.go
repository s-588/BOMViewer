package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/s-588/BOMViewer/cmd/config"
	"github.com/s-588/BOMViewer/cmd/http"
	"github.com/s-588/BOMViewer/internal/db"
)

func main() {
	cfg, err := config.NewConfig(config.ConfigName)
	if err != nil {
		panic(err)
	}
	err = initFolders(cfg)
	if err != nil {
		panic(err)
	}

	logPath := filepath.Join(cfg.BaseDirectory, fmt.Sprintf("log_%s.log",
		time.Now().Format("2006-01-02_15-04-05")))
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(
		io.MultiWriter(os.Stdout, logFile), &slog.HandlerOptions{
			AddSource: true,
			Level:     parseLogLevel(cfg.LogCfg.LogLevel),
		})))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo, err := db.NewRepository(ctx, cfg)
	if err != nil {
		slog.Info("error creating repository:", "error", err)
		return
	}
	slog.Info("database connected", "dbPath", cfg.DBCfg.DBName)
	defer repo.Close()

	portChan := make(chan int)
	go func() {
		slog.Info("starting server", "port", cfg.ServerCfg.ServerPort)

		err = http.NewServer(cancel, cfg.ServerCfg.ServerPort, repo).Start(portChan)
		if err != nil {
			slog.Error("error starting server:", "error", err)
			cancel()
		}
	}()
	time.Sleep(100 * time.Millisecond)

	port := <-portChan
	slog.Info("opening browser", "addr", fmt.Sprintf("http://localhost:%d/welcome", port))
	err = exec.Command("cmd", "/C", "start", fmt.Sprintf("http://localhost:%d/welcome", port)).Run()
	if err != nil {
		slog.Error("can't open browser", "error", err)
	}

	<-ctx.Done()
	slog.Info("received shutdown signal, stopping the app")
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "DEBUG":
		return slog.LevelDebug
	case "ERROR":
		return slog.LevelError
	case "WARN", "WARNING":
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

func initFolders(cfg config.Config) error {
	err := os.MkdirAll(cfg.BaseDirectory, 0644)
	if err != nil {
		return fmt.Errorf("can't create base folders: %w", err)
	}

	uploadsPath := filepath.Join(cfg.BaseDirectory, cfg.ServerCfg.UploadsDir)
	err = os.MkdirAll(uploadsPath, 0644)
	if err != nil {
		return fmt.Errorf("can't create uploads folders: %w", err)
	}

	return nil
}
