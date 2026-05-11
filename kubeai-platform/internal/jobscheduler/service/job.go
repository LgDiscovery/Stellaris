package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"kubeai/internal/jobscheduler/repository"
	"kubeai/internal/jobscheduler/req"
	"kubeai/internal/model"
	"kubeai/internal/pkg/k8s"
	"time"
)

type JobService struct {
	repo      *repository.JobRepository
	k8sClient *k8s.K8sClient
}

func NewJobService(repo *repository.JobRepository, k8sClient *k8s.K8sClient) *JobService {
	return &JobService{
		repo:      repo,
		k8sClient: k8sClient,
	}
}

func (s *JobService) CreateJob(ctx context.Context, req req.CreateJobRequest) (*model.TrainingJob, error) {
	// 检查名称唯一性
	existing, _ := s.repo.GetByName(req.Name)

	if existing != nil {
		return nil, fmt.Errorf("jobscheduler %s already exists", req.Name)
	}

	// 序列化 Resources 和 Env
	resourcesJSON, err := json.Marshal(req.Resources)
	if err != nil {
		return nil, err
	}
	envJSON, err := json.Marshal(req.Env)
	if err != nil {
		return nil, err
	}

	// 生产K8s Job名称（使用任务名+随机后缀,避免冲突）
	k8sJobName := fmt.Sprintf("%s-%s", req.Name, uuid.New().String()[:8])

	job := &model.TrainingJob{
		Name:      k8sJobName,
		ModelID:   string(*req.ModelID),
		Framework: req.Framework,
		Image:     req.Image,
		Command:   req.Command,
		Resources: datatypes.JSON(resourcesJSON),
		Env:       datatypes.JSON(envJSON),
		Status:    "pending",
		K8sName:   k8sJobName,
	}

	// 保存到数据库
	if err := s.repo.Create(job); err != nil {
		return nil, err
	}

	// 异步创建k8s Job (避免阻塞API响应)
	go func() {
		// 注意： 需要传递新的context,避免原请求取消导致任务终止
		ctx := context.Background()
		_, err := s.k8sClient.CreateJob(ctx, k8sJobName, req.Image, req.Command, req.Resources, req.Env)
		if err != nil {
			// 更新状态为失败
			job.Status = "failed"
			now := time.Now()
			job.FinishedAt = &now
			s.repo.Update(job)
			return
		}

		// 更新状态为running和开始时间
		job.Status = "running"
		now := time.Now()
		job.StartedAt = &now
		s.repo.Update(job)
	}()
	return job, nil
}

func (s *JobService) GetJob(id uint) (*model.TrainingJob, error) {
	return s.repo.GetByID(id)
}

func (s *JobService) ListJobs(page, size int, status string) ([]model.TrainingJob, int64, error) {
	return s.repo.List(page, size, status)
}

func (s *JobService) DeleteJob(ctx context.Context, id uint) error {
	job, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	// 如果job还在K8s中，先删除
	if job.K8sName != "" && (job.Status == "running" || job.Status == "pending") {
		_ = s.k8sClient.DeleteJob(ctx, job.K8sName)
	}
	return s.repo.Delete(job)
}

// SyncJobStatus 同步K8s Job状态到数据库（用于定时任务）
func (s *JobService) SyncJobStatus(ctx context.Context, id uint) error {
	job, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if job.Status == "succeeded" || job.Status == "failed" {
		return nil
	}

	k8sStatus, err := s.k8sClient.GetJobStatus(ctx, job.K8sName)
	if err != nil {
		return err
	}

	if k8sStatus != job.Status {
		job.Status = k8sStatus
		if k8sStatus == "succeeded" || k8sStatus == "failed" {
			now := time.Now()
			job.FinishedAt = &now
		}
		return s.repo.Update(job)
	}
	return nil
}

func (s *JobService) GetJobLog(ctx context.Context, id uint, tailLines int64) (string, error) {
	job, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}
	if job.K8sName == "" {
		return job.K8sName, fmt.Errorf("jobscheduler %s not found", id)
	}
	return s.k8sClient.GetJobLogs(ctx, job.K8sName, tailLines)
}
