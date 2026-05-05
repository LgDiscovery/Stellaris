import apiClient from './axios';

export interface Model {
  id: number;
  name: string;
  description?: string;
  framework?: string;
  task_type?: string;
  owner?: string;
  labels?: string;
  created_at: string;
  updated_at: string;
  versions?: ModelVersion[];
}

export interface ModelVersion {
  id: number;
  model_id: number;
  version: string;
  description?: string;
  storage_path: string;
  framework?: string;
  framework_version?: string;
  metadata?: Record<string, string>;
  metrics?: string;
  parameters?: string;
  training_job_name?: string;
  size: number;
  checksum: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface CreateModelReq {
  name: string;
  description?: string;
  framework?: string;
  task_type?: string;
  labels?: string;
  owner?: string;
}

export interface ModelListResp {
  items: Model[];
  total: number;
  page: number;
  page_size: number;
}

export interface ModelRegisterRequest {
  model_name: string;
  storage_path: string;
  framework: string;
  version: string;
  description?: string;
  metadata?: Record<string, string>;
  training_job_name?: string;
  namespace?: string;
  model_id?: string;
  task_type?: string;
}

export const modelsApi = {
  create: (data: CreateModelReq) =>
    apiClient.post<Model>('/models', data),

  list: (params?: { page?: number; page_size?: number; framework?: string; task_type?: string }) =>
    apiClient.get<ModelListResp>('/models', { params }),

  getByName: (name: string) =>
    apiClient.get<Model>(`/models/${name}`),

  update: (name: string, data: { name: string; description?: string; labels?: string }) =>
    apiClient.put<Model>(`/models/${name}`, data),

  delete: (name: string) =>
    apiClient.delete(`/models/${name}`),

  register: (data: ModelRegisterRequest) =>
    apiClient.post('/models/register', data),

  createVersion: (name: string, data: {
    name: string;
    version: string;
    description?: string;
    framework?: string;
    framework_version?: string;
    metrics?: string;
    parameters?: string;
  }) => apiClient.post<ModelVersion>(`/models/${name}/versions`, data),

  listVersions: (name: string) =>
    apiClient.get<{ items: ModelVersion[] }>(`/models/${name}/versions`),

  getVersion: (name: string, version: string) =>
    apiClient.get<ModelVersion>(`/models/${name}/versions/${version}`),

  getMetadata: (name: string, version: string) =>
    apiClient.get(`/models/${name}/versions/${version}/metadata`),

  downloadVersion: (name: string, version: string) =>
    apiClient.get(`/models/${name}/versions/${version}/download`),

  updateVersionStatus: (data: {
    name: string;
    version: string;
    status: string;
    message?: string;
    instance?: string;
    node?: string;
  }) => apiClient.post('/models/version/status', data),

  deleteVersion: (name: string, version: string) =>
    apiClient.delete(`/models/${name}/versions/${version}`),

  healthCheck: () => apiClient.get('/models/health'),
  readyCheck: () => apiClient.get('/models/ready'),
  metrics: () => apiClient.get('/models/metrics'),
};