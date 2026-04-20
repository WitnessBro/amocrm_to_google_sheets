package models

type Contact struct {
	ID                int                `json:"id"`
	Name              string             `json:"name"`
	FirstName         string             `json:"first_name"`
	LastName          string             `json:"last_name"`
	ResponsibleUserID int                `json:"responsible_user_id"`
	GroupID           int                `json:"group_id"`
	CreatedBy         int                `json:"created_by"`
	UpdatedBy         int                `json:"updated_by"`
	CreatedAt         int                `json:"created_at"`
	UpdatedAt         int                `json:"updated_at"`
	ClosestTaskAt     *int               `json:"closest_task_at"`
	IsDeleted         bool               `json:"is_deleted"`
	IsUnsorted        bool               `json:"is_unsorted"`
	CustomFields      []CustomFieldValue `json:"custom_fields_values"`
	AccountID         int                `json:"account_id"`
	Links             Links              `json:"_links"`
	Embedded          EmbeddedData       `json:"_embedded"`
}

type EmbeddedData struct {
	Tags      []Tag     `json:"tags"`
	Companies []Company `json:"companies"`
}
