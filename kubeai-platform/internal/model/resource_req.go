package model

type ResourceRequest struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
	GPU    string `json:"gpu,omitempty"` // 如需 GPU 检查，我也可以给你实现
}
