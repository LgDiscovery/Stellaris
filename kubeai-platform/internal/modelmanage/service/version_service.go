package service

import (
	"kubeai/internal/model"
	"kubeai/internal/modelmanage/repository"
)

type VersionService struct {
	repo *repository.VersionRepository
}

func NewVersionService(repo *repository.VersionRepository) *VersionService {
	return &VersionService{repo: repo}
}

func (s *VersionService) Create(mv model.ModelVersion) (*model.ModelVersion, error) {
	if err := s.repo.Create(&mv); err != nil {
		return nil, err
	}
	return &mv, nil
}

func (s *VersionService) GetByModelIDAndVersion(modelID uint, version string) (*model.ModelVersion, error) {
	return s.repo.GetByModelIDAndVersion(modelID, version)
}

func (s *VersionService) GetByModelID(modelID uint) ([]model.ModelVersion, int64, error) {
	return s.repo.GetByModelID(modelID)
}

func (s *VersionService) List(page, size int) ([]model.ModelVersion, int64, error) {
	return s.repo.List(page, size)
}

func (s *VersionService) Update(mv *model.ModelVersion) (*model.ModelVersion, error) {
	mv, err := s.repo.GetByModelIDAndVersion(mv.ModelID, mv.ModelVersion)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Update(mv); err != nil {
		return nil, err
	}
	return mv, nil
}

func (s *VersionService) Delete(modelID uint, version string) error {
	mv, err := s.repo.GetByModelIDAndVersion(modelID, version)
	if err != nil {
		return err
	}
	mv.Status = "archived"
	if err := s.repo.Update(mv); err != nil {
		return err
	}
	return nil
}
