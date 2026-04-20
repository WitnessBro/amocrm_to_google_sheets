package googlesheets

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Client struct {
	service *sheets.Service
	tableID string
}

// NewClient создает новый клиент для работы с Google Sheets
func NewClient(credentialsPath, tableID string) (*Client, error) {
	ctx := context.Background()

	service, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return nil, fmt.Errorf("unable to create sheets service: %v", err)
	}

	return &Client{
		service: service,
		tableID: tableID,
	}, nil
}

// AppendRow добавляет новую строку в первую пустую строку таблицы
func (c *Client) AppendRow(ctx context.Context, values []any) error {
	// Используем стандартное имя первого листа
	writeRange := "Sheet1!A:A"

	valueRange := &sheets.ValueRange{
		Values: [][]any{values},
	}

	// Используем APPEND для добавления в первую пустую строку
	_, err := c.service.Spreadsheets.Values.Append(c.tableID, writeRange, valueRange).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Context(ctx).
		Do()

	if err != nil {
		return fmt.Errorf("unable to append data to sheet: %v", err)
	}

	slog.Info("Данные успешно добавлены в Google Sheets", "table_id", c.tableID)
	return nil
}
