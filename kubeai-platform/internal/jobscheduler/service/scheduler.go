package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"kubeai/internal/model"
	"kubeai/internal/pkg/k8s"
	"kubeai/internal/pkg/modelmanager"
	"kubeai/internal/pkg/queue"
	"time"
)

type SchedulerService struct {
	modelClient    *modelmanager.Client
	inferenceQueue *queue.TaskQueue
	trainingQueue  *queue.TaskQueue
	k8sClient      *k8s.K8sClient
}

// SubmitInferenceTask 提交推理任务
func (s *SchedulerService) SubmitInferenceTask(ctx context.Context, req model.InferenceTaskRequest) (*model.InferenceTaskResponse, error) {
	// 1.检验模型可用性
	modelVersion, err := s.modelClient.CheckModelAvailable(ctx, req.ModelName, req.ModelVersion)
	if err != nil {
		return nil, fmt.Errorf("model availability check failed: %v", err)
	}

	// 2. 校验K8s集群算力资源
	available, err := s.k8sClient.CheckResourceAvailable(ctx, req.Resources)
	if err != nil || !available {
		return nil, fmt.Errorf("insufficient compute resources")
	}

	// 3. 构建推理任务，下发到队列
	task := model.InferenceTask{
		TaskID:       fmt.Sprintf("inference-%s", uuid.New().String()[:8]),
		ModelName:    req.ModelName,
		ModelVersion: req.ModelVersion,
		ModelPath:    modelVersion.StoragePath,
		Framework:    modelVersion.Framework,
		Resources:    req.Resources,
		InputData:    req.InputData,
		CreateTime:   time.Now(),
		Status:       "pending",
	}
	if err := s.inferenceQueue.PushTask(ctx, task); err != nil {
		return nil, fmt.Errorf("push task failed: %w", err)
	}

	return &model.InferenceTaskResponse{
		TaskID:  task.TaskID,
		Status:  task.Status,
		Message: "task submitted successfully",
	}, nil
}
