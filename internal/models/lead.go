package models

type Lead struct {
	ID                int                `json:"id"`
	Name              string             `json:"name"`
	Price             int                `json:"price"`
	ResponsibleUserID int                `json:"responsible_user_id"`
	GroupID           int                `json:"group_id"`
	StatusID          int                `json:"status_id"`
	PipelineID        int                `json:"pipeline_id"`
	LossReasonID      *int               `json:"loss_reason_id"`
	CreatedBy         int                `json:"created_by"`
	UpdatedBy         int                `json:"updated_by"`
	CreatedAt         int                `json:"created_at"`
	UpdatedAt         int                `json:"updated_at"`
	ClosedAt          *int               `json:"closed_at"`
	ClosestTaskAt     int                `json:"closest_task_at"`
	IsDeleted         bool               `json:"is_deleted"`
	CustomFields      []CustomFieldValue `json:"custom_fields_values"`
	Score             *int               `json:"score"`
	AccountID         int                `json:"account_id"`
	LaborCost         *float64           `json:"labor_cost"`
	IsPriceComputed   bool               `json:"is_price_computed"`
	Links             Links              `json:"_links"`
	Embedded          LeadEmbeddedData   `json:"_embedded"`
}

type CustomFieldValue struct {
	FieldID    int          `json:"field_id"`
	FieldName  string       `json:"field_name"`
	FieldCode  *string      `json:"field_code"`
	FieldType  string       `json:"field_type"`
	Values     []FieldValue `json:"values"`
	IsComputed bool         `json:"is_computed"`
}

type FieldValue struct {
	Value            any     `json:"value,omitempty"`
	EnumID           *int    `json:"enum_id,omitempty"`
	EnumCode         *string `json:"enum_code,omitempty"`
	CatalogID        *int    `json:"catalog_id,omitempty"`
	CatalogElementID *int    `json:"catalog_element_id,omitempty"`
}

type Links struct {
	Self *SelfLink `json:"self,omitempty"`
	Next *SelfLink `json:"next,omitempty"`
	Prev *SelfLink `json:"prev,omitempty"`
}

type SelfLink struct {
	Href string `json:"href"`
}

type LeadEmbeddedData struct {
	Tags      []Tag         `json:"tags"`
	Companies []Company     `json:"companies"`
	Contacts  []LeadContact `json:"contacts"`
}

type Tag struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Color *string `json:"color"`
}

type Company struct {
	ID    int   `json:"id"`
	Links Links `json:"_links"`
}

type LeadContact struct {
	ID     int   `json:"id"`
	IsMain bool  `json:"is_main"`
	Links  Links `json:"_links"`
}
