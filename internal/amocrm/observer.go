package amocrm

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/WitnessBro/amocrm_to_google_sheets/internal/models"
	"github.com/WitnessBro/amocrm_to_google_sheets/internal/storage"
	"github.com/cenkalti/backoff/v5"
	"golang.org/x/exp/slices"
)

const (
	MaxProcessedEvents = 1000
	EventsFilePath     = "configs/events.json"
)

// Observer представляет наблюдатель за событиями AmoCRM
type Observer struct {
	client              Client
	interval            time.Duration
	processedEvents     models.ProcessedEvents
	stopCh              chan struct{}
	notificationFn      func(lead *models.Lead)
	errorNotificationFn func(message string)
	prodFieldID         string
}

// NewObserver создает новый наблюдатель AmoCRM
func NewObserver(
	client Client,
	interval time.Duration,
	prodFieldID string,
	notificationFn func(lead *models.Lead),
	errorNotificationFn func(message string),
) *Observer {
	return &Observer{
		client:              client,
		interval:            interval,
		stopCh:              make(chan struct{}),
		processedEvents:     models.ProcessedEvents{},
		notificationFn:      notificationFn,
		errorNotificationFn: errorNotificationFn,
		prodFieldID:         prodFieldID,
	}
}

// Run запускает наблюдатель в бесконечном цикле
func (o *Observer) Run(ctx context.Context) {
	slog.Info("Наблюдатель AmoCRM запущен")

	// Загрузка обработанных событий из файла
	o.LoadEventsFromFile()

	ticker := time.NewTicker(o.interval)
	defer ticker.Stop()

	// Сразу проверяем обновления при запуске
	o.CheckForUpdates(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Наблюдатель AmoCRM остановлен по контексту")
			return
		case <-o.stopCh:
			slog.Info("Наблюдатель AmoCRM остановлен")
			return
		case <-ticker.C:
			slog.Info("Проверка обновлений AmoCRM...")
			o.CheckForUpdates(ctx)
			slog.Info(fmt.Sprintf("Пауза %v...", o.interval))
		}
	}
}

// Stop останавливает наблюдатель
func (o *Observer) Stop() {
	close(o.stopCh)
}

// retryOperation выполняет операцию с повторными попытками
func retryOperation(ctx context.Context, operation func() error, b *backoff.ExponentialBackOff, notify func(error, time.Duration)) error {
	// Попытки с экспоненциальной задержкой
	var err error
	for {
		// Проверка отмены контекста
		select {
		case <-ctx.Done():
			return ctx.Err() // Возвращаем ошибку контекста, если он отменен
		default:
			// Продолжаем выполнение
		}

		err = operation()
		if err == nil {
			return nil // Успешное выполнение
		}

		// Проверяем на постоянную ошибку
		if permanent, ok := err.(*backoff.PermanentError); ok {
			return permanent.Unwrap()
		}

		// Вычисляем задержку перед следующей попыткой
		nextBackoff := b.NextBackOff()
		if nextBackoff == backoff.Stop {
			break // Достигли лимита попыток
		}

		// Оповещаем о повторной попытке
		if notify != nil {
			notify(err, nextBackoff)
		}

		// Ждем перед следующей попыткой с проверкой контекста
		timer := time.NewTimer(nextBackoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			// Продолжаем следующую попытку
		}
	}

	return err
}

