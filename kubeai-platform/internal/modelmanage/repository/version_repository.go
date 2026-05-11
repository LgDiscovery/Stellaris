package repository

import (
	"gorm.io/gorm"
	"kubeai/internal/model"
)

type VersionRepository struct {
	db *gorm.DB
}

func NewVersionRepository(db *gorm.DB) *VersionRepository {
	return &VersionRepository{db: db}
}

func (v *VersionRepository) Create(mv *model.ModelVersion) error {
	return v.db.Create(mv).Error
}

func (v *VersionRepository) GetByModelIDAndVersion(modelID uint, version string) (*model.ModelVersion, error) {
	var mv model.ModelVersion
	err := v.db.Where(&model.ModelVersion{ModelID: modelID, ModelVersion: version}).First(&mv).Error
	return &mv, err
}

func (v *VersionRepository) GetByModelID(modelID uint) ([]model.ModelVersion, int64, error) {
	var mvs []model.ModelVersion
	var total int64
	v.db.Model(&model.ModelVersion{}).Where("status=active and modelID = ?", modelID).Count(&total)
	err := v.db.Find(&mvs).Error
	return mvs, total, err
}

func (v *VersionRepository) List(page, size int) ([]model.ModelVersion, int64, error) {
	var mvs []model.ModelVersion
	var total int64
	offset := (page - 1) * size
	v.db.Model(&model.ModelVersion{}).Count(&total)
	err := v.db.Offset(offset).Limit(size).Find(&mvs).Error
	return mvs, total, err
}

func (v *VersionRepository) Update(mv *model.ModelVersion) error {
	return v.db.Save(mv).Error
}

func (v *VersionRepository) Delete(mv *model.ModelVersion) error {
	return v.db.Delete(mv).Error
}
