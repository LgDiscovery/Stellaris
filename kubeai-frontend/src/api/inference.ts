import apiClient from './axios';

export interface InferenceRequest {
  model_name: string;
  model_version?: string;
  input: Record<string, unknown>;
  framework?: string;
}

export interface InferenceResponse {
  task_id: string;
  result: Record<string, unknown>;
}

export interface ControlReq {
  task_id: string;
  task_type: string;
}

export const inferenceApi = {
  execute: (data: InferenceRequest) =>
    apiClient.post<InferenceResponse>('/inference/execute', data),

  cancelTask: (taskId: string, data: ControlReq) =>
    apiClient.post(`/inference/control/tasks/${taskId}/cancel`, data),

  pauseTask: (taskId: string, data: ControlReq) =>
    apiClient.post(`/inference/control/tasks/${taskId}/pause`, data),

  resumeTask: (taskId: string, data: ControlReq) =>
    apiClient.post(`/inference/control/tasks/${taskId}/resume`, data),

  retryTask: (taskId: string, data: ControlReq) =>
    apiClient.post(`/inference/control/tasks/${taskId}/retry`, data),

  healthCheck: () => apiClient.get('/inference/health'),
  readyCheck: () => apiClient.get('/inference/ready'),
  metrics: () => apiClient.get('/inference/metrics'),
};