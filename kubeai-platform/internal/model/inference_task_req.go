package model

import "gorm.io/datatypes"

type InferenceTaskRequest struct {
	ModelName    string          `json:"model_name"`
	ModelVersion string          `json:"model_version"`
	ModelPath    string          `json:"model_path"`
	Framework    string          `json:"framework"`
	Resources    ResourceRequest `json:"resources"`
	InputData    datatypes.JSON  `json:"input_data"`
}
