import apiClient from './axios';

export interface ResourceRequest {
  cpu?: string;
  memory?: string;
  gpu?: string;
}

export interface InferenceTask {
  task_id: string;
  model_name: string;
  model_version: string;
  resources: ResourceRequest;
  input_data: Record<string, unknown>;
  output_topic?: string;
  status: string;
  priority: number;
  retry_count: number;
  max_retries: number;
  create_time: string;
  update_time: string;
  scheduled_node?: string;
  pod_name?: string;
  error_message?: string;
}

export interface TrainingTask {
  task_id: string;
  name: string;
  model_name?: string;
  framework: string;
  command: string[];
  args: string[];
  resources: ResourceRequest;
  dataset_path?: string;
  output_path?: string;
  status: string;
  priority: number;
  retry_count: number;
  max_retries: number;
  created_at: string;
  updated_at: string;
  pod_name?: string;
  error_message?: string;
}

export interface SubmitInferenceReq {
  request_id: string;
  model_name: string;
  model_version: string;
  resources: ResourceRequest;
  input_data: Record<string, unknown>;
  output_topic?: string;
  priority: number;
  max_retries: number;
}

export interface SubmitTrainingReq {
  request_id: string;
  name: string;
  model_name?: string;
  framework: string;
  image: string;
  command: string[];
  args: string[];
  resources: ResourceRequest;
  distributed?: boolean;
  worker_num?: number;
  master_num?: number;
  env?: { name: string; value: string }[];
  dataset_path?: string;
  output_path?: string;
  priority: number;
  max_retries: number;
  enable_monitor?: boolean;
  enable_logs?: boolean;
}

export interface ListTasksResp {
  items: Array<InferenceTask | TrainingTask>;
  total: number;
  page: number;
  page_size: number;
}

export const jobsApi = {
  submitInference: (data: SubmitInferenceReq) =>
    apiClient.post('/jobs/inference/tasks', data),

  getInferenceTask: (taskId: string) =>
    apiClient.get<InferenceTask>(`/jobs/inference/tasks/${taskId}`),

  submitTraining: (data: SubmitTrainingReq) =>
    apiClient.post('/jobs/training/tasks', data),

  getTrainingTask: (taskId: string) =>
    apiClient.get<TrainingTask>(`/jobs/training/tasks/${taskId}`),

  listTasks: (params: {
    task_type: 'inference' | 'training';
    status?: string;
    page?: number;
    page_size?: number;
  }) => apiClient.get<ListTasksResp>('/jobs/tasks/list', { params }),

  callbackStatus: (data: {
    task_id: string;
    status: string;
    task_type: 'inference' | 'training';
  }) => apiClient.post('/jobs/tasks/callback', data),

  healthCheck: () => apiClient.get('/jobs/health'),
  readyCheck: () => apiClient.get('/jobs/ready'),
  metrics: () => apiClient.get('/jobs/metrics'),
};