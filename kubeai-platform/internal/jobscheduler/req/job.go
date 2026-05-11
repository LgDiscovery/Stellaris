package req

type CreateJobRequest struct {
	Name      string                 `json:"name"`
	ModelID   *uint                  `json:"model_id,omitempty"`
	Framework string                 `json:"framework"`
	Image     string                 `json:"image"`
	Command   string                 `json:"command"`
	Resources map[string]interface{} `json:"resources"`
	Env       []map[string]string    `json:"env"`
}
