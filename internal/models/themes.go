package models

// Theme представляет структуру данных для чата с курсами
type Theme struct {
	ChatID  int64          `json:"chat_id"`
	Courses map[string]int `json:"courses"`
}

// ThemesData представляет структуру данных со всеми чатами и их курсами
type ThemesData map[string]Theme
