package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/WitnessBro/amocrm_to_google_sheets/internal/amocrm"
	"github.com/WitnessBro/amocrm_to_google_sheets/internal/config"
	"github.com/WitnessBro/amocrm_to_google_sheets/internal/googlesheets"
	"github.com/WitnessBro/amocrm_to_google_sheets/internal/models"
)

type LeadProcessor struct {
	cfg          *config.Config
	amoClient    amocrm.Client
	sheetsClient *googlesheets.Client
}

func NewLeadProcessor(cfg *config.Config) (*LeadProcessor, error) {
	sheetsClient, err := googlesheets.NewClient(cfg.GoogleCredentialsPath, cfg.GoogleTableID)
	if err != nil {
		return nil, err
	}

	return &LeadProcessor{
		cfg:          cfg,
		amoClient:    amocrm.NewClient(cfg.AmoAPIKey, fmt.Sprintf("https://%s.amocrm.ru/api/v4/", cfg.AmoSubdomain)),
		sheetsClient: sheetsClient,
	}, nil
}

func (p *LeadProcessor) ProcessLead(lead *models.Lead) {
	ctx := context.Background()

	programDetails, err := p.extractProgramDetails(ctx, lead)
	if err != nil {
		slog.Error("Не удалось определить программу сделки", "lead_id", lead.ID, "error", err)
		return
	}

	clientType, listenersCount, renewal, folderLink, edo, comment, paymentStatus, contractNumber := extractLeadFields(lead)
	companyName := p.extractCompanyName(ctx, lead)
	contacts := p.extractListenerContacts(ctx, lead)

	leadLink := fmt.Sprintf("https://%s.amocrm.ru/leads/detail/%d", p.cfg.AmoSubdomain, lead.ID)
	leadLinkGoogle := fmt.Sprintf("=HYPERLINK(\"%s\"; \"%d\")", leadLink, lead.ID)

	for _, contact := range contacts {
		rowData := []any{
			time.Now().Format("2006-01-02 15:04:05"),
			leadLinkGoogle,                           // ID сделки как гиперссылка
			programDetails,                           // Программа
			renewal,                                  // Продление
			listenersCount,                           // Количество слушателей
			clientType,                               // Тип клиента
			paymentStatus,                            // Статус оплаты
			edo,                                      // ЭДО
			comment,                                  // Комментарий
			contractNumber,                           // Номер договора
			folderLink,                               // Ссылка на папку
			companyName,                              // Название компании
			contact.Name,
		}

		if err := p.sheetsClient.AppendRow(ctx, rowData); err != nil {
			slog.Error("Ошибка записи контакта в Google Sheets", "lead_id", lead.ID, "contact", contact.Name, "error", err)
			continue
		}

		slog.Info("Контакт записан в Google Sheets", "lead_id", lead.ID, "contact", contact.Name)
	}
}

func (p *LeadProcessor) extractProgramDetails(ctx context.Context, lead *models.Lead) (string, error) {
	var programElements []string
	for _, customField := range lead.CustomFields {
		if customField.FieldName != "Курс" || len(customField.Values) == 0 {
			continue
		}

		for _, value := range customField.Values {
			if value.CatalogID == nil || value.CatalogElementID == nil {
				continue
			}

			catalogID := fmt.Sprintf("%d", *value.CatalogID)
			elementID := fmt.Sprintf("%d", *value.CatalogElementID)
			element, err := p.amoClient.GetCatalogElement(ctx, catalogID, elementID)
			if err != nil {
				slog.Error("Ошибка получения элемента каталога", "catalog_id", catalogID, "element_id", elementID, "error", err)
				continue
			}

			programElements = append(programElements, element.Name)
		}
	}

	if len(programElements) == 0 {
		return "", fmt.Errorf("программа не найдена")
	}

	return strings.Join(programElements, "; "), nil
}

func extractLeadFields(lead *models.Lead) (clientType, listenersCount, renewal, folderLink, edo, comment, paymentStatus, contractNumber string) {
	for _, field := range lead.CustomFields {
		switch field.FieldName {
		case "Тип клиента":
			if len(field.Values) > 0 {
				clientType = fmt.Sprintf("%v", field.Values[0].Value)
			}
		case "Количество слушателей":
			if len(field.Values) > 0 {
				listenersCount = fmt.Sprintf("%v", field.Values[0].Value)
			}
		case "Продление":
			if len(field.Values) > 0 {
				renewal = fmt.Sprintf("%v", field.Values[0].Value)
			}
		case "y_folder":
			if len(field.Values) > 0 {
				folderLink = fmt.Sprintf("%v", field.Values[0].Value)
			}
		case "ЭДО":
			if len(field.Values) > 0 {
				if val, ok := field.Values[0].Value.(bool); ok && val {
					edo = "Есть"
				} else {
					edo = "Нет"
				}
			}
		case "Коммент":
			if len(field.Values) > 0 {
				comment = fmt.Sprintf("%v", field.Values[0].Value)
			}
		case "Статус оплаты":
			if len(field.Values) > 0 {
				paymentStatus = fmt.Sprintf("%v", field.Values[0].Value)
			}
		case "№договора":
			if len(field.Values) > 0 {
				contractNumber = fmt.Sprintf("%v", field.Values[0].Value)
			}
		}
	}

	return
}

func (p *LeadProcessor) extractCompanyName(ctx context.Context, lead *models.Lead) string {
	for _, companyRef := range lead.Embedded.Companies {
		if companyRef.ID == 1 {
			continue
		}

		company, err := p.amoClient.GetCompanyFull(ctx, fmt.Sprintf("%d", companyRef.ID))
		if err != nil {
			slog.Error("Ошибка получения компании", "company_id", companyRef.ID, "error", err)
			return fmt.Sprintf("Компания ID: %d", companyRef.ID)
		}

		return company.Name
	}

	return ""
}

type listenerContact struct {
	Name      string
	Phone     string
	Email     string
	Education string
}

func (p *LeadProcessor) extractListenerContacts(ctx context.Context, lead *models.Lead) []listenerContact {
	result := make([]listenerContact, 0, len(lead.Embedded.Contacts))

	for _, contactRef := range lead.Embedded.Contacts {
		contact, err := p.amoClient.GetContacts(ctx, fmt.Sprintf("%d", contactRef.ID))
		if err != nil {
			slog.Error("Ошибка получения данных контакта", "contact_id", contactRef.ID, "error", err)
			continue
		}

		isListener := false
		for _, customField := range contact.CustomFields {
			if customField.FieldName == "Слушатель" {
				isListener = true
				break
			}
		}
		if !isListener {
			continue
		}

		var phone, email, education string
		for _, customField := range contact.CustomFields {
			if customField.FieldCode != nil {
				if *customField.FieldCode == "PHONE" && len(customField.Values) > 0 {
					phone = fmt.Sprintf("%v", customField.Values[0].Value)
				} else if *customField.FieldCode == "EMAIL" && len(customField.Values) > 0 {
					email = fmt.Sprintf("%v", customField.Values[0].Value)
				}
			} else if customField.FieldName == "Образование" && len(customField.Values) > 0 {
				education = fmt.Sprintf("%v", customField.Values[0].Value)
			}
		}

		result = append(result, listenerContact{
			Name:      contact.Name,
			Phone:     phone,
			Email:     email,
			Education: education,
		})
	}

	return result
}
