package modelmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"kubeai/internal/model"
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

// CheckModelAvailable 校验模型是否可用，返回模型元数据
func (c *Client) CheckModelAvailable(ctx context.Context, modelName, version string) (*model.ModelVersion, error) {
	url := fmt.Sprintf("%s/api/v1/models/%s/versions/%s", c.BaseURL, modelName, version)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("model not found, status: %d", resp.StatusCode)
	}

	var result struct {
		Code int                `json:"code"`
		Data model.ModelVersion `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 仅active状态的模型可用
	if result.Data.Status != "active" {
		return nil, fmt.Errorf("model version is not active, status: %s", result.Data.Status)
	}
	return &result.Data, nil
}
