package repository

import (
	"gorm.io/gorm"
	"kubeai/internal/model"
)

type JobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) *JobRepository {
	return &JobRepository{
		db: db,
	}
}

func (r *JobRepository) GetByID(id uint) (*model.TrainingJob, error) {
	var job model.TrainingJob
	err := r.db.Where("id = ?", id).First(&job).Error
	return &job, err
}

func (r *JobRepository) GetByName(name string) (*model.TrainingJob, error) {
	var job model.TrainingJob
	err := r.db.Where("name = ?", name).First(&job).Error
	return &job, err
}

func (r *JobRepository) List(page, size int, status string) ([]model.TrainingJob, int64, error) {
	var jobs []model.TrainingJob
	var total int64
	query := r.db.Model(&model.TrainingJob{})
	if status != "" {
		query = query.Where("status in (?)", status)
	}
	query.Count(&total)
	offset := (page - 1) * size
	err := query.Order("created_at desc").Offset(offset).Limit(size).Find(&jobs).Error
	return jobs, total, err
}

func (r *JobRepository) Create(job *model.TrainingJob) error {
	return r.db.Create(job).Error
}

func (r *JobRepository) Update(job *model.TrainingJob) error {
	return r.db.Save(job).Error
}

func (r *JobRepository) Delete(job *model.TrainingJob) error {
	return r.db.Delete(job).Error
}
