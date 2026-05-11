package model

type Metadata struct {
	ModelName    string `json:"modelName"`
	ModelVersion string `json:"modelVersion"`
	StoragePath  string `json:"storagePath"`
}
