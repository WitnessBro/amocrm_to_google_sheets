package models

type CatalogElement struct {
	ID                 int                    `json:"id"`
	Name               string                 `json:"name"`
	CreatedBy          int                    `json:"created_by"`
	UpdatedBy          int                    `json:"updated_by"`
	CreatedAt          int64                  `json:"created_at"`
	UpdatedAt          int64                  `json:"updated_at"`
	IsDeleted          *bool                  `json:"is_deleted"`
	CustomFieldsValues []CustomFieldValue     `json:"custom_fields_values"`
	CatalogID          int                    `json:"catalog_id"`
	AccountID          int                    `json:"account_id"`
	Links              Links                  `json:"_links"`
	Embedded           CatalogElementEmbedded `json:"_embedded"`
}

type CatalogElementEmbedded struct {
	Warning Warning `json:"warning"`
}

type Warning struct {
	Message *string `json:"message"`
}
