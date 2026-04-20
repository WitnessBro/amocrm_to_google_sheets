package storage

import (
	"encoding/json"
	"os"

	"github.com/WitnessBro/amocrm_to_google_sheets/internal/models"
)

// SaveEventsToFile сохраняет список обработанных событий в файл
func SaveEventsToFile(events models.ProcessedEvents, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(events)
}

// LoadEventsFromFile загружает список обработанных событий из файла
func LoadEventsFromFile(filePath string) (models.ProcessedEvents, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Если файл не существует, возвращаем пустой список
		return models.ProcessedEvents{}, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events models.ProcessedEvents
	if err := json.NewDecoder(file).Decode(&events); err != nil {
		return nil, err
	}

	return events, nil
}
