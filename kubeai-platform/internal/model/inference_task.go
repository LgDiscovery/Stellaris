package model

import (
	"gorm.io/datatypes"
	"time"
)

type InferenceTask struct {
	TaskID       string          `json:"task_id"`
	ModelName    string          `json:"model_name"`
	ModelVersion string          `json:"model_version"`
	ModelPath    string          `json:"model_path"`
	Framework    string          `json:"framework"`
	Resources    ResourceRequest `json:"resources"`
	InputData    datatypes.JSON  `json:"input_data"`
	CreateTime   time.Time       `json:"create_time"`
	Status       string          `json:"status"`
}