// CheckForUpdates проверяет обновления в AmoCRM с использованием механизма повторных попыток
func (o *Observer) CheckForUpdates(ctx context.Context) {
	// Создаем конфигурацию экспоненциальной задержки
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 2 * time.Second
	b.MaxInterval = 60 * time.Second

	// Функция, которая будет выполняться с повторными попытками
	operation := func() error {
		events, err := o.client.GetEvents(ctx)
		if err != nil {
			// Проверяем тип ошибки
			if isNetworkError(err) {
				slog.Error("Ошибка сети при получении событий из AmoCRM", "error", err)
				o.errorNotificationFn(fmt.Sprintf("Ошибка сети при получении событий из AmoCRM: %v. Повторная попытка...", err))
				return err // Возвращаем ошибку для повторной попытки
			}

			// Для других типов ошибок логируем и не делаем повторных попыток
			slog.Error("Ошибка получения событий из AmoCRM", "error", err)
			o.errorNotificationFn(fmt.Sprintf("Ошибка получения событий из AmoCRM: %v", err))
			return backoff.Permanent(err) // Постоянная ошибка, не пытаемся повторить
		}

		slog.Info(fmt.Sprintf("Получено %d событий", len(events.Embedded.Events)))

		// Обрабатываем события в обратном порядке (от новых к старым)
		if len(events.Embedded.Events) > 0 {
			for i := len(events.Embedded.Events) - 1; i >= 0; i-- {
				event := events.Embedded.Events[i]

				// Проверяем, было ли событие уже обработано
				if !o.isEventProcessed(event.ID) {
					slog.Info(fmt.Sprintf("Обработка события: %s", event.ID))
					o.ProcessEvent(ctx, &event)
					o.AddProcessedEvent(event.ID)
				}
			}

			// Сохраняем обработанные события в файл
			o.SaveEventsToFile()
		}

		return nil // Успешное выполнение
	}

	// Функция для отслеживания повторных попыток
	notify := func(err error, duration time.Duration) {
		slog.Info(fmt.Sprintf("Повторная попытка через %v после ошибки: %v", duration, err))
	}

	// Выполняем операцию с повторными попытками
	err := retryOperation(ctx, operation, b, notify)
	if err != nil {
		slog.Error("Не удалось получить события после нескольких попыток", "error", err)
	}
}

// isNetworkError проверяет, является ли ошибка сетевой ошибкой
func isNetworkError(err error) bool {
	// Проверяем различные типы сетевых ошибок
	if _, ok := err.(net.Error); ok {
		return true
	}

	// Проверяем на наш тип NetworkError
	if _, ok := err.(NetworkError); ok {
		return true
	}

	// Можно добавить дополнительные проверки для других типов сетевых ошибок
	return false
}

// ProcessEvent обрабатывает отдельное событие
func (o *Observer) ProcessEvent(ctx context.Context, event *models.Event) {
	// Проверяем, является ли событие изменением нужного поля
	targetFieldEvent := fmt.Sprintf("custom_field_%s_value_changed", o.prodFieldID)

	if event.Type == targetFieldEvent {
		// Проверяем, есть ли данные о сущности
		if event.Embedded.Entity.ID != 0 {
			leadID := event.EntityID

			// Проверяем, содержит ли значение "1"
			hasTargetValue := false
			for _, value := range event.ValueAfter {
				if value.CustomFieldValue != nil && value.CustomFieldValue.Text == "1" {
					hasTargetValue = true
					break
				}
			}

			if hasTargetValue {
				// Получаем данные сделки
				lead, err := o.client.GetLeadData(ctx, leadID)
				if err != nil {
					slog.Error("Не удалось получить данные сделки", "lead_id", leadID, "error", err)
					o.errorNotificationFn(fmt.Sprintf("Не удалось получить данные сделки %d: %v", leadID, err))
					return
				}

				slog.Info(fmt.Sprintf("Данные сделки получены: %s", lead.Name))
				o.notificationFn(lead)
			} else {
				slog.Info(fmt.Sprintf("Поле изменено, но значение не True: %d", leadID))
			}
		} else {
			slog.Info(fmt.Sprintf("Событие без информации о сущности: %s", event.ID))
		}
	}
}

// isEventProcessed проверяет, было ли событие уже обработано
func (o *Observer) isEventProcessed(eventID string) bool {
	return slices.Contains(o.processedEvents, eventID)
}

// AddProcessedEvent добавляет ID события в список обработанных
func (o *Observer) AddProcessedEvent(eventID string) {
	o.processedEvents = append(o.processedEvents, eventID)

	// Ограничиваем размер списка обработанных событий
	if len(o.processedEvents) > MaxProcessedEvents {
		o.processedEvents = o.processedEvents[1:]
	}
}

// SaveEventsToFile сохраняет обработанные события в файл
func (o *Observer) SaveEventsToFile() {
	err := storage.SaveEventsToFile(o.processedEvents, EventsFilePath)
	if err != nil {
		slog.Error("Ошибка сохранения обработанных событий в файл", "error", err)
	}
}

// LoadEventsFromFile загружает обработанные события из файла
func (o *Observer) LoadEventsFromFile() {
	events, err := storage.LoadEventsFromFile(EventsFilePath)
	if err != nil {
		slog.Error("Ошибка загрузки обработанных событий из файла", "error", err)
		return
	}

	o.processedEvents = events
	slog.Info(fmt.Sprintf("Загружено %d событий из файла", len(events)))
}
