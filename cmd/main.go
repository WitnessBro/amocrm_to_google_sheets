package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/WitnessBro/amocrm_to_google_sheets/internal/amocrm"
	"github.com/WitnessBro/amocrm_to_google_sheets/internal/config"
	"github.com/WitnessBro/amocrm_to_google_sheets/internal/pipeline"
)

const (
	observerIntervalSeconds = 60
	prodFieldID             = "642629"
)

func main() {
	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		slog.Error("Ошибка загрузки конфигурации", "error", err)
		os.Exit(1)
	}

	amoClient := amocrm.NewClient(cfg.AmoAPIKey, fmt.Sprintf("https://%s.amocrm.ru/api/v4/", cfg.AmoSubdomain))
	processor, err := pipeline.NewLeadProcessor(cfg)
	if err != nil {
		slog.Error("Ошибка инициализации обработчика", "error", err)
		os.Exit(1)
	}

	observer := amocrm.NewObserver(
		amoClient,
		time.Duration(observerIntervalSeconds)*time.Second,
		prodFieldID,
		processor.ProcessLead,
		func(message string) {
			slog.Error("Ошибка обработки", "message", message)
		},
	)

	slog.Info("Запуск наблюдателя amoCRM")
	observer.Run(ctx)
}
