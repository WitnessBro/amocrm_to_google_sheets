package models

// CompanyFull представляет полную информацию о компании из AmoCRM API
type CompanyFull struct {
	ID                int                `json:"id"`
	Name              string             `json:"name"`
	ResponsibleUserID int                `json:"responsible_user_id"`
	GroupID           int                `json:"group_id"`
	CreatedBy         int                `json:"created_by"`
	UpdatedBy         int                `json:"updated_by"`
	CreatedAt         int                `json:"created_at"`
	UpdatedAt         int                `json:"updated_at"`
	ClosestTaskAt     *int               `json:"closest_task_at"`
	IsDeleted         bool               `json:"is_deleted"`
	CustomFields      []CustomFieldValue `json:"custom_fields_values"`
	AccountID         int                `json:"account_id"`
	Links             Links              `json:"_links"`
}
