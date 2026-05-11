package repository

import (
	"gorm.io/gorm"
	"kubeai/internal/model"
)

type ModelRepository struct {
	db *gorm.DB
}

func NewModelRepository(db *gorm.DB) *ModelRepository {
	return &ModelRepository{db: db}
}

func (r *ModelRepository) Create(m *model.Model) error {
	return r.db.Create(m).Error
}

func (r *ModelRepository) GetByID(id uint) (*model.Model, error) {
	var m model.Model
	err := r.db.Preload("Versions").First(&m, id).First(&m).Error
	return &m, err
}

func (r *ModelRepository) List(page, size int) ([]model.Model, int64, error) {
	var models []model.Model
	var total int64
	offset := (page - 1) * size
	r.db.Model(&model.Model{}).Count(&total)
	err := r.db.Offset(offset).Limit(size).Find(&models).Error
	return models, total, err
}

func (r *ModelRepository) Update(m *model.Model) error {
	return r.db.Save(m).Error
}

func (r *ModelRepository) Delete(id uint) error {
	return r.db.Delete(&model.Model{}, id).Error
}
