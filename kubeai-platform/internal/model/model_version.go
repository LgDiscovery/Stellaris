package model

import (
	"gorm.io/datatypes"
	"time"
)

type ModelVersion struct {
	ID           uint           `gorm:"primary_key" json:"id"`
	ModelID      uint           `gorm:"not null;index" json:"model_id"`
	ModelVersion string         `gorm:"size:32;not null" json:"modelVersion"`
	StoragePath  string         `gorm:"size:512;not null" json:"storage_path"`
	Framework    string         `gorm:"size:32" json:"framework"`
	Metrics      datatypes.JSON `json:"metrics"`
	Status       string         `gorm:"size:16;default:active" json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
