package amocrm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/WitnessBro/amocrm_to_google_sheets/internal/models"
)

const (
	events    = "events"
	leads     = "leads"
	contacts  = "contacts"
	companies = "companies"
	catalog   = "catalogs"
)

// NetworkError представляет сетевую ошибку
type NetworkError struct {
	Err error
}

func (e NetworkError) Error() string {
	return fmt.Sprintf("network error: %v", e.Err)
}

func (e NetworkError) Unwrap() error {
	return e.Err
}

type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

func NewClient(apikey string, baseURL string) Client {
	return Client{
		apiKey:     apikey,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: time.Second * 30},
	}
}
func (c *Client) GetEvents(ctx context.Context) (*models.Events, error) {
	queryParams := url.Values{}
	queryParams.Add("page", "1")
	queryParams.Add("limit", "50")
	resp, err := c.makeRequest(ctx, http.MethodGet, events, queryParams, nil)
	if err != nil {
		// Определяем и категоризируем ошибки
		if netErr, ok := err.(net.Error); ok {
			return nil, NetworkError{Err: netErr}
		}

		// Проверяем на http ошибки транспорта
		if urlErr, ok := err.(*url.Error); ok {
			if _, ok := urlErr.Err.(net.Error); ok {
				return nil, NetworkError{Err: err}
			}
		}

		return nil, err
	}

	var events models.Events
	if err := json.Unmarshal(resp, &events); err != nil {
		slog.Error("Ошибка декодирования JSON событий", "error", err, "response", string(resp[:min(len(resp), 1000)]))
		return nil, err
	}

	return &events, nil
}

// Вспомогательная функция для предотвращения выхода за границы
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (c *Client) GetContacts(ctx context.Context, contactID string) (*models.Contact, error) {

	endpoint := fmt.Sprintf("%s/%s", contacts, contactID)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var contacts models.Contact
	if err := json.Unmarshal(resp, &contacts); err != nil {
		return nil, err
	}
	return &contacts, nil
}

func (c *Client) GetCatalogElement(ctx context.Context, catalogID string, elementID string) (*models.CatalogElement, error) {
	endpoint := fmt.Sprintf("%s/%s/elements/%s", catalog, catalogID, elementID)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog element: %w", err)
	}

	var element models.CatalogElement
	if err := json.Unmarshal(resp, &element); err != nil {
		return nil, err
	}

	return &element, nil
}

func (c *Client) GetCompanies(ctx context.Context, companyID string) (*models.Company, error) {

	endpoint := fmt.Sprintf("%s/%s", companies, companyID)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var company models.Company
	if err := json.Unmarshal(resp, &company); err != nil {
		return nil, err
	}
	return &company, nil
}

func (c *Client) GetCompanyFull(ctx context.Context, companyID string) (*models.CompanyFull, error) {

	endpoint := fmt.Sprintf("%s/%s", companies, companyID)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var company models.CompanyFull
	if err := json.Unmarshal(resp, &company); err != nil {
		return nil, err
	}
	return &company, nil
}

func (c *Client) GetLeadData(ctx context.Context, leadID int) (*models.Lead, error) {

	endpoint := fmt.Sprintf("%s/%d", leads, leadID)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, url.Values{"with": []string{"companies", "contacts"}}, nil)
	if err != nil {
		return nil, err
	}

	var lead models.Lead
	if err := json.Unmarshal(resp, &lead); err != nil {
		return nil, err
	}
	return &lead, nil
}

func (c *Client) UpdateLeadData(ctx context.Context, leadID int, data any) error {

	endpoint := fmt.Sprintf("%s/%d", leads, leadID)

	_, err := c.makeRequest(ctx, http.MethodPatch, endpoint, nil, data)
	if err != nil {
		slog.Info(fmt.Sprintf("Поле в сделке {%d} успешно изменено", leadID))
		return err
	}
	return nil
}

func (c *Client) UpdateContact(ctx context.Context, contactID string, data any) error {

	endpoint := fmt.Sprintf("%s/%s", contacts, contactID)

	_, err := c.makeRequest(ctx, http.MethodPatch, endpoint, nil, data)
	if err != nil {
		slog.Info(fmt.Sprintf("Поле в контакте {%s} успешно изменено", contactID))
		return err
	}
	return nil
}

func (c *Client) makeRequest(ctx context.Context, method string, endpoint string, query url.Values, payload any) ([]byte, error) {
	// Формируем полный URL
	fullURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	if query != nil {
		fullURL = fmt.Sprintf("%s?%s", fullURL, query.Encode())
	}

	// Преобразуем payload в JSON, если он есть
	var body io.Reader
	if payload != nil {
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonPayload)
	}

	// Создаем новый запрос
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Выполняем запрос с проверкой на сетевые ошибки
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Проверяем типы сетевых ошибок и оборачиваем их
		if netErr, ok := err.(net.Error); ok {
			return nil, NetworkError{Err: netErr}
		}

		// Проверяем на http ошибки транспорта
		if urlErr, ok := err.(*url.Error); ok {
			if _, ok := urlErr.Err.(net.Error); ok {
				return nil, NetworkError{Err: err}
			}
		}

		return nil, fmt.Errorf("failed to do request: %w", err)
	}

	defer resp.Body.Close()

	// Проверяем HTTP статус
	if resp.StatusCode >= 400 {
		// Если ошибка связана с сетевым соединением или сервером
		if resp.StatusCode == http.StatusGatewayTimeout ||
			resp.StatusCode == http.StatusBadGateway ||
			resp.StatusCode == http.StatusServiceUnavailable {
			return nil, NetworkError{Err: fmt.Errorf("server error: %d", resp.StatusCode)}
		}
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}
