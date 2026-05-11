package model

import "time"

type Model struct {
	ID          uint           `gorm:"primary_key" json:"id"`
	Name        string         `gorm:"uniqueIndex;size:128;not null" json:"name"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	Versions    []ModelVersion `gorm:"foreignKey:ModelID" json:"versions,omitempty"`
}
