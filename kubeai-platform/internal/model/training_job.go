package model

import (
	"gorm.io/datatypes"
	"time"
)

type TrainingJob struct {
	ID         uint           `grom:"primary_key" json:"id"`
	Name       string         `grom:"size:128;not null;uniqueIndex" json:"name"`
	ModelID    string         `json:"model_id"`
	Framework  string         `grom:"size:32" json:"framework"`
	Image      string         `grom:"size:256;not null" json:"image"`
	Command    string         `json:"command"`
	Resources  datatypes.JSON `grom:"not null" json:"resources"`
	Env        datatypes.JSON `json:"env"`
	volumes    datatypes.JSON `json:"volumes"`
	Status     string         `grom:"size:32;default:pending;index" json:"status"`
	K8sName    string         `grom:"size:256" json:"k8sName"`
	StartedAt  *time.Time     `json:"started_at,omitempty"`
	FinishedAt *time.Time     `json:"finished_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}
