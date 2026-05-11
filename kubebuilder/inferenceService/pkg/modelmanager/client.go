package modelmanager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client  is the HTTP client for the Model Manager
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Model Manager client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: time.Second * 10, // 设置超时
		},
	}
}

// ModelMetadata represents the response from the Model Manager API.
type ModelMetadata struct {
	ModelName    string `json:"modelName"`
	ModelVersion string `json:"modelVersion"`
	// 模型在 MinIO/S3 中的路径，例如: s3://models/bert/v1/model.tar.gz
	StoragePath string `json:"storagePath"`
}

// GetModelMetadata GetModel fetches the model metadata
func (c *Client) GetModelMetadata(modelName, modelVersion string) (*ModelMetadata, error) {
	url := fmt.Sprintf("%s/api/v1/models/%s/versions/%s/metadata", c.BaseURL, modelName, modelVersion)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call model manager: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to call model manager: invalid status code: %d", resp.StatusCode)
	}

	var modelMetadata ModelMetadata
	if err := json.NewDecoder(resp.Body).Decode(&modelMetadata); err != nil {
		return nil, fmt.Errorf("failed to decode model metadata: %w", err)
	}
	return &modelMetadata, nil
}
