package models

type Events struct {
	Page     int            `json:"_page"`
	Links    Links          `json:"_links"`
	Embedded EventsEmbedded `json:"_embedded"`
}

type EventsEmbedded struct {
	Events []Event `json:"events"`
}

type Event struct {
	ID          string        `json:"id"`
	Type        string        `json:"type"`
	EntityID    int           `json:"entity_id"`
	EntityType  string        `json:"entity_type"`
	CreatedBy   int           `json:"created_by"`
	CreatedAt   int64         `json:"created_at"`
	ValueAfter  []EventValue  `json:"value_after"`
	ValueBefore []EventValue  `json:"value_before"`
	AccountID   int           `json:"account_id"`
	Links       Links         `json:"_links"`
	Embedded    EventEmbedded `json:"_embedded"`
}

type EventValue struct {
	Note             *NoteValue             `json:"note,omitempty"`
	CustomFieldValue *CustomFieldEventValue `json:"custom_field_value,omitempty"`
	Link             *LinkValue             `json:"link,omitempty"`
	Tag              *TagValue              `json:"tag,omitempty"`
	SaleFieldValue   *SaleFieldValue        `json:"sale_field_value,omitempty"`
	NameFieldValue   *NameFieldValue        `json:"name_field_value,omitempty"`
}

type NoteValue struct {
	ID int `json:"id"`
}

type CustomFieldEventValue struct {
	FieldID   int    `json:"field_id"`
	FieldType int    `json:"field_type"`
	EnumID    int    `json:"enum_id"`
	Text      string `json:"text"`
}

type LinkValue struct {
	Entity EntityReference `json:"entity"`
}

type TagValue struct {
	Name string `json:"name"`
}

type SaleFieldValue struct {
	Sale float64 `json:"sale"`
}

type NameFieldValue struct {
	Name string `json:"name"`
}

type EntityReference struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
}

type EventEmbedded struct {
	Entity EventEntity `json:"entity"`
}

type EventEntity struct {
	ID         int   `json:"id"`
	IsDeleted  bool  `json:"is_deleted,omitempty"`
	IsUnsorted bool  `json:"is_unsorted,omitempty"`
	Links      Links `json:"_links"`
}
